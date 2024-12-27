package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity"
	EM "hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity/entity_manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePod(clientset *kubernetes.Clientset, podInfo *entity.Pod) error {
	pod, err := podInfo.CreatePodTemplate()
	if err != nil {
		return err
	}
	_, err2 := clientset.CoreV1().Pods(podInfo.Namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
	if err2 != nil {
		logs.Error(err, "fail to start pod")
		return err
	}
	EM.GetInstance().AddPod(podInfo.Namespace, podInfo.PodName)
	logs.Info("Pod started successfully", "Pod:", podInfo.PodName)
	//return CreatePod(groupName, runtime)
	return nil
}
