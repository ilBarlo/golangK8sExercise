package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type NodeInfo struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Arch   string `json:"arch"`
}

var nodeInfoMap = make(map[string]NodeInfo)

func main() {

	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", "/Users/il_barlo_/.kube/config")
		if err != nil {
			panic(err.Error())
		}
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	// Create a node informer to watch for changes to the nodes
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()

	// Add an event handler to update the nodeInfoMap whenever a node is added, updated, or deleted
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addNode,
		UpdateFunc: updateNode,
		DeleteFunc: deleteNode,
	})

	// Start the node informer
	go nodeInformer.Run(wait.NeverStop)

	// Wait until the node informer is synced
	if !cache.WaitForCacheSync(wait.NeverStop, nodeInformer.HasSynced) {
		log.Fatalf("Error syncing node informer")
	}

	// Create an HTTP handler to expose the nodeInfoMap through an API
	http.HandleFunc("/maps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodeInfoMap)
	})

	// Start the HTTP server
	log.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func addNode(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Printf("Object is not a node: %v", obj)
		return
	}
	updateNodeInfo(node)
}

func updateNode(oldObj, newObj interface{}) {
	node, ok := newObj.(*v1.Node)
	if !ok {
		log.Printf("Object is not a node: %v", newObj)
		return
	}
	updateNodeInfo(node)
}

func deleteNode(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Printf("Object is not a node: %v", obj)
		return
	}
	delete(nodeInfoMap, node.Name)
}

func updateNodeInfo(node *v1.Node) {
	nodeInfo := NodeInfo{}
	nodeInfo.CPU = node.Status.Capacity.Cpu().String()
	nodeInfo.Memory = node.Status.Capacity.Memory().String()
	nodeInfo.Arch = node.Status.NodeInfo.Architecture
	nodeInfoMap[node.Name] = nodeInfo
}
