package main

import (
	"aks-coach/internal/version"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientset, err := newKubeClient()
	if err != nil {
		log.Fatalf("failed to create Kubernetes client: %v", err)
	}

	scopeLabel := ""
	var deployments *appsv1.DeploymentList

	if *allNamespaces {
		scopeLabel = "all namespaces"
		deployments, err = clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Fatalf("failed to list deployments in all namespaces: %v", err)
		}
	} else {
		scopeLabel = fmt.Sprintf("namespace %q", *namespace)
		deployments, err = clientset.AppsV1().Deployments(*namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Fatalf("failed to list deployments in namespace %q: %v", *namespace, err)
		}
	}

	if len(deployments.Items) == 0 {
		fmt.Printf("No deployments found in namespace %q\n", *namespace)
		return
	}

	hpaMap, err := listHPAsForScope(ctx, clientset, *allNamespaces, *namespace)
	if err != nil {
		log.Fatalf("failed to list HPA objects in scope %s: %v", scopeLabel, err)
	}

	printDeploymentCapacityReport(scopeLabel, deployments.Items, hpaMap)
}

// newKubeClient tries in-cluster config, then falls back to $KUBECONFIG or ~/.kube/config.
func newKubeClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return nil, fmt.Errorf("cannot find home directory for kubeconfig: %w", homeErr)
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("cannot build kubeconfig: %w", err)
		}
	}
	return kubernetes.NewForConfig(config)
}

// printDeploymentCapacityReport prints a simple table of CPU/mem for each deployment.
func printDeploymentCapacityReport(
	scopeLabel string,
	deployments []appsv1.Deployment,
	hpaMap map[string]*autoscalingv2.HorizontalPodAutoscaler) {

	fmt.Printf("Scope: %s\n\n", scopeLabel)
	fmt.Printf("%-16.16s %-32.32s %8s %12s %13s %13s %15s %8s %8s %14.14s\n",
		"NAMESPACE", "NAME", "REPLICAS",
		"CPU_REQ(m)", "CPU_LIMIT(m)", "MEM_REQ(Mi)", "MEM_LIMIT(Mi)",
		"HPA_MIN", "HPA_MAX", "HPA_TARGET")
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------------------------")

	for _, d := range deployments {
		replicas := int32(1)
		if d.Spec.Replicas != nil {
			replicas = *d.Spec.Replicas
		}

		perPodCPUReq, perPodCPULimit, perPodMemReq, perPodMemLimit := aggregatePodResources(d)

		totalCPUReq := perPodCPUReq * float64(replicas)
		totalCPULimit := perPodCPULimit * float64(replicas)
		totalMemReq := perPodMemReq * float64(replicas)
		totalMemLimit := perPodMemLimit * float64(replicas)

		key := fmt.Sprintf("%s/%s", d.Namespace, d.Name)
		hpaMin := "-"
		hpaMax := "-"
		hpaTarget := "-"

		if h := hpaMap[key]; h != nil {
			if h.Spec.MinReplicas != nil {
				hpaMin = fmt.Sprintf("%d", *h.Spec.MinReplicas)
			} else {
				hpaMin = ""
			}

			hpaMax = fmt.Sprintf("%d", hpaMap[key].Spec.MaxReplicas)
			hpaTarget = summarizeHPACPUTarget(h)
		}

		fmt.Printf("%-16.16s %-32.32s %8d %12.0f %13.0f %13.0f %15.0f %8s %8s %14.14s\n",
			d.Namespace,
			d.Name,
			replicas,
			totalCPUReq,
			totalCPULimit,
			totalMemReq,
			totalMemLimit,
			hpaMin,
			hpaMax,
			hpaTarget,
		)
	}
}

