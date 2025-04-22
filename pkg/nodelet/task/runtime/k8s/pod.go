package k8s

//func CreatePod(clientset *kubernetes.Clientset, pod *corev1.Pod) {
//	if err := EnsureNamespace(clientset, pod.Namespace); err != nil {
//		logs.Errorf("无法确保命名空间存在: %v", err)
//		return
//	}
//	_, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
//	if err != nil {
//		logs.Error(err, "fail to start pod")
//	}
//	//EM.GetInstance().AddPod(podInfo.Namespace, podInfo.PodName)
//	logs.Info("Pod started successfully", "Pod:", pod.Name)
//	//return CreatePod(groupName, runtime)
//}
//func CreatePodFromYAML(clientset *kubernetes.Clientset, pod *corev1.Pod) error {
//	// 调用原生API创建
//	_, err := clientset.CoreV1().Pods(pod.Namespace).Create(
//		context.TODO(),
//		pod,
//		metav1.CreateOptions{FieldValidation: "Strict"})
//	return err
//}
