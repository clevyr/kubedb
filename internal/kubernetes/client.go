package kubernetes

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	ClientSet    *kubernetes.Clientset
	ClientConfig *rest.Config
	Namespace    string
}

func (config KubeClient) Pods() v1.PodInterface {
	return config.ClientSet.CoreV1().Pods(config.Namespace)
}

func (config KubeClient) Secrets() v1.SecretInterface {
	return config.ClientSet.CoreV1().Secrets(config.Namespace)
}

func CreateClient(kubeconfigPath string, namespace string) (config KubeClient, err error) {
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}, nil)

	config.ClientConfig, err = configLoader.ClientConfig()
	if err != nil {
		return config, err
	}

	if namespace == "" {
		config.Namespace, _, err = configLoader.Namespace()
		if err != nil {
			return config, err
		}
	} else {
		config.Namespace = namespace
	}

	config.ClientSet, err = kubernetes.NewForConfig(config.ClientConfig)
	if err != nil {
		return config, err
	}

	return config, err
}

func CreateClientForCmd(cmd *cobra.Command) (KubeClient, error) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	namespace, _ := cmd.Flags().GetString("namespace")
	return CreateClient(kubeconfig, namespace)
}