// aggregatePodResources sums CPU/mem requests/limits across all containers in the pod template.
//
// Returns values as:
//
//	cpuReqMilli, cpuLimitMilli, memReqMiB, memLimitMiB
func aggregatePodResources(d appsv1.Deployment) (float64, float64, float64, float64) {
	var (
		cpuReqMilli   float64
		cpuLimitMilli float64
		memReqMiB     float64
		memLimitMiB   float64
	)

	for _, c := range d.Spec.Template.Spec.Containers {
		reqs := c.Resources.Requests
		limits := c.Resources.Limits

		if cpuQty, ok := reqs["cpu"]; ok {
			cpuReqMilli += cpuQty.AsApproximateFloat64() * 1000
		}
		if cpuQty, ok := limits["cpu"]; ok {
			cpuLimitMilli += cpuQty.AsApproximateFloat64() * 1000
		}

		if memQty, ok := reqs["memory"]; ok {
			memReqMiB += quantityToMiB(memQty)
		}
		if memQty, ok := limits["memory"]; ok {
			memLimitMiB += quantityToMiB(memQty)
		}
	}

	return cpuReqMilli, cpuLimitMilli, memReqMiB, memLimitMiB
}

// listHPAsForScope returns a map keyed by "namespace/name" of the target Deployment.
func listHPAsForScope(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	allNamespaces bool,
	namespace string,
) (map[string]*autoscalingv2.HorizontalPodAutoscaler, error) {

	ns := ""
	if !allNamespaces {
		ns = namespace
	}

	hpas, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make(map[string]*autoscalingv2.HorizontalPodAutoscaler)

	for i := range hpas.Items {
		hpa := &hpas.Items[i]
		if hpa.Spec.ScaleTargetRef.Kind == "Deployment" {
			key := fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Spec.ScaleTargetRef.Name)
			result[key] = hpa
		}
	}

	return result, nil
}

// summarizeHPACPUTarget returns a string like "cpu: 28%/60%" for CPU resource metrics,
// or "-" if no CPU metric is configured/found.
func summarizeHPACPUTarget(h *autoscalingv2.HorizontalPodAutoscaler) string {
	if h == nil {
		return "-"
	}

	var targetCPU string
	var currentCPU string

	// Find CPU metric in spec (target)
	for _, metric := range h.Spec.Metrics {
		if metric.Type == autoscalingv2.ResourceMetricSourceType &&
			metric.Resource != nil &&
			metric.Resource.Name == corev1.ResourceCPU {

			switch metric.Resource.Target.Type {
			case autoscalingv2.UtilizationMetricType:
				if metric.Resource.Target.AverageUtilization != nil {
					targetCPU = fmt.Sprintf("%d%%", *metric.Resource.Target.AverageUtilization)
				}
			case autoscalingv2.AverageValueMetricType:
				if metric.Resource.Target.AverageValue != nil {
					targetCPU = metric.Resource.Target.AverageValue.String()
				}
			case autoscalingv2.ValueMetricType:
				if metric.Resource.Target.Value != nil {
					targetCPU = metric.Resource.Target.Value.String()
				}
			}
			break
		}
	}

	// Find CPU metric in status (current)
	for _, currentMetric := range h.Status.CurrentMetrics {
		if currentMetric.Type == autoscalingv2.ResourceMetricSourceType &&
			currentMetric.Resource != nil &&
			currentMetric.Resource.Name == corev1.ResourceCPU {

			if currentMetric.Resource.Current.AverageUtilization != nil {
				currentCPU = fmt.Sprintf("%d%%", *currentMetric.Resource.Current.AverageUtilization)
			} else if currentMetric.Resource.Current.AverageValue != nil {
				currentCPU = currentMetric.Resource.Current.AverageValue.String()
			} else if currentMetric.Resource.Current.Value != nil {
				currentCPU = currentMetric.Resource.Current.Value.String()
			}
			break
		}
	}

	if targetCPU == "" && currentCPU == "" {
		return "-"
	}

	if targetCPU == "" {
		targetCPU = "?"
	}

	if currentCPU == "" {
		currentCPU = "?"
	}

	return fmt.Sprintf("cpu: %s/%s", currentCPU, targetCPU)
}

// quantityToMiB converts a memory quantity to MiB (approximate).
func quantityToMiB(q resource.Quantity) float64 {
	// AsApproximateFloat64 returns bytes for memory quantities.
	bytes := q.AsApproximateFloat64()
	return bytes / (1024 * 1024)
}
