package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	currentTime := time.Now()

	// Load kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	// Get all pods in all namespaces
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	// for _, pod := range pods.Items {
	// 	fmt.Printf("Namespace: %s, Pod Name: %s, \n",
	// 		pod.Namespace,
	// 		pod.Name)
	// }

	fmt.Println(currentTime.Format("2006-01-02 15:04:05"), "Pods running without resource requests and limits set:")
	for _, pod := range pods.Items {
		noResourcesSet := true
		for _, container := range pod.Spec.Containers {
			resources := container.Resources
			if resources.Requests != nil && resources.Limits != nil {
				_, cpuRequestSet := resources.Requests[corev1.ResourceCPU]
				_, memoryRequestSet := resources.Requests[corev1.ResourceMemory]
				_, cpuLimitSet := resources.Limits[corev1.ResourceCPU]
				_, memoryLimitSet := resources.Limits[corev1.ResourceMemory]

				if cpuRequestSet || memoryRequestSet || cpuLimitSet || memoryLimitSet {
					noResourcesSet = false
					break
				}
			}
		}

		if noResourcesSet {
			fmt.Printf("Namespace: %s, Pod: %s\n", pod.Namespace, pod.Name)
		}
	}

	// Get all jobs in all namespaces
	jobs, err := clientset.BatchV1().Jobs("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing jobs: %v", err)
	}

	// for _, job := range jobs.Items {
	// 	fmt.Printf("Namespace: %s, Job Name: %s, Completions: %d, Active: %d\n",
	// 		job.Namespace,
	// 		job.Name,
	// 		*job.Spec.Completions,
	// 		job.Status.Active)
	// }

	fmt.Println(currentTime.Format("2006-01-02 15:04:05"), "Jobs without resources requests and limits set:")
	for _, job := range jobs.Items {
		noResourcesSet := true
		for _, container := range job.Spec.Template.Spec.Containers {
			resources := container.Resources
			if resources.Requests != nil && resources.Limits != nil {
				_, cpuRequestSet := resources.Requests[corev1.ResourceCPU]
				_, memoryRequestSet := resources.Requests[corev1.ResourceMemory]
				_, cpuLimitSet := resources.Limits[corev1.ResourceCPU]
				_, memoryLimitSet := resources.Limits[corev1.ResourceMemory]

				if cpuRequestSet || memoryRequestSet || cpuLimitSet || memoryLimitSet {
					noResourcesSet = false
					break
				}
			}
		}
		if noResourcesSet {
			fmt.Printf("Namespace: %s, Job: %s\n", job.Namespace, job.Name)
		}
	}
}
