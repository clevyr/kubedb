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
	cmd.PersistentFlags().String(consts.FlagKubeConfig, filepath.Join("$HOME", ".kube", "config"), "Paths to the kubeconfig file")
}

func BindKubeconfig(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyKubeConfig, cmd.Flags().Lookup(consts.FlagKubeConfig)))

	if env := os.Getenv(KubeconfigEnv); env != "" {
		viper.SetDefault(consts.KeyKubeConfig, env)
	}
}

func Context(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.FlagContext, "", "Kubernetes context name")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagContext,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			kubeconfig := viper.GetString(consts.KeyKubeConfig)
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
	cmd.PersistentFlags().StringP(consts.FlagNamespace, "n", "", "Kubernetes namespace")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagNamespace,
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
	cmd.PersistentFlags().String(consts.FlagPod, "", "Perform detection from a pod instead of searching the namespace")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagPod,
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
	cmd.Flags().StringToString(consts.FlagJobPodLabels, map[string]string{}, "Pod labels to add to the job")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagJobPodLabels, cobra.NoFileCompletions))
}

func BindJobPodLabels(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyJobPodLabels, cmd.Flags().Lookup(consts.FlagJobPodLabels)))
}

func CreateJob(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagCreateJob, true, "Create a job that will run the database client")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagCreateJob, util.BoolCompletion))
}

func BindCreateJob(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyCreateJob, cmd.Flags().Lookup(consts.FlagCreateJob)))
}

func CreateNetworkPolicy(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagCreateNetworkPolicy, true, "Creates a network policy allowing the KubeDB job to talk to the database.")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagCreateNetworkPolicy, util.BoolCompletion))
}

func BindCreateNetworkPolicy(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.KeyCreateNetworkPolicy, cmd.Flags().Lookup(consts.FlagCreateNetworkPolicy)))
}
