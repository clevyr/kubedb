package util

import (
	"context"
	"errors"
	"maps"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log_hooks"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

type SetupOptions struct {
	Name       string
	DisableJob bool
}

func DefaultSetup(cmd *cobra.Command, conf *config.Global, opts SetupOptions) (err error) {
	cmd.SilenceUsage = true

	conf.Client, err = kubernetes.NewClient(conf.Kubeconfig, conf.Context, conf.Namespace)
	if err != nil {
		return err
	}
	log.WithField("namespace", conf.Client.Namespace).Debug("created kube client")
	conf.Namespace = conf.Client.Namespace

	access := config.NewNamespaceRegexp(cmd.Annotations["access"])
	if !access.Match(conf.Client.Namespace) {
		return errors.New("The current action is disabled for namespace " + conf.Client.Namespace)
	}

	dialectFlag, err := cmd.Flags().GetString("dialect")
	if err != nil {
		panic(err)
	}
	if dialectFlag == "" {
		dialectFlag, err = cmd.Flags().GetString("grammar")
		if err != nil {
			panic(err)
		}
	}

	var pods []corev1.Pod
	if dialectFlag == "" {
		// Configure via detection
		conf.Dialect, pods, err = database.DetectDialect(cmd.Context(), conf.Client)
		if err != nil {
			return err
		}
		log.WithField("dialect", conf.Dialect.Name()).Debug("detected dialect")
	} else {
		// Configure via flag
		conf.Dialect, err = database.New(dialectFlag)
		if err != nil {
			return err
		}
		log.WithField("dialect", conf.Dialect.Name()).Debug("configured database")

		podFlag, err := cmd.Flags().GetString("pod")
		if err != nil {
			panic(err)
		}

		if podFlag == "" {
			pods, err = conf.Client.GetPodsFiltered(cmd.Context(), conf.Dialect.PodLabels())
			if err != nil {
				return err
			}
		} else {
			slashIdx := strings.IndexRune(podFlag, '/')
			if slashIdx != 0 && slashIdx+1 < len(podFlag) {
				podFlag = podFlag[slashIdx+1:]
			}
			pod, err := conf.Client.Pods().Get(cmd.Context(), podFlag, metav1.GetOptions{})
			if err != nil {
				return err
			}
			pods = []corev1.Pod{*pod}
		}
	}

	pods, err = conf.Dialect.FilterPods(cmd.Context(), conf.Client, pods)
	if err != nil {
		log.WithError(err).Warn("could not query primary instance")
	}

	if len(pods) == 1 {
		conf.Pod = pods[0]
	} else {
		names := make([]string, 0, len(pods))
		for _, pod := range pods {
			names = append(names, pod.Name)
		}
		var idx int
		err = survey.AskOne(&survey.Select{
			Message: "Found multiple database pods. Select the desired instance.",
			Options: names,
		}, &idx)
		if err != nil {
			return err
		}
		conf.Pod = pods[idx]
	}

	conf.Database, err = cmd.Flags().GetString("dbname")
	if err != nil {
		panic(err)
	}
	if conf.Database == "" {
		conf.Database, err = conf.Client.GetValueFromEnv(cmd.Context(), conf.Pod, conf.Dialect.DatabaseEnvNames())
		if err != nil {
			log.Debug("could not detect database from pod env")
		} else {
			log.WithField("database", conf.Database).Debug("found db name in pod env")
		}
	}

	conf.Username, err = cmd.Flags().GetString("username")
	if err != nil {
		panic(err)
	}
	if conf.Username == "" {
		conf.Username, err = conf.Client.GetValueFromEnv(cmd.Context(), conf.Pod, conf.Dialect.UserEnvNames())
		if err != nil {
			conf.Username = conf.Dialect.DefaultUser()
			log.WithField("user", conf.Username).Debug("could not detect user from pod env, using default")
		} else {
			log.WithField("user", conf.Username).Debug("found user in pod env")
		}
	}

	conf.Password, err = cmd.Flags().GetString("password")
	if err != nil {
		panic(err)
	}
	if conf.Password == "" {
		conf.Password, err = conf.Client.GetValueFromEnv(cmd.Context(), conf.Pod, conf.Dialect.PasswordEnvNames(*conf))
		if err != nil {
			return err
		}
	}
	if viper.GetBool("redact") {
		log.AddHook(log_hooks.Redact(conf.Password))
	}

	if opts.DisableJob || viper.GetBool("kubernetes.no-job") {
		conf.Host = "127.0.0.1"
	} else {
		if err := createJob(cmd, conf, opts.Name); err != nil {
			return err
		}

		if err := waitForPod(cmd, conf); err != nil {
			Teardown(cmd, conf)
			return err
		}
	}

	return nil
}

func createJob(cmd *cobra.Command, conf *config.Global, actionName string) error {
	image := conf.Pod.Spec.Containers[0].Image

	name := "kubedb-"
	if actionName != "" {
		name += actionName + "-"
	}

	standardLabels := map[string]string{
		"app.kubernetes.io/name":      "kubedb",
		"app.kubernetes.io/instance":  "kubedb",
		"app.kubernetes.io/component": actionName,
	}

	podLabels := map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	maps.Copy(podLabels, standardLabels)
	maps.Copy(podLabels, viper.GetStringMapString("kubernetes.job-pod-labels"))

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
			Namespace:    conf.Namespace,
			Labels:       standardLabels,
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: ptr.To(int64(24 * time.Hour.Seconds())),
			BackoffLimit:          ptr.To(int32(0)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"linkerd.io/inject": "disabled",
					},
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 corev1.RestartPolicyNever,
					TerminationGracePeriodSeconds: ptr.To(int64(0)),
					Affinity: &corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "kubernetes.io/hostname",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.Pod.ObjectMeta.Labels,
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "kubedb",
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"sleep"},
							Args:            []string{"infinity"},
							TTY:             true,
							Stdin:           true,
						},
					},
				},
			},
		},
	}

	log.Debug("creating job")

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Minute)
	defer cancel()

	var err error
	if conf.Job, err = conf.Client.ClientSet.BatchV1().Jobs(conf.Namespace).Create(
		ctx, &job, metav1.CreateOptions{},
	); err != nil {
		return err
	}

	return nil
}

func waitForPod(cmd *cobra.Command, conf *config.Global) error {
	log.WithField("name", conf.Job.ObjectMeta.Name).Info("waiting for job...")

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()

	return wait.PollUntilContextCancel(
		ctx, time.Second, true, func(ctx context.Context) (done bool, err error) {
			list, err := conf.Client.ClientSet.CoreV1().Pods(conf.Namespace).List(
				cmd.Context(), metav1.ListOptions{
					LabelSelector: "batch.kubernetes.io/controller-uid=" + string(conf.Job.ObjectMeta.UID),
				},
			)
			if err != nil {
				return false, err
			}

			if len(list.Items) == 0 {
				return false, nil
			}

			switch list.Items[0].Status.Phase {
			case corev1.PodRunning:
				conf.Host = conf.Pod.Status.PodIP
				conf.Pod = list.Items[0]
				return true, nil
			case corev1.PodFailed:
				return false, errors.New("pod failed")
			case corev1.PodSucceeded:
				return false, errors.New("pod exited early")
			default:
				return false, nil
			}
		},
	)
}
