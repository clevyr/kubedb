//nolint:gocyclo
package util

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"gabe565.com/utils/must"
	"github.com/charmbracelet/huh"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/finalizer"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log/mask"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/spf13/cobra"
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
	Name     string
	NoSurvey bool
}

func DefaultSetup(cmd *cobra.Command, conf *conftypes.Global, opts SetupOptions) error {
	if opts.Name == "" {
		opts.Name = cmd.Name()
	}

	cmd.SilenceUsage = true
	ctx := cmd.Context()

	var err error
	conf.Client, err = kubernetes.NewClient(conf.Kubeconfig, conf.Context, conf.Namespace)
	if err != nil {
		return err
	}

	slog.Debug("Created kube client", "namespace", conf.Client.Namespace)
	conf.Context = conf.Client.Context
	conf.Namespace = conf.Client.Namespace

	var pods []corev1.Pod
	if conf.PodName != "" {
		slashIdx := strings.IndexRune(conf.PodName, '/')
		if slashIdx != 0 && slashIdx+1 < len(conf.PodName) {
			conf.PodName = conf.PodName[slashIdx+1:]
		}
		pod, err := conf.Client.Pods().Get(ctx, conf.PodName, metav1.GetOptions{})
		if err != nil {
			checkNamespaceExists(ctx, conf)
			return err
		}
		pods = []corev1.Pod{*pod}
	}

	switch {
	case conf.DialectName != "":
		conf.Dialect, err = database.New(conf.DialectName)
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
	case len(pods) == 0:
		result, err := database.DetectDialect(ctx, conf.Client)
		if err != nil {
			checkNamespaceExists(ctx, conf)
			return err
		}
		if len(result) == 1 || opts.NoSurvey {
			for _, v := range result {
				conf.Dialect = v.Dialect
				pods = v.Pods
				break
			}
		} else {
			slices.SortFunc(result, func(a, b database.DetectResult) int {
				return cmp.Compare(a.Dialect.PrettyName(), b.Dialect.PrettyName())
			})
			opts := make([]huh.Option[int], 0, len(result))
			for i, v := range result {
				opts = append(opts, huh.NewOption(v.Dialect.PrettyName(), i))
			}
			var chosen int
			if err := tui.NewForm(huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select database type").
					Options(opts...).
					Value(&chosen),
			)).Run(); err != nil {
				return err
			}
			conf.Dialect = result[chosen].Dialect
			pods = result[chosen].Pods
		}
		slog.Debug("Detected dialect", "dialect", conf.Dialect.Name())
	default:
		conf.Dialect, err = database.DetectDialectFromPod(pods[0])
		if err != nil {
			checkNamespaceExists(ctx, conf)
			return err
		}
	}

	if len(pods) > 1 {
		if db, ok := conf.Dialect.(conftypes.DBFilterer); ok && conf.PodName == "" {
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

	// Detect port
	if db, ok := conf.Dialect.(conftypes.DBHasPort); ok && conf.Port == 0 {
		port, err := db.PortEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			slog.Debug("Could not detect port")
		} else {
			port, err := strconv.ParseUint(port, 10, 16)
			if err != nil {
				slog.Debug("Failed to parse port", "error", err)
			} else {
				conf.Port = uint16(port)
				slog.Debug("Found port", "port", conf.Port)
			}
		}

		if conf.Port == 0 {
			conf.Port = db.PortDefault()
		}
	}

	// Detect database
	if db, ok := conf.Dialect.(conftypes.DBHasDatabase); ok && conf.Database == "" {
		conf.Database, err = db.DatabaseEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			slog.Debug("Could not detect db name", "error", err)
		} else {
			slog.Debug("Found db name", "database", conf.Database)
		}
	}

	// Detect username
	if db, ok := conf.Dialect.(conftypes.DBHasUser); ok && conf.Username == "" {
		conf.Username, err = db.UserEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			conf.Username = db.UserDefault()
			slog.Debug("Could not detect user, using default", "error", err, "user", conf.Username)
		} else {
			slog.Debug("Found user", "user", conf.Username)
		}
	}

	// Detect password
	if db, ok := conf.Dialect.(conftypes.DBHasPassword); ok && conf.Password == "" {
		conf.Password, err = db.PasswordEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			slog.Warn("Could not detect password", "error", err)
		} else {
			slog.Debug("Found password")
		}
	}

	if conf.Password != "" && conf.Log.Mask {
		mask.Add(conf.Password)
	}

	if db, ok := conf.Dialect.(conftypes.DBCanDisableJob); ok && db.DisableJob() {
		must.Must(config.K.Set(consts.FlagCreateJob, false))
	}
	if !conf.CreateJob {
		conf.Host = "127.0.0.1"
		conf.JobPod = conf.DBPod
	}

	return nil
}

func CreateJob(ctx context.Context, cmd *cobra.Command, conf *conftypes.Global, opts SetupOptions) error {
	if opts.Name == "" {
		opts.Name = cmd.Name()
	}

	if conf.CreateJob {
		if err := createJob(ctx, conf, opts.Name); err != nil {
			return err
		}
		finalizer.Add(func(_ error) {
			Teardown(conf)
		})

		if err := watchJobPod(ctx, conf); err != nil {
			return err
		}
	}
	return nil
}

func createJob(ctx context.Context, conf *conftypes.Global, actionName string) error {
	defaultContainer := conf.DBPod.Spec.Containers[0]
	if name := conf.DBPod.Annotations[podcmd.DefaultContainerAnnotationName]; name != "" {
		for _, container := range conf.DBPod.Spec.Containers {
			if container.Name == name {
				defaultContainer = container
				break
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
	maps.Copy(podLabels, conf.JobPodLabels)

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
							Image:           defaultContainer.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"sleep", "infinity"},
							SecurityContext: defaultContainer.SecurityContext,
						},
					},
					SecurityContext: conf.DBPod.Spec.SecurityContext,
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

	if conf.CreateNetworkPolicy {
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
			nsLog.Warn("Failed to create network policy", "error", err)
			conf.CreateNetworkPolicy = false
		}
	}

	return nil
}

var (
	ErrJobPodFailed    = errors.New("job pod failed")
	ErrJobPodEarlyExit = errors.New("job pod exited early")
	ErrJobPodInvalid   = errors.New("unexpected job pod object type")
)

func watchJobPod(ctx context.Context, conf *conftypes.Global) error {
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

func pollJobPod(ctx context.Context, conf *conftypes.Global) error {
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

func jobPodUIDLabel(conf *conftypes.Global, job *batchv1.Job) (string, string) {
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

func jobPodLabelSelector(conf *conftypes.Global, job *batchv1.Job) string {
	k, v := jobPodUIDLabel(conf, job)
	return k + "=" + v
}

func jobPodNameLabel(conf *conftypes.Global, job *batchv1.Job) (string, string) {
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

func checkNamespaceExists(ctx context.Context, conf *conftypes.Global) {
	if _, err := conf.Client.Namespaces().Get(ctx, conf.Namespace, metav1.GetOptions{}); err != nil {
		slog.Warn("Namespace may not exist", "error", err)
	}
}
