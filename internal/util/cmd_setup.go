//nolint:gocyclo
package util

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubectl/pkg/cmd/util/podcmd"
	"k8s.io/utils/ptr"
)

type SetupOptions struct {
	Name             string
	DisableAuthFlags bool
	NoSurvey         bool
}

func DefaultSetup(cmd *cobra.Command, conf *config.Global, opts SetupOptions) error {
	cmd.SilenceUsage = true

	ctx := cmd.Context()

	conf.Kubeconfig = viper.GetString(consts.KubeconfigKey)
	var err error
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
	slog.Debug("Created kube client", "namespace", conf.Client.Namespace)
	conf.Context = conf.Client.Context
	conf.Namespace = conf.Client.Namespace

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
			checkNamespaceExists(ctx, conf)
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
			result, err := database.DetectDialect(ctx, conf.Client)
			if err != nil {
				checkNamespaceExists(ctx, conf)
				return err
			}
			if len(result) == 1 || opts.NoSurvey {
				for dialect, p := range result {
					conf.Dialect = dialect
					pods = p
					break
				}
			} else {
				opts := make([]huh.Option[config.Database], 0, len(result))
				for dialect := range result {
					opts = append(opts, huh.NewOption(dialect.PrettyName(), dialect))
				}
				var chosen config.Database
				if err := tui.NewForm(huh.NewGroup(
					huh.NewSelect[config.Database]().
						Title("Select database type").
						Options(opts...).
						Value(&chosen),
				)).Run(); err != nil {
					return err
				}
				conf.Dialect = chosen
				pods = result[chosen]
			}
			slog.Debug("Detected dialect", "dialect", conf.Dialect.Name())
		} else {
			conf.Dialect, err = database.DetectDialectFromPod(pods[0])
			if err != nil {
				checkNamespaceExists(ctx, conf)
				return err
			}
		}
	} else {
		// Configure via flag
		conf.Dialect, err = database.New(dialectFlag)
		if err != nil {
			return err
		}
		slog.Debug("Configured database", "dialect", conf.Dialect.Name())

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
		if db, ok := conf.Dialect.(config.DBFilterer); ok && podFlag == "" {
			filtered, err := db.FilterPods(ctx, conf.Client, pods)
			if err != nil {
				slog.Warn("Could not query primary instance", "error", err)
			}

			if len(filtered) != 0 {
				pods = filtered
			}
		}
	}

	if len(pods) == 1 || opts.NoSurvey {
		conf.DBPod = pods[0]
	} else {
		opts := make([]huh.Option[int], 0, len(pods))
		for i, pod := range pods {
			opts = append(opts, huh.NewOption(pod.Name, i))
		}
		var idx int
		if err := tui.NewForm(huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select " + conf.Dialect.PrettyName() + " instance").
				Options(opts...).
				Value(&idx),
		)).Run(); err != nil {
			return err
		}
		conf.DBPod = pods[idx]
	}

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		conf.Port, err = cmd.Flags().GetUint16(consts.PortFlag)
		if err != nil {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DBHasPort); ok && conf.Port == 0 {
			port, err := db.PortEnvs().Search(ctx, conf.Client, conf.DBPod)
			if err != nil {
				slog.Debug("Could not detect port from pod env")
			} else {
				port, err := strconv.ParseUint(port, 10, 16)
				if err != nil {
					slog.Debug("Failed to parse port from pod env", "error", err)
				} else {
					conf.Port = uint16(port)
					slog.Debug("Found port in pod env", "port", conf.Port)
				}
			}

			if conf.Port == 0 {
				conf.Port = db.PortDefault()
			}
		}
		return nil
	})

	group.Go(func() error {
		conf.Database, err = cmd.Flags().GetString(consts.DbnameFlag)
		if err != nil && !opts.DisableAuthFlags {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DBHasDatabase); ok && conf.Database == "" {
			conf.Database, err = db.DatabaseEnvs().Search(ctx, conf.Client, conf.DBPod)
			if err != nil {
				slog.Debug("Could not detect database from pod env", "error", err)
			} else {
				slog.Debug("Found db name in pod env", "database", conf.Database)
			}
		}
		return nil
	})

	group.Go(func() error {
		conf.Username, err = cmd.Flags().GetString(consts.UsernameFlag)
		if err != nil && !opts.DisableAuthFlags {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DBHasUser); ok && conf.Username == "" {
			conf.Username, err = db.UserEnvs().Search(ctx, conf.Client, conf.DBPod)
			if err != nil {
				conf.Username = db.UserDefault()
				slog.Debug("Could not detect user from pod env, using default", "error", err, "user", conf.Username)
			} else {
				slog.Debug("Found user in pod env", "user", conf.Username)
			}
		}

		conf.Password, err = cmd.Flags().GetString(consts.PasswordFlag)
		if err != nil && !opts.DisableAuthFlags {
			panic(err)
		}

		if db, ok := conf.Dialect.(config.DBHasPassword); ok && conf.Password == "" {
			conf.Password, err = db.PasswordEnvs(*conf).Search(ctx, conf.Client, conf.DBPod)
			if err != nil {
				slog.Error("Could not detect password from pod env", "error", err)
				return err
			}
			slog.Debug("Found password in pod env")
		}

		if conf.Password != "" && viper.GetBool(consts.LogMaskKey) {
			log.AddMask(conf.Password)
		}
		return nil
	})

	group.Go(func() error {
		if db, ok := conf.Dialect.(config.DBCanDisableJob); ok && db.DisableJob() {
			viper.Set(consts.CreateJobKey, false)
		}
		if !viper.GetBool(consts.CreateJobKey) {
			conf.Host = "127.0.0.1"
			conf.JobPod = conf.DBPod
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
		OnFinalize(func(_ error) {
			Teardown(conf)
		})

		if err := watchJobPod(ctx, conf); err != nil {
			return err
		}
	}
	return nil
}

func createJob(ctx context.Context, conf *config.Global, actionName string) error {
	image := conf.DBPod.Spec.Containers[0].Image
	if name := conf.DBPod.Annotations[podcmd.DefaultContainerAnnotationName]; name != "" {
		for _, container := range conf.DBPod.Spec.Containers {
			if container.Name == name {
				image = container.Image
			}
		}
	}

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
											MatchLabels: conf.DBPod.ObjectMeta.Labels,
										},
									},
								},
								{
									Weight: 90,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "topology.kubernetes.io/zone",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.DBPod.ObjectMeta.Labels,
										},
									},
								},
								{
									Weight: 80,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "topology.kubernetes.io/region",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.DBPod.ObjectMeta.Labels,
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

	nsLog := slog.With("namespace", conf.Namespace)
	nsLog.Info("Creating job")
	var err error
	if conf.Job, err = conf.Client.Jobs().Create(ctx, &job, metav1.CreateOptions{}); err != nil {
		return err
	}

	if viper.GetBool(consts.CreateNetworkPolicyKey) {
		jobPodKey, jobPodVal := jobPodNameLabel(conf, conf.Job)
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
				PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
				Ingress:     []networkingv1.NetworkPolicyIngressRule{{}},
				Egress: []networkingv1.NetworkPolicyEgressRule{
					{
						To: []networkingv1.NetworkPolicyPeer{{
							NamespaceSelector: ptr.To(metav1.LabelSelector{MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": conf.Client.Namespace,
							}}),
						}},
						Ports: []networkingv1.NetworkPolicyPort{{
							Port: ptr.To(intstr.FromInt32(int32(conf.Port))),
						}},
					},
				},
			},
		}

		nsLog.Debug("Creating network policy")
		if _, err := conf.Client.NetworkPolicies().Create(ctx, &policy, metav1.CreateOptions{}); err != nil {
			nsLog.Error("Failed to create network policy", "error", err)
			viper.Set(consts.CreateNetworkPolicyKey, "false")
		}
	}

	return nil
}

