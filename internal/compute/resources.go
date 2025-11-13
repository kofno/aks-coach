package compute

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Row struct {
	Namespace string
	Name      string
	Replicas  int32

	CPUReqMilli   float64
	CPULimitMilli float64
	MemReqMi      float64
	MemLimitMi    float64

	HPAMin    string
	HPAMax    string
	HPATarget string
}

// BuildRows generates a slice of Row objects based on the provided deployments and HPA map.
// It calculates resource metrics and updates HPA-related fields for each deployment.
func BuildRows(deps []appsv1.Deployment, hpaMap map[string]*autoscalingv2.HorizontalPodAutoscaler) []Row {
	rows := make([]Row, 0, len(deps))
	for _, d := range deps {
		replicas := int32(1)
		if d.Spec.Replicas != nil {
			replicas = *d.Spec.Replicas
		}

		cr, cl, mr, ml := aggregate(d)

		r := Row{
			Namespace:     d.Namespace,
			Name:          d.Name,
			Replicas:      replicas,
			CPUReqMilli:   cr * float64(replicas),
			CPULimitMilli: cl * float64(replicas),
			MemReqMi:      mr * float64(replicas),
			MemLimitMi:    ml * float64(replicas),
			HPAMin:        "-",
			HPAMax:        "-",
			HPATarget:     "-",
		}

		if hpa, ok := hpaMap[d.Namespace+"/"+d.Name]; ok && hpa != nil {
			if hpa.Spec.MinReplicas != nil {
				r.HPAMin = itoa(*hpa.Spec.MinReplicas)
			} else {
				r.HPAMin = ""
			}
			r.HPAMax = itoa(hpa.Spec.MaxReplicas)
			r.HPATarget = SummarizeCPU(hpa)
		}

		rows = append(rows, r)
	}
	return rows
}

func aggregate(d appsv1.Deployment) (cpuReqMilli, cpuLimitMilli, memReqMi, memLimitMi float64) {
	for _, c := range d.Spec.Template.Spec.Containers {
		if q, ok := c.Resources.Requests["cpu"]; ok {
			cpuReqMilli += q.AsApproximateFloat64() * 1000
		}
		if q, ok := c.Resources.Limits["cpu"]; ok {
			cpuLimitMilli += q.AsApproximateFloat64() * 1000
		}
		if q, ok := c.Resources.Requests["memory"]; ok {
			memReqMi += toMi(q)
		}
		if q, ok := c.Resources.Limits["memory"]; ok {
			memLimitMi += toMi(q)
		}
	}
	return
}

func toMi(q resource.Quantity) float64 { return q.AsApproximateFloat64() / (1024 * 1024) }
func itoa(i int32) string              { return fmt.Sprintf("%d", i) }
