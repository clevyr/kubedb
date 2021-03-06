package flags

import (
	"context"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
)

func Kubeconfig(cmd *cobra.Command) {
	cmd.PersistentFlags().String("kubeconfig", filepath.Join("$HOME", ".kube", "config"), "absolute path to the kubeconfig file")
}

func Context(cmd *cobra.Command) {
	cmd.PersistentFlags().String("context", "", "name of the kubeconfig context to use")
	err := cmd.RegisterFlagCompletionFunc(
		"context",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			kubeconfig, err := cmd.Flags().GetString("kubeconfig")
			if err != nil {
				panic(err)
			}
			conf, err := kubernetes.RawConfig(kubeconfig)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
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
	cmd.PersistentFlags().StringP("namespace", "n", "", "the namespace scope for this CLI request")
	err := cmd.RegisterFlagCompletionFunc(
		"namespace",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.NewClientFromCmd(cmd)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			namespaces, err := client.Namespaces().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			names := make([]string, 0, len(namespaces.Items))
			access := config.NewNamespaceRegexp(cmd.Annotations["access"])
			for _, namespace := range namespaces.Items {
				if access.Match(namespace.Name) {
					names = append(names, namespace.Name)
				}
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}

func Pod(cmd *cobra.Command) {
	cmd.PersistentFlags().String("pod", "", "force a specific pod. if this flag is set, dialect is required.")
	err := cmd.RegisterFlagCompletionFunc(
		"pod",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.NewClientFromCmd(cmd)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			pods, err := client.GetNamespacedPods()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
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
