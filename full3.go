package main

import (
	"fmt"
	"github.com/sanity-io/litter"
	"log"

	batchv1 "k8s.io/api/batch/v1"
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
		DeleteFunc: func(obj interface{}) {
			fmt.Printf("Job deletefunc called\n")
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			fmt.Printf("Job updated:\n")
			//fmt.Printf("old job:\n")
			//litter.Dump(oldObj)
			//fmt.Printf("new job:\n")
			//litter.Dump(newObj)
			oldJob, _ := newObj.(*batchv1.Job)
			newJob, _ := newObj.(*batchv1.Job)

			//  Here we want to see if the job succeeded or failed (1 = true)
			//  If both are 0, do nothing
			// maybe we should check for Active instead?
			if newJob.Status.Active == 1 {
				fmt.Printf("Doing nothing as job is still active\n")
				return
			}
			fmt.Printf("Checking status of job\n")
			if newJob.Status.Active != 0 && newJob.Status.Failed == 1 {
				fmt.Printf("The job failed\n")
			}
			if newJob.Status.Active != 0 && newJob.Status.Succeeded == 1 {
				fmt.Printf("The job succeeded\n")
			}
			litter.Dump(oldJob.Status.Succeeded)
			litter.Dump(newJob.Status.Succeeded)
			// Here we can delete the job?

			fmt.Printf("Deleting job %s from namespace %s\n", newJob.Name, newJob.Namespace)

			fg := metav1.DeletePropagationForeground
			deleteOptions := metav1.DeleteOptions{PropagationPolicy: &fg}

			if err := clientset.BatchV1().Jobs(newJob.Namespace).Delete(newJob.Name, &deleteOptions); err != nil {
				fmt.Printf("Failed to delete job: %s\n", err)
			}
			fmt.Printf("Job deleted\n")
		},
	})

	informer.Run(stopper)
}
