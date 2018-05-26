package kubectl

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/version"
	"os"
)

type Node struct {
	Name string

	Hostname   string
	InternalIP string
	ExternalIP string
}

func getKubeConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return config, nil
	}
	config, err:= rest.InClusterConfig()
	if err!= nil {
		return nil, err
	}
	return config, nil
}

func GetKubeClient() (*kubernetes.Clientset, *rest.Config, error){
	kubeconfig, err := getKubeConfig()
	if err != nil {
		return nil, nil, err
	}
	kubeclient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	return kubeclient, kubeconfig, nil
}

func GetSVCDescribe(rlsName, namespace string) {
	clientset, _, err := GetKubeClient()
	if err != nil {
		return nil, err
	}
	clientset.CoreV1().Services(namespace)
	//TODO
}

func CheckVersion() (*version.Info, error){
	clientset, _, err := GetKubeClient()
	if err != nil {
		return nil, err
	}
	info, err := clientset.ServerVersion()
	if err != nil {
		return nil, err
	}
	return info, nil
}

func GetNodes() ([] Node, error) {
	var nodes [] Node

	clientset, _, err := GetKubeClient()
	if err != nil {
		return nil, err
	}

	items , err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range items.Items {
		node := Node{}
		node.Name = item.Name
		for _, address := range item.Status.Addresses {
			if address.Type == "Hostname" {
				node.Hostname = address.Address
			}
			if address.Type == "InternalIP" {
				node.InternalIP = address.Address
			}

			if address.Type == "ExternalIP" {
				node.ExternalIP = address.Address
			}
		}
		nodes = append(nodes,node)
	}
	return nodes, nil
}