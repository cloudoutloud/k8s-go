package main

import (
	"context"
	"fmt"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Load the kubeconfig file to connect to the cluster
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	// Create the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	// Get the list of deployments in all namespaces
	deployments, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing deployments: %v", err)
	}

	// Iterate through deployments and print those with less than 2 replicas
	fmt.Println("Deployments with less than 2 replicas:")
	for _, deployment := range deployments.Items {
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas < 2 {
			fmt.Printf("Namespace: %s, Deployment Name: %s, Replicas: %d\n",
				deployment.Namespace, deployment.Name, *deployment.Spec.Replicas)
		}
	}
}
