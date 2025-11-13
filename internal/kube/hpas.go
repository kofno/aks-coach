package kube

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListHPAs retrieves a map of HorizontalPodAutoscalers filtered by the provided scope from the Kubernetes cluster.
// ctx is the execution context for the operation.
// clientset is the Kubernetes clientset used to interact with the cluster.
// scope defines the namespace constraints for the retrieval (all namespaces or a specific namespace).
// Returns a map where the key is a combination of namespace and deployment name, and the value is the HPA object.
// Returns an error if the operation fails while listing HorizontalPodAutoscalers.
func ListHPAs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	scope Scope,
) (map[string]*autoscalingv2.HorizontalPodAutoscaler, error) {

	hpas, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(scope.NS()).List(ctx, metav1.ListOptions{})
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
