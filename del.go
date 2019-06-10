package main

import (
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		log.Fatalf("Couldn't get Kubernetes default config: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Deleting job %s from namespace %s\n", "example-job2", "default")

	fg := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &fg}

	if err := clientset.BatchV1().Jobs("default").Delete("example-job2", &deleteOptions); err != nil {
		fmt.Printf("Failed to delete job: %s\n", err)
	}

	fmt.Printf("err: %s\n", err)
	fmt.Printf("We might have deleted the job\n")
}
