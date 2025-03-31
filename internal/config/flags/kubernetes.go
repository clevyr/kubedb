package flags

import (
	"log/slog"
	"path/filepath"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/completion"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func Kubeconfig(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagKubeConfig, filepath.Join("$HOME", ".kube", "config"),
		"Paths to the kubeconfig file",
	)
}

func Context(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagContext, "", "Kubernetes context name")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagContext,
		func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			if err := completion.LoadConfig(cmd); err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{Precedence: filepath.SplitList(config.Global.Kubeconfig)},
				nil,
			)
			conf, err := configLoader.RawConfig()
			if err != nil {
				slog.Error("Failed to load config", "error", err)
				return nil, cobra.ShellCompDirectiveError
			}
			names := make([]string, 0, len(conf.Contexts))
			for name := range conf.Contexts {
				names = append(names, name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func Namespace(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(consts.FlagNamespace, "n", "", "Kubernetes namespace")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagNamespace,
		func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			if err := completion.LoadConfig(cmd); err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			client, err := kubernetes.NewClient(
				config.Global.Kubeconfig,
				config.Global.Context,
				config.Global.Namespace,
			)
			if err != nil {
				slog.Error("Failed to create Kubernetes client", "error", err)
				return nil, cobra.ShellCompDirectiveError
			}
			namespaces, err := client.Namespaces().List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				slog.Error("Failed to list namespaces", "error", err)
				return nil, cobra.ShellCompDirectiveError
			}
			names := make([]string, 0, len(namespaces.Items))
			for _, namespace := range namespaces.Items {
				names = append(names, namespace.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func Pod(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagPod, "", "Perform detection from a pod instead of searching the namespace")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagPod,
		func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			if err := completion.LoadConfig(cmd); err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			client, err := kubernetes.NewClient(
				config.Global.Kubeconfig,
				config.Global.Context,
				config.Global.Namespace,
			)
			if err != nil {
				slog.Error("Failed to create Kubernetes client", "error", err)
				return nil, cobra.ShellCompDirectiveError
			}
			pods, err := client.GetNamespacedPods(cmd.Context())
			if err != nil {
				slog.Error("Failed to list pods", "error", err)
				return nil, cobra.ShellCompDirectiveError
			}
			names := make([]string, 0, len(pods.Items))
			for _, pod := range pods.Items {
				names = append(names, pod.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		}),
	)
}

func JobPodLabels(cmd *cobra.Command) {
	cmd.Flags().StringToString(consts.FlagJobPodLabels, map[string]string{}, "Pod labels to add to the job")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagJobPodLabels, cobra.NoFileCompletions))
}

func CreateJob(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagCreateJob, true, "Create a job that will run the database client")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagCreateJob, completion.BoolCompletion))
}

func CreateNetworkPolicy(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagCreateNetworkPolicy, true,
		"Creates a network policy allowing the KubeDB job to talk to the database.",
	)
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagCreateNetworkPolicy, completion.BoolCompletion))
}
