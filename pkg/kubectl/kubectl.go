package kubectl

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"os"
)

type Node struct {
	Name string

	Hostname   string
	InternalIP string
	ExternalIP string
}

func createClient() (*kubernetes.Clientset, error){
	//kubernetes kubeconfig set
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(config)
	}

	//in kubernetes cluster without kubeconfig
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func GetNodes() ([] Node, error) {
	var nodes [] Node

	clientset, err := createClient()
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