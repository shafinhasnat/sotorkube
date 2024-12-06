package main

import (
	"context"
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

type PodStatus struct {
	name    string
	isReady bool
}

func kubeConfig() (KubeClient, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	var kubeconfig string = filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
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

func (k *KubeClient) listPods(namespace string, failing_only bool) []PodStatus {
	pods, err := k.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	var podinfo []PodStatus
	for _, pod := range pods.Items {
		var isReady bool = true
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "False" {
				isReady = false
				break
			}
		}
		if failing_only {
			if !isReady {
				podinfo = append(podinfo, PodStatus{pod.Name, isReady})
			}
		} else {
			podinfo = append(podinfo, PodStatus{pod.Name, isReady})
		}
	}
	return podinfo
}

func main() {
	client, err := kubeConfig()
	if err != nil {
		panic(err.Error())
	}
	var namespace string = ``
	pods := client.listPods(namespace, false)
	fmt.Print(pods)
}
