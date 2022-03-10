package flags

import (
	"context"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func Kubeconfig(cmd *cobra.Command) {
	var kubeconfigDefault string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigDefault = filepath.Join(home, ".kube", "config")
	}
	cmd.PersistentFlags().String("kubeconfig", kubeconfigDefault, "absolute path to the kubeconfig file")
}

func Namespace(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("namespace", "n", "", "the namespace scope for this CLI request")
	err := cmd.RegisterFlagCompletionFunc(
		"namespace",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.CreateClientForCmd(cmd)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			namespaces, err := client.Namespaces().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var names []string
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
	cmd.PersistentFlags().String("pod", "", "force a specific pod. if this flag is set, grammar is required.")
	err := cmd.RegisterFlagCompletionFunc(
		"pod",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := kubernetes.CreateClientForCmd(cmd)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			pods, err := kubernetes.GetNamespacedPods(client)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var names []string
			for _, pod := range pods.Items {
				names = append(names, pod.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(err)
	}
}
