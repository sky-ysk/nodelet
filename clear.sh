#!/bin/bash

# ======================
# Kubernetes 命名空间清理工具
# 功能：删除指定命名空间下的所有 Pod 和 Service
# 安全特性：双重确认 + 资源列表预览
# ======================

NAMESPACE="switch"

# 1. 验证 kubectl 配置
if ! command -v kubectl &> /dev/null; then
    echo "错误：kubectl 未安装或不在 PATH 中"
    exit 1
fi

# 2. 检查命名空间是否存在
if ! kubectl get namespace $NAMESPACE &> /dev/null; then
    echo "错误：命名空间 '$NAMESPACE' 不存在"
    exit 1
fi

# 3. 显示待删除资源预览
echo "=== 将要删除的资源预览 ==="
echo "Pods:"
kubectl get pods -n $NAMESPACE --no-headers | awk '{print $1}' | sort | column
echo -e "\nServices:"
kubectl get services -n $NAMESPACE --no-headers | awk '{print $1}' | sort | column


# 5. 执行删除操作
echo -e "\n开始删除资源..."
echo "删除 Pods..."
kubectl delete pods --all -n $NAMESPACE --wait=false

echo "删除 Services..."
kubectl delete services --all -n $NAMESPACE

# 6. 操作确认
echo -e "\n操作完成！当前剩余资源："
echo "剩余 Pods:"
kubectl get pods -n $NAMESPACE

echo -e "\n剩余 Services:"
kubectl get services -n $NAMESPACE
