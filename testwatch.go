package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/sanity-io/litter"

	batchv1 "k8s.io/api/batch/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorln(err)
	}

	fs := fields.Set(map[string]string{"label.app": "sba"}).AsSelector()

	watchlist := cache.NewListWatchFromClient(
		clientset.BatchV1().RESTClient(),
		"jobs",
		v1.NamespaceAll,
		fs, //fields.Everything(),
	)
	_, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&batchv1.Job{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("job added: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("job deleted: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("job changed from:\n")
				litter.Dump(oldObj)
				fmt.Printf("job changed to:\n")
				litter.Dump(newObj)
			},
		},
	)
	// I found it in k8s scheduler module. Maybe it's help if you interested in.
	// serviceInformer := cache.NewSharedIndexInformer(watchlist, &v1.Service{}, 0, cache.Indexers{
	//     cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	// })
	// go serviceInformer.Run(stop)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
