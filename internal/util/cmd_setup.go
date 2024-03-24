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
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log_hooks"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

type SetupOptions struct {
	Name             string
	DisableAuthFlags bool
	NoSurvey         bool
}

func DefaultSetup(cmd *cobra.Command, conf *config.Global, opts SetupOptions) (err error) {
	cmd.SilenceUsage = true

	ctx := cmd.Context()

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

	if _, err := conf.Client.Namespaces().Get(ctx, conf.Namespace, metav1.GetOptions{}); err != nil {
		log.WithError(err).Warn("namespace may not exist")
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
		pod, err := conf.Client.Pods().Get(ctx, podFlag, metav1.GetOptions{})
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
			conf.Dialect, pods, err = database.DetectDialect(ctx, conf.Client)
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
			pods, err = conf.Client.GetPodsFiltered(ctx, conf.Dialect.PodFilters())
			if err != nil {
				return err
			}

			if len(pods) == 0 {
				return kubernetes.ErrPodNotFound
			}
		}
	}

	if len(pods) > 1 {
		if db, ok := conf.Dialect.(config.DatabaseFilter); ok && podFlag == "" {
			filtered, err := db.FilterPods(ctx, conf.Client, pods)
			if err != nil {
				log.WithError(err).Warn("could not query primary instance")
			}

			if len(filtered) != 0 {
				pods = filtered
			}
		}
	}

	if len(pods) == 1 || opts.NoSurvey {
		conf.DbPod = pods[0]
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

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		conf.Port, err = cmd.Flags().GetUint16(consts.PortFlag)
		if err != nil {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DatabasePort); ok && conf.Port == 0 {
			port, err := db.PortEnvNames().Search(ctx, conf.Client, conf.DbPod)
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

			if conf.Port == 0 {
				conf.Port = db.DefaultPort()
			}
		}
		return nil
	})

	group.Go(func() error {
		conf.Database, err = cmd.Flags().GetString(consts.DbnameFlag)
		if err != nil && !opts.DisableAuthFlags {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DatabaseDb); ok && conf.Database == "" {
			conf.Database, err = db.DatabaseEnvNames().Search(ctx, conf.Client, conf.DbPod)
			if err != nil {
				log.Debug("could not detect database from pod env")
			} else {
				log.WithField("database", conf.Database).Debug("found db name in pod env")
			}
		}
		return nil
	})

	group.Go(func() error {
		conf.Username, err = cmd.Flags().GetString(consts.UsernameFlag)
		if err != nil && !opts.DisableAuthFlags {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DatabaseUsername); ok && conf.Username == "" {
			conf.Username, err = db.UserEnvNames().Search(ctx, conf.Client, conf.DbPod)
			if err != nil {
				conf.Username = db.DefaultUser()
				log.WithField("user", conf.Username).Debug("could not detect user from pod env, using default")
			} else {
				log.WithField("user", conf.Username).Debug("found user in pod env")
			}
		}

		conf.Password, err = cmd.Flags().GetString(consts.PasswordFlag)
		if err != nil && !opts.DisableAuthFlags {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DatabasePassword); ok && conf.Password == "" {
			conf.Password, err = db.PasswordEnvNames(*conf).Search(ctx, conf.Client, conf.DbPod)
			if err != nil {
				return err
			}
		}

		if viper.GetBool(consts.LogRedactKey) {
			log.AddHook(log_hooks.Redact(conf.Password))
		}

		return nil
	})

	group.Go(func() error {
		if db, ok := conf.Dialect.(config.DatabaseDisableJob); ok && db.DisableJob() {
			viper.Set(consts.CreateJobKey, false)
		}
		if !viper.GetBool(consts.CreateJobKey) {
			conf.Host = "127.0.0.1"
			conf.JobPod = conf.DbPod
		}
		return nil
	})

	return group.Wait()
}

func CreateJob(ctx context.Context, conf *config.Global, opts SetupOptions) error {
	if viper.GetBool(consts.CreateJobKey) {
		if err := createJob(ctx, conf, opts.Name); err != nil {
			return err
		}
		cobra.OnFinalize(func() {
			Teardown(conf)
		})

		if err := watchJobPod(ctx, conf); err != nil {
			return err
		}
	}
	return nil
}

func createJob(ctx context.Context, conf *config.Global, actionName string) (err error) {
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

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	log.WithField("namespace", conf.Namespace).Info("creating job")
	if conf.Job, err = conf.Client.Jobs().Create(ctx, &job, metav1.CreateOptions{}); err != nil {
		return err
	}

	if viper.GetBool(consts.CreateNetworkPolicyKey) {
		jobPodKey, jobPodVal := jobPodLabel(conf, conf.Job)
		policy := networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      conf.Job.Name,
				Namespace: conf.Client.Namespace,
				Labels:    standardLabels,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{
					jobPodKey: jobPodVal,
				}},
				PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
				Egress: []networkingv1.NetworkPolicyEgressRule{
					{
						To: []networkingv1.NetworkPolicyPeer{{
							PodSelector: ptr.To(metav1.LabelSelector{MatchLabels: conf.DbPod.Labels}),
						}},
						Ports: []networkingv1.NetworkPolicyPort{{
							Port: ptr.To(intstr.FromInt32(int32(conf.Port))),
						}},
					},
				},
			},
		}

		log.WithField("namespace", conf.Namespace).Info("creating network policy")
		if _, err := conf.Client.NetworkPolicies().Create(ctx, &policy, metav1.CreateOptions{}); err != nil {
			log.WithError(err).Error("failed to create network policy")
			viper.Set(consts.CreateNetworkPolicyKey, "false")
		}
	}

	return nil
}

func watchJobPod(ctx context.Context, conf *config.Global) error {
	log.WithFields(log.Fields{
		"namespace": conf.Namespace,
		"name":      "job.batch/" + conf.Job.ObjectMeta.Name,
	}).Info("waiting for job...")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
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

func jobPodLabel(conf *config.Global, job *batchv1.Job) (string, string) {
	useNewLabel, err := conf.Client.MinServerVersion(1, 27)
	if err != nil {
		log.WithError(err).Warn("failed to query server version; assuming v1.27+")
		useNewLabel = true
	}

	var key string
	if useNewLabel {
		key = "batch.kubernetes.io/controller-uid"
	} else {
		key = "controller-uid"
	}
	return key, string(job.ObjectMeta.UID)
}

func jobPodLabelSelector(conf *config.Global, job *batchv1.Job) string {
	k, v := jobPodLabel(conf, job)
	return k + "=" + v
}
