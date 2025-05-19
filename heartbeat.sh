#!/bin/bash

# 定义目标 URL
URL1="172.130.0.59:8080/api/ability-heartbeat"
URL1="172.130.0.61:8080/api/ability-heartbeat"

# 使用 curl 发起请求
echo "正在请求 $URL1..."
curl -s "$URL1"
echo "正在请求 $URL2..."
curl -s "$URL2"

# 检查 curl 命令的退出状态
if [ $? -eq 0 ]; then
  echo "请求成功！"
else
  echo "请求失败！"
fi