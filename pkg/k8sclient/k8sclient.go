package k8sclient

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	K8sClient     *kubernetes.Clientset
	DynamicClient dynamic.Interface
	RestConfig    *rest.Config
)

func init() {
	var err error
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		RestConfig, err = rest.InClusterConfig()
	} else {
		RestConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		panic(err)
	}

	K8sClient, err = kubernetes.NewForConfig(RestConfig)
	if err != nil {
		panic(err)
	}

	DynamicClient, err = dynamic.NewForConfig(RestConfig)
	if err != nil {
		panic(err)
	}
}
