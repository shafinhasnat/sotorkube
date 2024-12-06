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
	namespace    string
	name         string
	isReady      bool
	restartCount int32
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
		var restartCount int32
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restartCount = containerStatus.RestartCount
			if !containerStatus.Ready {
				isReady = false
				break
			}
		}
		if failing_only {
			if !isReady {
				podinfo = append(podinfo, PodStatus{pod.Namespace, pod.Name, isReady, restartCount})
			}
		} else {
			podinfo = append(podinfo, PodStatus{pod.Namespace, pod.Name, isReady, restartCount})
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
	fmt.Print(client.listPods(namespace, false))

}
