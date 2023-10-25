package kubernetes

import (
	"path/filepath"

	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	ClientSet    kubernetes.Interface
	ClientConfig *rest.Config
	Discovery    *discovery.DiscoveryClient
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

func NewClient(kubeconfig, context, namespace string) (config KubeClient, err error) {
	configLoader := NewConfigLoader(kubeconfig, context)

	config.ClientConfig, err = configLoader.ClientConfig()
	if err != nil {
		return config, err
	}

	config.ClientConfig.UserAgent = rest.DefaultKubernetesUserAgent()

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

	config.Discovery, err = discovery.NewDiscoveryClientForConfig(config.ClientConfig)
	if err != nil {
		return config, err
	}

	return config, err
}

func NewClientFromCmd(cmd *cobra.Command) (KubeClient, error) {
	kubeconfig := viper.GetString(consts.KubeconfigKey)

	context, err := cmd.Flags().GetString(consts.ContextFlag)
	if err != nil {
		panic(err)
	}

	namespace, err := cmd.Flags().GetString(consts.NamespaceFlag)
	if err != nil {
		panic(err)
	}

	return NewClient(kubeconfig, context, namespace)
}
