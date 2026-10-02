package util

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"time"

	"al.essio.dev/pkg/shellescape"
	"charm.land/huh/v2"
	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/command"
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
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
		found, err := db.PortEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			slog.Debug("Could not detect port")
		} else {
			port, err := strconv.ParseUint(found.Value, 10, 16)
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
		found, err := db.DatabaseEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			slog.Debug("Could not detect db name", "error", err)
		} else {
			conf.Database = found.Value
			slog.Debug("Found db name", "database", conf.Database)
		}
	}

	// Detect username
	if db, ok := conf.Dialect.(conftypes.DBHasUser); ok && conf.Username == "" {
		found, err := db.UserEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			conf.Username = db.UserDefault()
			slog.Debug("Could not detect user, using default", "error", err, "user", conf.Username)
		} else {
			conf.Username = found.Value
			slog.Debug("Found user", "user", conf.Username)
		}
	}

	foundPassword := detectPassword(ctx, conf)

	// Connection details are detected from the primary since replicas may not expose them
	if conf.Replica && conf.PodName == "" {
		if replica, ok := discovery.FindReplica(ctx, conf.Client, conf.Dialect, conf.DBPod); ok {
			conf.DBPod = replica
		}
	}

	if db, ok := conf.Dialect.(conftypes.DBCanDisableJob); ok && db.DisableJob() {
		must.Must(config.K.Set(consts.FlagCreateJob, false))
	}
	if !conf.CreateJob {
		conf.Host = "127.0.0.1"
		conf.JobPod = conf.DBPod
		if conf.Password != "" {
			conf.PasswordRef = podPasswordRef(conf.DBPod, foundPassword)
			if conf.PasswordRef == "" {
				slog.Debug("Password is not available within the pod; it will be passed in the command")
			}
		}
	}

	return nil
}

func detectPassword(ctx context.Context, conf *conftypes.Global) kubernetes.ConfigValue {
	var found kubernetes.ConfigValue
	if db, ok := conf.Dialect.(conftypes.DBHasPassword); ok && conf.Password == "" {
		var err error
		found, err = db.PasswordEnvs(conf).Search(ctx, conf.Client, conf.DBPod)
		if err != nil {
			slog.Warn("Could not detect password", "error", err)
		} else {
			conf.Password = found.Value
			conf.PasswordSource = found.Source
			slog.Debug("Found password")
		}
	}

	if conf.Password != "" && conf.Log.Mask {
		mask.Add(conf.Password)
	}
	return found
}

var shellVarNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// podPasswordRef returns a shell expression that expands to the password within the pod's default container.
// Returns an empty string if the password is not exposed to that container.
func podPasswordRef(pod corev1.Pod, found kubernetes.ConfigValue) string {
	// The API server only defaults the exec container when the pod has a single container.
	if len(pod.Spec.Containers) != 1 || found.Container != pod.Spec.Containers[0].Name {
		return ""
	}
	switch {
	case found.EnvName != "" && shellVarNameRe.MatchString(found.EnvName):
		return command.Var(found.EnvName).Quote()
	case found.File != "":
		return `"$(cat ` + shellescape.Quote(found.File) + `)"`
	default:
		return ""
	}
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

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	nsLog := slog.With("namespace", conf.Namespace)

	var env []corev1.EnvVar
	var secretRef *corev1.SecretKeySelector
	if conf.Password != "" {
		source := conf.PasswordSource
		if source == nil {
			// The password was not discovered from a Secret or ConfigMap, so store it in a
			// short-lived Secret to keep it out of the job spec and exec requests.
			// The Secret shares the job's name, and is created once the job exists.
			secretRef = &corev1.SecretKeySelector{Key: jobSecretPasswordKey}
			source = &corev1.EnvVarSource{SecretKeyRef: secretRef}
		}
		env = append(env, corev1.EnvVar{Name: jobPasswordEnv, ValueFrom: source})
	}

	job := batchv1.Job{
		Namespace: conf.Namespace,
		Labels:    standardLabels,
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
							Env:             env,
							SecurityContext: defaultContainer.SecurityContext,
						},
					},
					SecurityContext: conf.DBPod.Spec.SecurityContext,
				},
			},
		},
	}

	nsLog.Info("Creating job")
	var err error
	if conf.Job, err = createNamedJob(ctx, conf, &job, name, secretRef); err != nil {
		return err
	}

	if secretRef != nil {
		if err := createJobSecret(ctx, conf, standardLabels); err != nil {
			deleteJob(conf)
			return err
		}
	}

	if conf.Password != "" {
		conf.PasswordRef = command.Var(jobPasswordEnv).Quote()
	}

	if conf.CreateNetworkPolicy {
		jobPodKey, jobPodVal := jobPodNameLabel(conf, conf.Job)
		policy := networkingv1.NetworkPolicy{
			Name:      conf.Job.Name,
			Namespace: conf.Client.Namespace,
			Labels:    standardLabels,
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

const (
	jobNameAttempts      = 8
	jobPasswordEnv       = "KUBEDB_PASSWORD"
	jobSecretPasswordKey = "password"
)

// createNamedJob creates the job with a random name suffix. The name is generated here
// instead of using generateName so that it is known up front and can be shared by other
// resources. Like generateName, conflicts are retried with a new name.
func createNamedJob(
	ctx context.Context,
	conf *conftypes.Global,
	job *batchv1.Job,
	prefix string,
	secretRef *corev1.SecretKeySelector,
) (*batchv1.Job, error) {
	var created *batchv1.Job
	var err error
	for range jobNameAttempts {
		job.Name = prefix + utilrand.String(5)
		if secretRef != nil {
			secretRef.Name = job.Name
		}
		created, err = conf.Client.Jobs().Create(ctx, job, metav1.CreateOptions{})
		if !apierrors.IsAlreadyExists(err) {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return created, nil
}

func createJobSecret(ctx context.Context, conf *conftypes.Global, labels map[string]string) error {
	secret := corev1.Secret{
		Name:      conf.Job.Name,
		Namespace: conf.Namespace,
		Labels:    labels,
		Type:      corev1.SecretTypeOpaque,
		Data:      map[string][]byte{jobSecretPasswordKey: []byte(conf.Password)},
	}

	slog.Debug("Creating password secret", "namespace", conf.Namespace)
	var err error
	conf.JobSecret, err = conf.Client.Secrets().Create(ctx, &secret, metav1.CreateOptions{})
	if err != nil {
		conf.JobSecret = nil
	}
	return err
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
