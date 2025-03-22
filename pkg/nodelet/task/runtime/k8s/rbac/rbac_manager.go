package rbac

//import (
//	"context"
//	"fmt"
//	corev1 "k8s.io/api/core/v1"
//	rbacv1 "k8s.io/api/rbac/v1"
//	kerrors "k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/client-go/kubernetes"
//	"time"
//)
//
//const (
//	SystemNamespace = "mynodelet"
//	ServiceAccount  = "deploy-system-sa"
//	ClusterRoleName = "deploy-system-global"
//)
//
//type RBACManager struct {
//	clientset     *kubernetes.Clientset
//	maxRetries    int
//	retryInterval time.Duration
//}
//
//func NewRBACManager(clientset *kubernetes.Clientset) *RBACManager {
//	return &RBACManager{
//		clientset:     clientset,
//		maxRetries:    5,
//		retryInterval: 3 * time.Second,
//	}
//}
//
//// 核心方法：初始化全局权限
//func (rm *RBACManager) Initialize() error {
//	if err := rm.createNamespace(SystemNamespace); err != nil {
//		return fmt.Errorf("创建系统命名空间失败: %v", err)
//	}
//
//	if err := rm.createServiceAccount(); err != nil {
//		return fmt.Errorf("创建ServiceAccount失败: %v", err)
//	}
//
//	if err := rm.createGlobalClusterRole(); err != nil {
//		return fmt.Errorf("创建ClusterRole失败: %v", err)
//	}
//
//	if err := rm.createClusterRoleBinding(); err != nil {
//		return fmt.Errorf("创建ClusterRoleBinding失败: %v", err)
//	}
//
//	return nil
//}
//
//// 创建命名空间（带重试逻辑）
//func (rm *RBACManager) createNamespace(name string) error {
//	for i := 0; i < rm.maxRetries; i++ {
//		_, err := rm.clientset.CoreV1().Namespaces().Get(
//			context.TODO(), name, metav1.GetOptions{},
//		)
//		if err == nil {
//			return nil
//		}
//
//		if kerrors.IsNotFound(err) {
//			ns := &corev1.Namespace{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: name,
//					Labels: map[string]string{
//						"managed-by": "deploy-system",
//					},
//				},
//			}
//			_, createErr := rm.clientset.CoreV1().Namespaces().Create(
//				context.TODO(), ns, metav1.CreateOptions{},
//			)
//			if createErr == nil || kerrors.IsAlreadyExists(createErr) {
//				return nil
//			}
//		}
//
//		time.Sleep(rm.retryInterval)
//	}
//	return fmt.Errorf("命名空间 %s 创建失败", name)
//}
//
//// 创建ServiceAccount
//func (rm *RBACManager) createServiceAccount() error {
//	sa := &corev1.ServiceAccount{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      ServiceAccount,
//			Namespace: SystemNamespace,
//		},
//	}
//
//	_, err := rm.clientset.CoreV1().ServiceAccounts(SystemNamespace).Update(
//		context.TODO(), sa, metav1.UpdateOptions{},
//	)
//	if err == nil {
//		return nil
//	}
//
//	if kerrors.IsNotFound(err) {
//		_, err = rm.clientset.CoreV1().ServiceAccounts(SystemNamespace).Create(
//			context.TODO(), sa, metav1.CreateOptions{},
//		)
//	}
//	return err
//}
//
//// 全局ClusterRole定义（支持跨命名空间操作）
//func (rm *RBACManager) createGlobalClusterRole() error {
//	cr := &rbacv1.ClusterRole{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: ClusterRoleName,
//		},
//		Rules: []rbacv1.PolicyRule{
//			// 核心资源权限
//			{
//				APIGroups: []string{""},
//				Resources: []string{"pods", "services", "namespaces"},
//				Verbs:     []string{"*"},
//			},
//			// Deployment权限
//			{
//				APIGroups: []string{"apps"},
//				Resources: []string{"deployments"},
//				Verbs:     []string{"*"},
//			},
//			// 查看节点信息
//			{
//				APIGroups: []string{""},
//				Resources: []string{"nodes"},
//				Verbs:     []string{"get", "list", "watch"},
//			},
//		},
//	}
//
//	_, err := rm.clientset.RbacV1().ClusterRoles().Update(
//		context.TODO(), cr, metav1.UpdateOptions{},
//	)
//	if err == nil {
//		return nil
//	}
//
//	if kerrors.IsNotFound(err) {
//		_, err = rm.clientset.RbacV1().ClusterRoles().Create(
//			context.TODO(), cr, metav1.CreateOptions{},
//		)
//	}
//	return err
//}
//
//// 绑定到集群范围
//func (rm *RBACManager) createClusterRoleBinding() error {
//	binding := &rbacv1.ClusterRoleBinding{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "deploy-system-binding",
//		},
//		Subjects: []rbacv1.Subject{
//			{
//				Kind:      "ServiceAccount",
//				Name:      ServiceAccount,
//				Namespace: SystemNamespace,
//			},
//		},
//		RoleRef: rbacv1.RoleRef{
//			APIGroup: "rbac.authorization.k8s.io",
//			Kind:     "ClusterRole",
//			Name:     ClusterRoleName,
//		},
//	}
//
//	_, err := rm.clientset.RbacV1().ClusterRoleBindings().Update(
//		context.TODO(), binding, metav1.UpdateOptions{},
//	)
//	if err == nil {
//		return nil
//	}
//
//	if kerrors.IsNotFound(err) {
//		_, err = rm.clientset.RbacV1().ClusterRoleBindings().Create(
//			context.TODO(), binding, metav1.CreateOptions{},
//		)
//	}
//	return err
//}
//
//// 检查命名空间是否存在
//func (rm *RBACManager) CheckOrCreateNamespace(targetNs string) error {
//	_, err := rm.clientset.CoreV1().Namespaces().Get(
//		context.TODO(), targetNs, metav1.GetOptions{},
//	)
//	if err == nil {
//		return nil
//	}
//
//	if kerrors.IsNotFound(err) {
//		return rm.createNamespace(targetNs)
//	}
//	return err
//}
