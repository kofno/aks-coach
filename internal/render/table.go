package render

import (
	"aks-coach/internal/compute"
	"aks-coach/internal/resources"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

// PrintTable displays a formatted table of Deployment details, including HPA related data, resource requests, and limits.
func PrintTable(
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

		perPodCPUReq, perPodCPULimit, perPodMemReq, perPodMemLimit := resources.AggregatePodResources(d)

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
			hpaTarget = compute.SummarizeCPU(h)
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
