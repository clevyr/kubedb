package util

import (
	"context"
	"errors"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/namespace_filter"
	"github.com/clevyr/kubedb/internal/consts"
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
	Name             string
	DisableAuthFlags bool
}

func DefaultSetup(cmd *cobra.Command, conf *config.Global, opts SetupOptions) (err error) {
	cmd.SilenceUsage = true

	conf.Kubeconfig = viper.GetString(consts.KubeconfigKey)
	conf.Context, err = cmd.Flags().GetString(consts.ContextFlag)
	if err != nil {
		panic(err)
	}
	conf.Namespace, err = cmd.Flags().GetString(consts.NamespaceFlag)
	if err != nil {
		panic(err)
	}
	conf.Client, err = kubernetes.NewClient(conf.Kubeconfig, conf.Context, conf.Namespace)
	if err != nil {
		return err
	}
	log.WithField("namespace", conf.Client.Namespace).Debug("created kube client")
	conf.Context = conf.Client.Context
	conf.Namespace = conf.Client.Namespace

	access := namespace_filter.NewFromContext(cmd.Context())
	if !access.Match(conf.Client.Namespace) {
		return errors.New("The current action is disabled for namespace " + conf.Client.Namespace)
	}

	podFlag, err := cmd.Flags().GetString(consts.PodFlag)
	if err != nil {
		panic(err)
	}
	var pods []corev1.Pod
	if podFlag != "" {
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

	dialectFlag, err := cmd.Flags().GetString(consts.DialectFlag)
	if err != nil {
		panic(err)
	}
	if dialectFlag == "" {
		// Configure via detection
		if len(pods) == 0 {
			conf.Dialect, pods, err = database.DetectDialect(cmd.Context(), conf.Client)
			if err != nil {
				return err
			}
			log.WithField("dialect", conf.Dialect.Name()).Debug("detected dialect")
		} else {
			conf.Dialect, err = database.DetectDialectFromPod(pods[0])
			if err != nil {
				return err
			}
		}
	} else {
		// Configure via flag
		conf.Dialect, err = database.New(dialectFlag)
		if err != nil {
			return err
		}
		log.WithField("dialect", conf.Dialect.Name()).Debug("configured database")

		if len(pods) == 0 {
			pods, err = conf.Client.GetPodsFiltered(cmd.Context(), conf.Dialect.PodLabels())
			if err != nil {
				return err
			}
		}
	}

	if podFlag == "" {
		pods, err = conf.Dialect.FilterPods(cmd.Context(), conf.Client, pods)
		if err != nil {
			log.WithError(err).Warn("could not query primary instance")
		}
	}

	if len(pods) == 1 {
		conf.DbPod = pods[0]
		conf.JobPod = pods[0]
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
		conf.DbPod = pods[idx]
	}

	conf.Port, err = cmd.Flags().GetUint16(consts.PortFlag)
	if err != nil {
		panic(err)
	}
	if conf.Port == 0 {
		port, err := conf.Client.GetValueFromEnv(cmd.Context(), conf.DbPod, conf.Dialect.PortEnvNames())
		if err != nil {
			log.Debug("could not detect port from pod env")
		} else {
			port, err := strconv.ParseUint(port, 10, 16)
			if err != nil {
				log.WithField("port", port).Debug("failed to parse port from pod env")
			} else {
				conf.Port = uint16(port)
				log.WithField("port", conf.Port).Debug("found port in pod env")
			}
		}
	}
	if conf.Port == 0 {
		conf.Port = conf.Dialect.DefaultPort()
	}

	conf.Database, err = cmd.Flags().GetString(consts.DbnameFlag)
	if err != nil && !opts.DisableAuthFlags {
		panic(err)
	}
	if conf.Database == "" {
		conf.Database, err = conf.Client.GetValueFromEnv(cmd.Context(), conf.DbPod, conf.Dialect.DatabaseEnvNames())
		if err != nil {
			log.Debug("could not detect database from pod env")
		} else {
			log.WithField("database", conf.Database).Debug("found db name in pod env")
		}
	}

	conf.Username, err = cmd.Flags().GetString(consts.UsernameFlag)
	if err != nil && !opts.DisableAuthFlags {
		panic(err)
	}
	if conf.Username == "" {
		conf.Username, err = conf.Client.GetValueFromEnv(cmd.Context(), conf.DbPod, conf.Dialect.UserEnvNames())
		if err != nil {
			conf.Username = conf.Dialect.DefaultUser()
			log.WithField("user", conf.Username).Debug("could not detect user from pod env, using default")
		} else {
			log.WithField("user", conf.Username).Debug("found user in pod env")
		}
	}

	conf.Password, err = cmd.Flags().GetString(consts.PasswordFlag)
	if err != nil && !opts.DisableAuthFlags {
		panic(err)
	}
	if conf.Password == "" {
		conf.Password, err = conf.Client.GetValueFromEnv(cmd.Context(), conf.DbPod, conf.Dialect.PasswordEnvNames(*conf))
		if err != nil {
			return err
		}
	}
	if viper.GetBool(consts.LogRedactKey) {
		log.AddHook(log_hooks.Redact(conf.Password))
	}

	return nil
}

func CreateJob(cmd *cobra.Command, conf *config.Global, opts SetupOptions) error {
	if viper.GetBool(consts.NoJobKey) {
		conf.Host = "127.0.0.1"
		conf.JobPod = conf.DbPod
	} else {
		if err := createJob(cmd, conf, opts.Name); err != nil {
			return err
		}

		if err := watchJobPod(cmd, conf); err != nil {
			Teardown(cmd, conf)
			return err
		}
	}
	return nil
}

func createJob(cmd *cobra.Command, conf *config.Global, actionName string) error {
	image := conf.DbPod.Spec.Containers[0].Image

	name := "kubedb-"
	if actionName != "" {
		name += actionName + "-"
	}

	standardLabels := map[string]string{
		"app.kubernetes.io/name":      "kubedb",
		"app.kubernetes.io/instance":  "kubedb",
		"app.kubernetes.io/component": actionName,
		"app.kubernetes.io/version":   GetVersion(),
	}

	podLabels := map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	maps.Copy(podLabels, standardLabels)
	maps.Copy(podLabels, viper.GetStringMapString(consts.JobPodLabelsKey))

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
			Namespace:    conf.Namespace,
			Labels:       standardLabels,
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds:   ptr.To(int64(24 * time.Hour.Seconds())),
			TTLSecondsAfterFinished: ptr.To(int32(time.Hour.Seconds())),
			BackoffLimit:            ptr.To(int32(0)),
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
											MatchLabels: conf.DbPod.ObjectMeta.Labels,
										},
									},
								},
								{
									Weight: 90,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "topology.kubernetes.io/zone",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.DbPod.ObjectMeta.Labels,
										},
									},
								},
								{
									Weight: 80,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "topology.kubernetes.io/region",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.DbPod.ObjectMeta.Labels,
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

	log.WithField("namespace", conf.Namespace).Info("creating job")

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Minute)
	defer cancel()

	var err error
	if conf.Job, err = conf.Client.Jobs().Create(ctx, &job, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func watchJobPod(cmd *cobra.Command, conf *config.Global) error {
	log.WithFields(log.Fields{
		"namespace": conf.Namespace,
		"name":      "job.batch/" + conf.Job.ObjectMeta.Name,
	}).Info("waiting for job...")

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()

	watch, err := conf.Client.Pods().Watch(ctx, metav1.ListOptions{
		LabelSelector: jobPodLabelSelector(conf, conf.Job),
	})
	if err != nil {
		return pollJobPod(ctx, conf)
	}
	defer func() {
		watch.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watch.ResultChan():
			if pod, ok := event.Object.(*corev1.Pod); ok {
				switch pod.Status.Phase {
				case corev1.PodRunning:
					conf.Host = conf.DbPod.Status.PodIP
					pod.DeepCopyInto(&conf.JobPod)
					return nil
				case corev1.PodFailed:
					return errors.New("pod failed")
				case corev1.PodSucceeded:
					return errors.New("pod exited early")
				}
			} else {
				return errors.New("unexpected runtime object type")
			}
		}
	}
}

func pollJobPod(ctx context.Context, conf *config.Global) error {
	return wait.PollUntilContextCancel(
		ctx, time.Second, true, func(ctx context.Context) (done bool, err error) {
			list, err := conf.Client.Pods().List(ctx, metav1.ListOptions{
				LabelSelector: jobPodLabelSelector(conf, conf.Job),
			})
			if err != nil {
				return false, err
			}

			if len(list.Items) == 0 {
				return false, nil
			}

			switch list.Items[0].Status.Phase {
			case corev1.PodRunning:
				conf.Host = conf.DbPod.Status.PodIP
				conf.JobPod = list.Items[0]
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

func jobPodLabelSelector(conf *config.Global, job *batchv1.Job) string {
	useNewLabel, err := conf.Client.MinServerVersion(1, 27)
	if err != nil {
		log.WithError(err).Warn("failed to query server version; assuming v1.27+")
		useNewLabel = true
	}

	if useNewLabel {
		return "batch.kubernetes.io/controller-uid=" + string(job.ObjectMeta.UID)
	}
	return "controller-uid=" + string(job.ObjectMeta.UID)
}
