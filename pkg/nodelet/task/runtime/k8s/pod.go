package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePod(clientset *kubernetes.Clientset, pod *corev1.Pod) {

	_, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		logs.Error(err, "fail to start pod")
	}
	//EM.GetInstance().AddPod(podInfo.Namespace, podInfo.PodName)
	logs.Info("Pod started successfully", "Pod:", pod.Name)
	//return CreatePod(groupName, runtime)
}
