package controller

//func (c *Controller) syncHandler(key string) error {
//	namespace, name, err := cache.SplitMetaNamespaceKey(key)
//	if err != nil {
//		return fmt.Errorf("invalid resource key: %s", key)
//	}
//
//	// 获取 Group 对象
//	group, err := c.groupLister.Groups(namespace).Get(name)
//	if errors.IsNotFound(err) {
//		logs.Infof("Group %s/%s has been deleted", namespace, name)
//		return nil
//	}
//	if err != nil {
//		return err
//	}
//
//	// 获取关联节点状态
//	node, err := c.kubeClient.CoreV1().Nodes().Get(context.TODO(), group.Spec.NodeName, metav1.GetOptions{})
//	if err != nil {
//		return fmt.Errorf("failed to get node %s: %v", group.Spec.NodeName, err)
//	}
//
//	// 检查迁移条件
//	if c.switchCheck.SwitchCondition.CheckCondition(node) {
//		return c.handleMigration(group.DeepCopy())
//	}
//
//	return nil
//}
//
//func (c *Controller) handleMigration(group *corev1.Group) error {
//	// 此处整合原有 groupMigration 逻辑
//	// 保持原有迁移逻辑，但改为基于最新状态操作
//	// 示例：更新 Group 状态为 Migrating
//	patch := []byte(`{"status":{"phase":"Migrating"}}`)
//	_, err := c.groupClient.CoresV1().Groups(group.Namespace).Patch(
//		context.TODO(),
//		group.Name,
//		types.MergePatchType,
//		patch,
//		metav1.PatchOptions{},
//	)
//	if err != nil {
//		return fmt.Errorf("failed to patch group status: %v", err)
//	}
//
//	// 执行具体迁移操作（保持原有逻辑）
//	// ...
//
//	return nil
//}
