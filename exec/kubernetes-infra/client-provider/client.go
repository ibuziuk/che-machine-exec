package client_provider

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesClientProvider struct {
	config *rest.Config
	client *kubernetes.Clientset
}

func New() *KubernetesClientProvider {
	config := createConfig()
	return &KubernetesClientProvider{
		config:config,
		client:createClient(config),
	}
}

// Create in clients set to get access to the kubernetes api inside pod.
func createConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return config
}

// Create configuration to get access to the kubernetes configuration inside pod
func createClient(config *rest.Config) *kubernetes.Clientset {
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// Get Kubernetes configuration inside cluster pod.
func (clientProvider *KubernetesClientProvider) GetKubernetesConfig() *rest.Config {
	return clientProvider.config
}

// Get Kubernetes client inside cluster pod.
func (clientProvider *KubernetesClientProvider) GetKubernetesClient() *kubernetes.Clientset {
	return clientProvider.client
}
