package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/sanity-io/litter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getClient() (*kubernetes.Clientset, error) {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return config
}

func main() {
	//	var kubeconfig *string
	//if home := homedir.HomeDir(); home != "" {
	//kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//} else {
	//kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	//}
	//flag.Parse()
	//
	//config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//panic(err)
	//}
	client, err := getclient()

	if err != nil {
		log.Fatal("Unable to get client for k8s\n")
	}

	clientset, err := kubernetes.NewForConfig(client)
	if err != nil {
		glog.Errorln(err)
	}

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
			log.Printf("New Job Added to Store: %s", mObj.GetName())
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			log.Printf("Job updated:\n")
			log.Printf("old job:\n")
			litter.Dump(oldObj)
			log.Printf("new job:\n")
			litter.Dump(newObj)
		},
	})

	informer.Run(stopper)
}
