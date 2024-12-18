package flags

import (
	"log/slog"
	"os"
	"path/filepath"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const KubeconfigEnv = "KUBECONFIG"

func Kubeconfig(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.KubeconfigFlag, filepath.Join("$HOME", ".kube", "config"), "Paths to the kubeconfig file")
}

func BindKubeconfig(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KubeconfigKey, cmd.Flags().Lookup(consts.KubeconfigFlag)))

	if env := os.Getenv(KubeconfigEnv); env != "" {
		viper.SetDefault(consts.KubeconfigKey, env)
	}
}

func Context(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.ContextFlag, "", "Kubernetes context name")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.ContextFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			kubeconfig := viper.GetString(consts.KubeconfigKey)
			configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{Precedence: filepath.SplitList(kubeconfig)},
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
	cmd.PersistentFlags().StringP(consts.NamespaceFlag, "n", "", "Kubernetes namespace")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.NamespaceFlag,
		func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.NewClientFromCmd(cmd)
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
	cmd.PersistentFlags().String(consts.PodFlag, "", "Perform detection from a pod instead of searching the namespace")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.PodFlag,
		func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.NewClientFromCmd(cmd)
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
	cmd.Flags().StringToString(consts.JobPodLabelsFlag, map[string]string{}, "Pod labels to add to the job")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.JobPodLabelsFlag, cobra.NoFileCompletions))
}

func BindJobPodLabels(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.JobPodLabelsKey, cmd.Flags().Lookup(consts.JobPodLabelsFlag)))
}

func CreateJob(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.CreateJobFlag, true, "Create a job that will run the database client")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.CreateJobFlag, util.BoolCompletion))
}

func BindCreateJob(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.CreateJobKey, cmd.Flags().Lookup(consts.CreateJobFlag)))
}

func CreateNetworkPolicy(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.CreateNetworkPolicyFlag, true, "Creates a network policy allowing the KubeDB job to talk to the database.")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.CreateNetworkPolicyFlag, util.BoolCompletion))
}

func BindCreateNetworkPolicy(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.CreateNetworkPolicyKey, cmd.Flags().Lookup(consts.CreateNetworkPolicyFlag)))
}
