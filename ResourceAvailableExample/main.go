package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	// Indicare il percorso del file di configurazione del cluster Kubernetes
	kubeconfig := flag.String("kubeconfig", "/Users/il_barlo_/.kube/config", "location to our kubeconfig file")

	// Creare la configurazione del client Go utilizzando il file di configurazione
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Crea il clientset dei metrics
	metricsClientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Crea il clientset di K8s
	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Crea un ticker che gira ogni 10 secondi
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		// Ottieni la lista di nodi nel cluster
		nodeList, err := kubeClientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		// Per ogni nodo, recupera le metriche e stampa le risorse disponibili
		for _, node := range nodeList.Items {
			metrics, err := metricsClientset.MetricsV1beta1().NodeMetricses().Get(context.Background(), node.Name, metav1.GetOptions{})
			if err != nil {
				panic(err.Error())
			}

			nodeInfo := node.Status.NodeInfo.Architecture
			cpuQuantity := metrics.Usage.Cpu()
			memoryQuantity := metrics.Usage.Memory()

			// Converti le quantità di CPU e memoria in valori leggibili
			cpu := cpuQuantity.MilliValue()
			memory := memoryQuantity.Value() / 1024 / 1024 // Converti da byte a megabyte

			fmt.Printf("Nodo %s - Architecture %s, CPU disponibili: %d milliCPU, Memoria disponibile: %d MB\n", node.Name, nodeInfo, cpu, memory)
		}
	}
}
