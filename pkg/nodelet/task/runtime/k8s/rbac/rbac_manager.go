package rbac

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"time"
)

type RBACManager struct {
	clientset     *kubernetes.Clientset
	globalRole    string
	saName        string
	saNamespace   string
	maxRetries    int
	retryInterval time.Duration
}

func NewRBACManager(clientset *kubernetes.Clientset, opts ...RBACOption) *RBACManager {
	mrg := &RBACManager{
		clientset:     clientset,
		globalRole:    "deploy-system-global", //全局权限规则
		saName:        "deploy-system-sa",
		saNamespace:   "deploy-system",
		maxRetries:    3,
		retryInterval: 2 * time.Second,
	}
	for _, opt := range opts {
		opt(mrg)
	}
	return mrg
}

type RBACOption func(*RBACManager)

func WithRetryPolicy(maxRetries int, interval time.Duration) RBACOption {
	return func(m *RBACManager) {
		m.maxRetries = maxRetries
		m.retryInterval = interval
	}
}

// 环境检测逻辑
func IsRunningInPod() bool {
	// 检查Kubernetes服务账号文件
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	return err == nil
}

// 增强版RBAC管理
func (rm *RBACManager) EnsureAccessFor(namespace string) error {
	if IsRunningInPod() {
		return rm.EnsureNamespaceAccess(namespace)
	}
	// 宿主机环境不需要额外配置
	return nil
}

// 添加ServiceAccount管理
func (rm *RBACManager) ensureServiceAccount() error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rm.saName,
			Namespace: rm.saNamespace,
		},
	}

	// 创建或更新ServiceAccount
	_, err := rm.clientset.CoreV1().ServiceAccounts(rm.saNamespace).Update(
		context.TODO(), sa, metav1.UpdateOptions{},
	)
	if err == nil {
		return nil
	}

	if kerrors.IsNotFound(err) {
		_, createErr := rm.clientset.CoreV1().ServiceAccounts(rm.saNamespace).Create(
			context.TODO(), sa, metav1.CreateOptions{},
		)
		return createErr
	}
	return err
}

// 创建ClusterRole
func (rm *RBACManager) createClusterRole() error {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: rm.globalRole,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"", "apps"},
				Resources: []string{"services", "pods", "deployments"},
				Verbs:     []string{"get", "list", "watch", "create"},
			},
			// 新增 namespaces 权限
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "create"},
			},
		},
	}

	// 尝试更新，不存在则创建
	_, err := rm.clientset.RbacV1().ClusterRoles().Update(
		context.TODO(), cr, metav1.UpdateOptions{},
	)
	if err == nil {
		return nil
	}

	if kerrors.IsNotFound(err) {
		_, createErr := rm.clientset.RbacV1().ClusterRoles().Create(
			context.TODO(), cr, metav1.CreateOptions{},
		)
		return createErr
	}
	return err
}

