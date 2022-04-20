package kubernetes

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type KubeClient struct {
	ClientSet    kubernetes.Interface
	ClientConfig *rest.Config
	Namespace    string
}

func (client KubeClient) Namespaces() v1.NamespaceInterface {
	return client.ClientSet.CoreV1().Namespaces()
}

func (client KubeClient) Pods() v1.PodInterface {
	return client.ClientSet.CoreV1().Pods(client.Namespace)
}

func (client KubeClient) Secrets() v1.SecretInterface {
	return client.ClientSet.CoreV1().Secrets(client.Namespace)
}

func NewClient(kubeconfigPath, context, namespace string) (config KubeClient, err error) {
	var overrides *clientcmd.ConfigOverrides
	if context != "" {
		overrides = &clientcmd.ConfigOverrides{CurrentContext: context}
	}

	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		overrides,
	)

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

func NewClientFromCmd(cmd *cobra.Command) (KubeClient, error) {
	kubeconfig, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		panic(err)
	}

	context, err := cmd.Flags().GetString("context")
	if err != nil {
		panic(err)
	}

	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		panic(err)
	}

	return NewClient(kubeconfig, context, namespace)
}

func RawConfig(kubeconfigPath string) (api.Config, error) {
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}, nil,
	)
	return configLoader.RawConfig()
}
