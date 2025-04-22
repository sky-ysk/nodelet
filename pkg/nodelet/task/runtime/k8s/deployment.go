package k8s

//func CreateDeployment(clientset *kubernetes.Clientset, deployment *appsv1.Deployment) {
//	if err := EnsureNamespace(clientset, deployment.Namespace); err != nil {
//		logs.Errorf("无法确保命名空间存在: %v", err)
//		return
//	}
//	_, err := clientset.AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
//	if err != nil {
//		logs.Error("k8s create deployment template fail")
//	}
//	logs.Info("Deployment created successfully", "Deployment:", deployment.Name)
//}