// 带重试的命名空间创建
func (rm *RBACManager) createNamespaceWithRetry(targetNs string) error {
	for i := 0; i < rm.maxRetries; i++ {
		// 检查是否已存在
		_, err := rm.clientset.CoreV1().Namespaces().Get(
			context.TODO(), targetNs, metav1.GetOptions{},
		)
		if err == nil {
			return nil
		}

		// 如果不存在则创建
		if kerrors.IsNotFound(err) {
			_, createErr := rm.clientset.CoreV1().Namespaces().Create(
				context.TODO(),
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: targetNs,
						Labels: map[string]string{
							"managed-by": "deploy-system",
						},
					},
				},
				metav1.CreateOptions{},
			)
			if createErr == nil || kerrors.IsAlreadyExists(createErr) {
				return nil
			}
		}
		// 等待重试间隔
		time.Sleep(rm.retryInterval)
	}
	return fmt.Errorf("命名空间创建重试%d次后失败", rm.maxRetries)
}
func (rm *RBACManager) createRoleBindingWithRetry(targetNs string) error {
	bindingName := fmt.Sprintf("deploy-access-%s", targetNs)
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: targetNs,
			Labels: map[string]string{
				"managed-by": "deploy-system",
			},
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      rm.saName,
			Namespace: rm.saNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     rm.globalRole,
		},
	}

	for i := 0; i < rm.maxRetries; i++ {
		_, err := rm.clientset.RbacV1().RoleBindings(targetNs).Update(
			context.TODO(), rb, metav1.UpdateOptions{},
		)
		if err == nil {
			return nil
		}

		if kerrors.IsNotFound(err) {
			_, createErr := rm.clientset.RbacV1().RoleBindings(targetNs).Create(
				context.TODO(), rb, metav1.CreateOptions{},
			)
			if createErr == nil || kerrors.IsAlreadyExists(createErr) {
				return nil
			}
		}

		time.Sleep(rm.retryInterval)
	}
	return fmt.Errorf("RoleBinding创建重试%d次后失败", rm.maxRetries)
}
func (rm *RBACManager) EnsureNamespaceAccess(targetNs string) error {
	// 1. 确保基础资源存在
	if err := rm.initializeGlobalResources(); err != nil {
		return err
	}
	// 2、创建RoleBinding，注意：必须首先创建命名空间，才能在该命名空间中创建RoleBinding
	return rm.createRoleBindingWithRetry(targetNs)
}

// 新增初始化方法
func (rm *RBACManager) initializeGlobalResources() error {
	if err := rm.createClusterRole(); err != nil {
		return fmt.Errorf("ClusterRole初始化失败: %v", err)
	}

	if err := rm.ensureServiceAccount(); err != nil {
		return fmt.Errorf("ServiceAccount初始化失败: %v", err)
	}

	// 确保系统命名空间存在
	if err := rm.createNamespaceWithRetry(rm.saNamespace); err != nil {
		return fmt.Errorf("系统命名空间初始化失败: %v", err)
	}
	return nil
}

//func SetupRBAC(clientset *kubernetes.Clientset, namespace string) error {
//	// 创建ServiceAccount
//	sa := &corev1.ServiceAccount{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      ServiceAccountName,
//			Namespace: namespace,
//		},
//	}
//	if _, err := clientset.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), sa, metav1.CreateOptions{}); err != nil {
//		return fmt.Errorf("failed to create ServiceAccount: %v", err)
//	}
//
//	// 创建ClusterRole
//	cr := &rbacv1.ClusterRole{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: ClusterRoleName,
//		},
//		Rules: []rbacv1.PolicyRule{
//			{
//				APIGroups: []string{"", "apps"}, // ""：表示核心 API 组,"apps"：表示 apps API 组（如 Deployment）。
//				Resources: []string{"services", "pods", "deployments"},
//				Verbs:     []string{"get", "list", "watch"},
//			},
//		},
//	}
//	if _, err := clientset.RbacV1().ClusterRoles().Create(context.TODO(), cr, metav1.CreateOptions{}); err != nil {
//		return fmt.Errorf("failed to create ClusterRole: %v", err)
//	}
//
//	// 创建RoleBinding
//	rb := &rbacv1.RoleBinding{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      RoleBindingName,
//			Namespace: namespace,
//		},
//		Subjects: []rbacv1.Subject{
//			{
//				Kind:      "ServiceAccount",
//				Name:      ServiceAccountName,
//				Namespace: namespace,
//			},
//		},
//		RoleRef: rbacv1.RoleRef{
//			APIGroup: "rbac.authorization.k8s.io",
//			Kind:     "ClusterRole",
//			Name:     ClusterRoleName,
//		},
//	}
//	if _, err := clientset.RbacV1().RoleBindings(namespace).Create(context.TODO(), rb, metav1.CreateOptions{}); err != nil {
//		return fmt.Errorf("failed to create RoleBinding: %v", err)
//	}
//	return nil
//}
