package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeployment(clientset *kubernetes.Clientset, deployment *entity.Deployment) error {
	template := deployment.CreateDeploymentTemplate()
	_, err := clientset.AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), &template, metav1.CreateOptions{})
	if err != nil {
		logs.Error("k8s create deployment template fail")
		return err
	}
	logs.Info("Deployment created successfully", "Deployment:", deployment.DeploymentName)
	return nil
}
