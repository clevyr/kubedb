package kubernetes

import (
	"path/filepath"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Loads auth plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	ClientSet    kubernetes.Interface
	ClientConfig *rest.Config
	Discovery    *discovery.DiscoveryClient
	Context      string
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

func (client KubeClient) Jobs() batchv1.JobInterface {
	return client.ClientSet.BatchV1().Jobs(client.Namespace)
}

func (client KubeClient) NetworkPolicies() networkingv1.NetworkPolicyInterface {
	return client.ClientSet.NetworkingV1().NetworkPolicies(client.Namespace)
}

func (client KubeClient) ConfigMaps() v1.ConfigMapInterface {
	return client.ClientSet.CoreV1().ConfigMaps(client.Namespace)
}

func NewConfigLoader(kubeconfig, context string) clientcmd.ClientConfig {
	var overrides *clientcmd.ConfigOverrides
	if context != "" {
		overrides = &clientcmd.ConfigOverrides{CurrentContext: context}
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: filepath.SplitList(kubeconfig)},
		overrides,
	)
}

func NewClient(kubeconfig, context, namespace string) (KubeClient, error) {
	configLoader := NewConfigLoader(kubeconfig, context)
	var client KubeClient

	if rawConfig, err := configLoader.RawConfig(); err == nil {
		client.Context = rawConfig.CurrentContext
	}

	var err error
	if client.ClientConfig, err = configLoader.ClientConfig(); err != nil {
		return client, err
	}

	client.ClientConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	if namespace == "" {
		if client.Namespace, _, err = configLoader.Namespace(); err != nil {
			return client, err
		}
	} else {
		client.Namespace = namespace
	}

	if client.ClientSet, err = kubernetes.NewForConfig(client.ClientConfig); err != nil {
		return client, err
	}

	if client.Discovery, err = discovery.NewDiscoveryClientForConfig(client.ClientConfig); err != nil {
		return client, err
	}

	return client, err
}
