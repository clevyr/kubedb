package flags

import (
	"os"
	"path/filepath"

	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/kubernetes"
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
	if err := viper.BindPFlag(consts.KubeconfigKey, cmd.Flags().Lookup(consts.KubeconfigFlag)); err != nil {
		panic(err)
	}

	if env := os.Getenv(KubeconfigEnv); env != "" {
		viper.SetDefault(consts.KubeconfigKey, env)
	}
}

func Context(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.ContextFlag, "", "Kubernetes context name")
	err := cmd.RegisterFlagCompletionFunc(
		consts.ContextFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			kubeconfig := viper.GetString(consts.KubeconfigKey)
			configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{Precedence: filepath.SplitList(kubeconfig)},
				nil,
			)
			conf, err := configLoader.RawConfig()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			names := make([]string, 0, len(conf.Contexts))
			for name := range conf.Contexts {
				names = append(names, name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func Namespace(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(consts.NamespaceFlag, "n", "", "Kubernetes namespace")
	err := cmd.RegisterFlagCompletionFunc(
		consts.NamespaceFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.NewClientFromCmd(cmd)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			namespaces, err := client.Namespaces().List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			names := make([]string, 0, len(namespaces.Items))
			for _, namespace := range namespaces.Items {
				names = append(names, namespace.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func Pod(cmd *cobra.Command) {
	cmd.PersistentFlags().String(consts.PodFlag, "", "Perform detection from a pod instead of searching the namespace")
	err := cmd.RegisterFlagCompletionFunc(
		"pod",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.NewClientFromCmd(cmd)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			pods, err := client.GetNamespacedPods(cmd.Context())
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			names := make([]string, 0, len(pods.Items))
			for _, pod := range pods.Items {
				names = append(names, pod.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func JobPodLabels(cmd *cobra.Command) {
	cmd.Flags().StringToString(consts.JobPodLabelsFlag, map[string]string{}, "Pod labels to add to the job")
}

func BindJobPodLabels(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.JobPodLabelsKey, cmd.Flags().Lookup(consts.JobPodLabelsFlag)); err != nil {
		panic(err)
	}
}

func NoJob(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.NoJobFlag, false, "Database commands will be run in the database pod instead of a dedicated job")
}

func BindNoJob(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.NoJobKey, cmd.Flags().Lookup(consts.NoJobFlag)); err != nil {
		panic(err)
	}
}

func CreateNetworkPolicy(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.CreateNetworkPolicyFlag, true, "Creates a network policy allowing the KubeDB job to talk to the database.")
}

func BindCreateNetworkPolicy(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.CreateNetworkPolicyKey, cmd.Flags().Lookup(consts.CreateNetworkPolicyFlag)); err != nil {
		panic(err)
	}
}
