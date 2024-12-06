package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	clientset kubernetes.Clientset
}

func kubeConfig() (KubeClient, error) {
	var kubeconfig *string
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	if home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("Kubeconfig", "", "Absolute path to the kubeconfig")
	}
	flag.Parse()
	fmt.Printf("Using kubeconfig: %s\n", *kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return KubeClient{}, err
	}
	kubeclient := KubeClient{clientset: *clientset}
	return kubeclient, nil
}

func (k *KubeClient) listPods(namespace string) {
	pods, err := k.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range pods.Items {
		var isReady bool = true
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "False" {
				isReady = false
				break
			}
		}
		fmt.Println(pod.Name, isReady)
	}
}

func main() {
	client, err := kubeConfig()
	if err != nil {
		panic(err.Error())
	}
	var namespace string = ``
	client.listPods(namespace)
}
