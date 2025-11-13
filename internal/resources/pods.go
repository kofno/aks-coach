package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// AggregatePodResources calculates total CPU and memory requests and limits for all containers in a Deployment.
// Returns CPU requests and limits in milli-units, and memory requests and limits in MiB.
func AggregatePodResources(d appsv1.Deployment) (float64, float64, float64, float64) {
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

// quantityToMiB converts a resource.Quantity to its approximate size in MiB (Mebibytes).
// Returns the float64 representation of the size in MiB.
func quantityToMiB(q resource.Quantity) float64 {
	// AsApproximateFloat64 returns bytes for memory quantities.
	bytes := q.AsApproximateFloat64()
	return bytes / (1024 * 1024)
}
