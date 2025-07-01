package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	sourceClusterContext     = flag.String("sourceCluster", "", "Context name to use for source cluster")
	desinationClusterContext = flag.String("desinationCluster", "", "Context name to use for destination cluster")
	namespace                = flag.String("namespace", "", "Namespace to migrate secrets from and to")
	secretPrefix             = flag.String("secret-prefix", "", "Prefix of secrets to migrate")
)

func main() {
	flag.Parse()

	if *secretPrefix == "" {
		log.Fatal("secret-prefix is required")
	}

	if *namespace == "" {
		log.Fatal("namespace is required")
	}

	if *sourceClusterContext == "" {
		log.Fatal("Source cluster context is required")
	}

	if *desinationClusterContext == "" {
		log.Fatal("Destination cluster context is required")
	}

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
		log.Printf("KUBECONFIG environment variable not set, using default path: %s", kubeconfigPath)
	}

	clientsetA, err := newClient(kubeconfigPath, *sourceClusterContext)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	clientsetB, err := newClient(kubeconfigPath, *desinationClusterContext)
	if err != nil {
		log.Fatalf("Failed to create client for cluster B: %v", err)
	}

	secrets, err := listSecrets(clientsetA, *namespace, *secretPrefix)
	if err != nil {
		log.Fatalf("Failed to list secrets in cluster: %v", err)
	}

	if len(secrets) > 0 {
		fmt.Println("Secrets found:")
		for _, secret := range secrets {
			fmt.Printf("- %s\n", secret.Name)
		}
	} else {
		Prefix := *secretPrefix
		fmt.Printf("No secrets found to migrate, matching the specified prefix of: %s", Prefix)
    os.Exit(0)
	}

	for _, secret := range secrets {
		err := migrateSecret(clientsetB, *namespace, secret)
		contextName := *desinationClusterContext
		if err != nil {
			log.Printf("Failed to migrate secret %s to cluster (context: %s): %v", secret.Name, contextName, err)
		} else {
			fmt.Printf("Successfully migrated secret %s to cluster (context: %s)\n", secret.Name, contextName)
		}
	}

	fmt.Println("Secrets migration completed.")
}

// Creates a Kubernetes client from a kubeconfig file and a specific context.
func newClient(kubeconfigPath, contextName string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		}).ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return clientSet, nil
}

func listSecrets(client kubernetes.Interface, namespace, prefix string) ([]corev1.Secret, error) {
	secrets, err := client.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var filteredSecrets []corev1.Secret
	for _, secret := range secrets.Items {
		if strings.HasPrefix(secret.Name, prefix) {
			filteredSecrets = append(filteredSecrets, secret)
		}
	}

	return filteredSecrets, nil
}

// migrateSecret creates a secret in destination cluster based on the secret from source cluster.
func migrateSecret(client kubernetes.Interface, namespace string, secret corev1.Secret) error {
	// Create a copy of the secret to avoid modifying the original
	secretCopy := secret.DeepCopy()

	// Remove unwanted resource version and UID so that the secret can be created.
	secretCopy.ObjectMeta.ResourceVersion = ""
	secretCopy.ObjectMeta.UID = ""

	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), secretCopy, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}
