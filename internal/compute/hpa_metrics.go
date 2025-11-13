package compute

import (
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
)

// SummarizeCPU generates a summary string of current and target CPU metrics for a HorizontalPodAutoscaler.
// It evaluates metrics from both the spec and status fields to compute utilization or absolute values.
func SummarizeCPU(h *autoscalingv2.HorizontalPodAutoscaler) string {
	var currentCPU string
	var targetCPU string

	for _, specMetric := range h.Spec.Metrics {
		if specMetric.Type == autoscalingv2.ResourceMetricSourceType &&
			specMetric.Resource != nil &&
			specMetric.Resource.Name == corev1.ResourceCPU {

			switch specMetric.Resource.Target.Type {
			case autoscalingv2.UtilizationMetricType:
				if specMetric.Resource.Target.AverageUtilization != nil {
					targetCPU = fmt.Sprintf("%d%%", *specMetric.Resource.Target.AverageUtilization)
				}
			case autoscalingv2.AverageValueMetricType:
				if specMetric.Resource.Target.AverageValue != nil {
					targetCPU = specMetric.Resource.Target.AverageValue.String()
				}
			case autoscalingv2.ValueMetricType:
				if specMetric.Resource.Target.Value != nil {
					targetCPU = specMetric.Resource.Target.Value.String()
				}
			}
			break
		}
	}

	for _, statusMetric := range h.Status.CurrentMetrics {
		if statusMetric.Type == autoscalingv2.ResourceMetricSourceType &&
			statusMetric.Resource != nil &&
			statusMetric.Resource.Name == corev1.ResourceCPU {

			if statusMetric.Resource.Current.AverageUtilization != nil {
				currentCPU = fmt.Sprintf("%d%%", *statusMetric.Resource.Current.AverageUtilization)
			} else if statusMetric.Resource.Current.AverageValue != nil {
				currentCPU = statusMetric.Resource.Current.AverageValue.String()
			} else if statusMetric.Resource.Current.Value != nil {
				currentCPU = statusMetric.Resource.Current.Value.String()
			}
			break
		}
	}

	if currentCPU == "" && targetCPU == "" {
		return "-"
	}
	if currentCPU == "" {
		currentCPU = "?"
	}
	if targetCPU == "" {
		targetCPU = "?"
	}

	return fmt.Sprintf("cpu: %s/%s", currentCPU, targetCPU)
}
