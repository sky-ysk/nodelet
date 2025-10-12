#!/bin/bash

# ======================
# Kubernetes 命名空间清理工具
# 功能：删除指定命名空间下的所有 Pod 和 Service
# 安全特性：双重确认 + 资源列表预览
# ======================

NAMESPACE="switch"
CLEAR_SCRIPT_PATH="$HOME/workspace/etcd-v3.5.17-linux-amd64/clear.sh"
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


# 4. 执行删除操作
echo -e "\n开始删除资源..."
echo "删除 Pods..."
kubectl delete pods --all -n $NAMESPACE --wait=false

echo "删除 Services..."
kubectl delete services --all -n $NAMESPACE

# 5. 操作确认
echo -e "\n操作完成！当前剩余资源："
echo "剩余 Pods:"
kubectl get pods -n $NAMESPACE

echo -e "\n剩余 Services:"
kubectl get services -n $NAMESPACE
#删除etcd数据
echo -e "\n执行额外的清理脚本: $CLEAR_SCRIPT_PATH"
if [[ -f $CLEAR_SCRIPT_PATH && -x $CLEAR_SCRIPT_PATH ]]; then
    echo "找到可执行的清理脚本，正在执行..."
    bash $CLEAR_SCRIPT_PATH
    echo "清理脚本执行完成"
else
    echo "警告：清理脚本不存在或不可执行"
    echo "请确保脚本位于: $CLEAR_SCRIPT_PATH"
    echo "并具有执行权限 (chmod +x $CLEAR_SCRIPT_PATH)"
fi