package main

import (
	"aks-coach/internal/kube"
	"aks-coach/internal/render"
	"aks-coach/internal/version"
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// entry point
func main() {
	namespace := flag.String("namespace", "default", "Kubernetes namespace to inspect")
	allNamespaces := flag.Bool("all-namespaces", false, "If true, list deployments across all namespaces")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("aks-coach %s\n", version.String())
		return
	}

	var scope kube.Scope
	scope.AllNamespaces = *allNamespaces
	scope.Namespace = *namespace

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientset, err := kube.NewClient()
	if err != nil {
		log.Fatalf("failed to create Kubernetes client: %v", err)
	}

	var deployments *appsv1.DeploymentList

	deployments, err = clientset.AppsV1().Deployments(scope.NS()).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("failed to list deployments in %s: %v", scope.Label(), err)
	}

	if len(deployments.Items) == 0 {
		fmt.Printf("No deployments found in %q\n", scope.Label())
		return
	}

	hpaMap, err := kube.ListHPAs(ctx, clientset, scope)
	if err != nil {
		log.Fatalf("failed to list HPA objects in scope %s: %v", scope.Label(), err)
	}

	render.PrintTable(scope.Label(), deployments.Items, hpaMap)
}
