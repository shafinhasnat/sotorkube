package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	clientset kubernetes.Clientset
}

type PodStatus struct {
	Namespace    string
	Name         string
	IsReady      bool
	RestartCount int32
}

type PodState struct {
	gorm.Model
	Name         string
	MailCount    int32
	MailSentTime time.Time
}

func kubeConfig() (KubeClient, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return KubeClient{}, err
	}
	var kubeconfig string = filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return KubeClient{}, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return KubeClient{}, err
	}
	kubeclient := KubeClient{clientset: *clientset}
	return kubeclient, nil
}

func (k *KubeClient) listPods(namespace string, failing_only bool) ([]PodStatus, error) {
	pods, err := k.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []PodStatus{}, err
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
	return podinfo, nil
}

func podName(namespace string, name string) string {
	podname := fmt.Sprint(namespace, "/", name)
	return podname
}

func watchFailingPods(client KubeClient, db *gorm.DB) {
	var namespace string = ``
	all_pods, err := client.listPods(namespace, false)
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range all_pods {
		var podname string = podName(pod.Namespace, pod.Name)
		var podinfo PodState
		query := db.First(&podinfo, "name = ?", podname)
		if query.Error == nil { // Existing entry found in db
			if pod.IsReady {
				db.Delete(&podinfo, "name = ?", podname)
			}
		} else {
			if !pod.IsReady {
				if pod.RestartCount > 3 {
					db.Create(&PodState{Name: podname, MailCount: 0, MailSentTime: time.Now().UTC()})
				}
			}
		}
	}
}

func sendAlert(db *gorm.DB) {
	var unalertedPods []PodState
	db.Find(&unalertedPods, "mail_count = ?", 0)
	var pods_sb strings.Builder
	if len(unalertedPods) == 0 {
		fmt.Println("No new failing pods")
	} else {
		for _, pod := range unalertedPods {
			pod.MailCount = 1
			pod.MailSentTime = time.Now().UTC()
			db.Save(&pod)
			pods_sb.WriteString(pod.Name)
			pods_sb.WriteString("\\n\\n")
		}
		webhook(pods_sb.String())
	}
}

func webhook(data string) int {
	var webhook_url string = os.Getenv("WEBHOOK_URL")
	var webhook_title string = os.Getenv("WEBHOOK_TITLE")
	var webhook_body string = os.Getenv("WEBHOOK_BODY")
	formatted_title := strings.ReplaceAll(webhook_body, "<TITLE>", webhook_title)
	var message string = fmt.Sprint("Failing pod alert\\n\\n", data)
	formatted_body := strings.ReplaceAll(formatted_title, "<MESSAGE>", message)
	var b bytes.Buffer
	b.WriteString(formatted_body)
	resp, err := http.Post(webhook_url, "application/json; charset=utf-8", &b)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()
	status_code := resp.StatusCode
	return status_code
}

func main() {
	db, err := gorm.Open(sqlite.Open("podstate.db"), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	db.AutoMigrate(&PodState{})
	client, err := kubeConfig()
	if err != nil {
		panic(err.Error())
	}
	interval_s := os.Getenv("INTERVAL")
	interval, _ := strconv.Atoi(interval_s)
	for {
		watchFailingPods(client, db)
		sendAlert(db)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
