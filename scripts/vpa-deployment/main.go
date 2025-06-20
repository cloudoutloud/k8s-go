package main

import (
	"context"
	"fmt"
	"log"
	"os"

	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpaClientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

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

	// Create the VPA clientset
	vpaClient, err := vpaClientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating VPA clientset: %s", err.Error())
	}

	// List all Deployment in all namespaces except kube-system
	deployments, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing Deployment: %s", err.Error())
	}
	if len(deployments.Items) == 0 {
		fmt.Println("No deployments found in the cluster. Exiting.")
		os.Exit(0)
	}

	for _, deploy := range deployments.Items {
		// Skip deployments in the kube-system namespace or any other specified namespace
		if deploy.Namespace == "kube-system" || deploy.Namespace == "namespace-to-exclude" {
			continue
		}

		// Create a VPA resource for the Deployment
		updateMode := autoscalingv1.UpdateModeOff
		vpaResource := &autoscalingv1.VerticalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploy.Name + "-vpa",
				Namespace: deploy.Namespace,
			},
			Spec: autoscalingv1.VerticalPodAutoscalerSpec{
				TargetRef: &v1.CrossVersionObjectReference{
					Kind:       "Deployment",
					Name:       deploy.Name,
					APIVersion: "apps/v1",
				},
				UpdatePolicy: &autoscalingv1.PodUpdatePolicy{
					UpdateMode: &updateMode,
				},
			},
		}

		// Create the VPA resource in the cluster
		_, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(deploy.Namespace).Create(context.TODO(), vpaResource, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Error creating VPA for Deployment %s in namespace %s: %s", deploy.Name, deploy.Namespace, err.Error())
		} else {
			fmt.Printf("Created VPA for Deployment %s in namespace %s\n", deploy.Name, deploy.Namespace)
		}
	}
}