var (
	ErrJobPodFailed    = errors.New("job pod failed")
	ErrJobPodEarlyExit = errors.New("job pod exited early")
	ErrJobPodInvalid   = errors.New("unexpected job pod object type")
)

func watchJobPod(ctx context.Context, conf *config.Global) error {
	slog.Info("Waiting for job...",
		"namespace", conf.Namespace,
		"job", conf.Job.ObjectMeta.Name,
	)

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
					conf.Host = conf.DBPod.Status.PodIP
					pod.DeepCopyInto(&conf.JobPod)
					return nil
				case corev1.PodFailed:
					return ErrJobPodFailed
				case corev1.PodSucceeded:
					return ErrJobPodEarlyExit
				}
			} else {
				return ErrJobPodInvalid
			}
		}
	}
}

func pollJobPod(ctx context.Context, conf *config.Global) error {
	return wait.PollUntilContextCancel(
		ctx, time.Second, true, func(ctx context.Context) (bool, error) {
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
				conf.Host = conf.DBPod.Status.PodIP
				conf.JobPod = list.Items[0]
				return true, nil
			case corev1.PodFailed:
				return false, ErrJobPodFailed
			case corev1.PodSucceeded:
				return false, ErrJobPodEarlyExit
			default:
				return false, nil
			}
		},
	)
}

func jobPodUIDLabel(conf *config.Global, job *batchv1.Job) (string, string) {
	useNewLabel, err := conf.Client.MinServerVersion(1, 27)
	if err != nil {
		slog.Warn("Failed to query server version; assuming v1.27+", "error", err)
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
	k, v := jobPodUIDLabel(conf, job)
	return k + "=" + v
}

func jobPodNameLabel(conf *config.Global, job *batchv1.Job) (string, string) {
	useNewLabel, err := conf.Client.MinServerVersion(1, 27)
	if err != nil {
		slog.Warn("Failed to query server version; assuming v1.27+", "error", err)
		useNewLabel = true
	}

	var key string
	if useNewLabel {
		key = "batch.kubernetes.io/job-name"
	} else {
		key = "job-name"
	}
	return key, job.Name
}

func checkNamespaceExists(ctx context.Context, conf *config.Global) {
	if _, err := conf.Client.Namespaces().Get(ctx, conf.Namespace, metav1.GetOptions{}); err != nil {
		slog.Warn("Namespace may not exist", "error", err)
	}
}
