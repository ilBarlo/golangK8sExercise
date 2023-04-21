package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Ottieni la configurazione del client per accedere al cluster K8s
	config, err := rest.InClusterConfig()
	if err != nil {
		// Se l'applicazione non sta girando dentro un pod, ottieni la configurazione dal file kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", "/Users/il_barlo_/.kube/config")
		if err != nil {
			panic(err.Error())
		}
	}

	// Crea un nuovo clientset per accedere alle API di K8s
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Crea un nuovo informer per monitorare i nuovi pod
	podInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return clientset.CoreV1().Pods("").Watch(context.Background(), metav1.ListOptions{})
			},
		},
		&v1.Pod{},
		0,
		cache.Indexers{},
	)

	// Registra un callback per quando viene creato un nuovo Pod
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Converto l'oggetto generico in un pod specifico
			pod := obj.(*v1.Pod)
			fmt.Printf("New Pod created: %s\n", pod.Name)
		},
	})

	// Avvia l'informer
	podInformer.Run(context.Background().Done())
}
