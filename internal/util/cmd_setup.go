package util

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"time"

	"charm.land/huh/v2"
	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/discovery"
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
)

func DefaultSetup(cmd *cobra.Command, conf *conftypes.Global) error {
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

	results, err := discovery.Discover(ctx, conf.Client, conf.PodName, conf.DialectName)
	if err != nil {
		checkNamespaceExists(ctx, conf)
		return err
	}

	var pods []corev1.Pod
	if len(results) == 1 || config.IsCompletion {
		conf.Dialect = results[0].Dialect
		pods = results[0].Pods
	} else {
		slices.SortFunc(results, func(a, b discovery.Result) int {
			return cmp.Compare(a.Dialect.PrettyName(), b.Dialect.PrettyName())
		})
		opts := make([]huh.Option[int], 0, len(results))
		for i, v := range results {
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
		conf.Dialect = results[chosen].Dialect
		pods = results[chosen].Pods
	}
	slog.Debug("Detected dialect", "dialect", conf.Dialect.Name())

	if len(pods) == 1 || config.IsCompletion {
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

func CreateJob(ctx context.Context, cmd *cobra.Command, conf *conftypes.Global) error {
	if conf.CreateJob {
		if err := createJob(ctx, conf, cmd.Name()); err != nil {
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

	const appName = "kubedb"

	name := appName + "-"
	if actionName != "" {
		name += actionName + "-"
	}

	standardLabels := map[string]string{
		"app.kubernetes.io/name":      appName,
		"app.kubernetes.io/instance":  appName,
		"app.kubernetes.io/component": actionName,
		"app.kubernetes.io/version":   GetVersion(),
	}

	podLabels := map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	if instance, ok := conf.DBPod.Labels["app.kubernetes.io/instance"]; ok {
		podLabels[instance+"-client"] = "true"
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
			ActiveDeadlineSeconds:   new(int64(24 * time.Hour.Seconds())),
			TTLSecondsAfterFinished: new(int32(time.Hour.Seconds())),
			BackoffLimit:            new(int32(0)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"linkerd.io/inject": "disabled",
					},
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 corev1.RestartPolicyNever,
					TerminationGracePeriodSeconds: new(int64(0)),
					Affinity: &corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "kubernetes.io/hostname",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.DBPod.Labels,
										},
									},
								},
								{
									Weight: 90,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "topology.kubernetes.io/zone",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.DBPod.Labels,
										},
									},
								},
								{
									Weight: 80,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "topology.kubernetes.io/region",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: conf.DBPod.Labels,
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
							NamespaceSelector: new(metav1.LabelSelector{MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": conf.Client.Namespace,
							}}),
						}},
						Ports: []networkingv1.NetworkPolicyPort{{
							Port: new(intstr.FromInt32(int32(conf.Port))),
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
		"job", conf.Job.Name,
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
	return key, string(job.UID)
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
