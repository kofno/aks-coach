package kube

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListDeployments retrieves a list of Deployments based on the provided scope and namespace context.
func ListDeployments(ctx context.Context, client *kubernetes.Clientset, scope Scope) ([]appsv1.Deployment, error) {
	deployments, err := client.AppsV1().Deployments(scope.NS()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deployments.Items, nil
}
