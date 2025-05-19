#!/bin/bash

# 定义目标 URL
URL1="http://172.130.0.59:8080/api/ability-heartbeat"
URL2="http://172.130.0.61:8080/api/ability-heartbeat"

# 当前时间
CURRENT_TIME=$(date +"%Y-%m-%d %H:%M:%S")

# 函数：检查指定 URL 的心跳状态
check_heartbeat() {
  local url=$1
  local response=$(curl -s "$url")
  local status_code=$(curl -o /dev/null -s -w "%{http_code}\n" "$url")

  if [ "$status_code" -eq 200 ]; then
    echo "[$CURRENT_TIME] 请求 $url 成功，HTTP 状态码: $status_code"
    echo "响应内容："
    echo "$response"
  else
    echo "[$CURRENT_TIME] 请求 $url 失败，HTTP 状态码: $status_code"
    echo "响应内容："
    echo "$response"
  fi
  echo "----------------------------------------"
}

# 检查两个 URL 的心跳状态
check_heartbeat "$URL1"
check_heartbeat "$URL2"