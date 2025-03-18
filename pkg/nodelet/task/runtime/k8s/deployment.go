package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeployment(clientset *kubernetes.Clientset, deployment *appsv1.Deployment) {

	_, err := clientset.AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		logs.Error("k8s create deployment template fail")
	}
	logs.Info("Deployment created successfully", "Deployment:", deployment.Name)
}
