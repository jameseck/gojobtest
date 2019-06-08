package main

import (
	"fmt"
	"github.com/sanity-io/litter"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
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

	fmt.Printf("Setting up the informer\n")
	//factory := informers.NewSharedInformerFactory(clientset, 0)

	factory := informers.NewFilteredSharedInformerFactory(
		clientset,
		0,
		"default",
		func(opt *metav1.ListOptions) {
			opt.LabelSelector = "app=sba"
		},
	)

	informer := factory.Batch().V1().Jobs().Informer()
	stopper := make(chan struct{})
	defer close(stopper)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// "k8s.io/apimachinery/pkg/apis/meta/v1" provides an Object
			// interface that allows us to get metadata easily
			mObj := obj.(metav1.Object)
			fmt.Printf("New Job Added to Store: %s\n", mObj.GetName())
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			fmt.Printf("Job updated:\n")
			fmt.Printf("old job:\n")
			litter.Dump(oldObj)
			fmt.Printf("new job:\n")
			litter.Dump(newObj)
		},
	})

	informer.Run(stopper)
}
