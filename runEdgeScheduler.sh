#!/bin/bash

#test deployment work status 312
# 定义相对路径下 apiserver、scheduler 和 nodelet 文件所在的目录
BIN_DIR="./_output/local/go/bin"
APISERVER_PATH="$BIN_DIR/apiserver"
SCHEDULER_PATH="$BIN_DIR/scheduler"
NODELET_PATH="$BIN_DIR/nodelet"
PROXY_PATH="$BIN_DIR/proxy"



# 函数用于检查进程是否存在
check_process_running() {
    local process_name="$1"
    if pgrep -f "$process_name" > /dev/null; then
        return 0
    else
        return 1
    fi
}

## 判断 apiserver 文件是否存在
#if [ -f "$APISERVER_PATH" ]; then
#  if check_process_running "$APISERVER_PATH"; then
#      echo "The apiserver is already running."
#  else
#    # 使用 nohup 将 apiserver 放到后台运行，并将输出重定向到 apiserver_log.log 文件，带上相应参数
#    nohup "$APISERVER_PATH" --etcd-servers=127.0.0.1:2379 > apiserver_log.log 2>&1 &
#  fi
#else
#    echo "The apiserver file at $APISERVER_PATH does not exist."
#    exit 1
#fi
#
#echo "The apiserver output is redirected to apiserver_log.log, and the scheduler output is shown in the foreground."

# 休眠 2 秒
#sleep 2

# 判断 proxy 文件是否存在
#if [ -f "$PROXY_PATH" ]; then
#  if check_process_running "$PROXY_PATH"; then
#      echo "The proxy is already running."
#  else
#    # 使用 nohup 将 proxy 放到后台运行，并将输出重定向到 proxy_log.log 文件，带上相应参数
#    nohup "$PROXY_PATH" > proxy_log.log 2>&1 &
#  fi
#else
#    echo "The proxy file at $PROXY_PATH does not exist."
#    exit 1
#fi
#
#echo "The proxy output is redirected to proxy_log.log."
#
## 休眠 2 秒
#sleep 2

# 判断 scheduler 文件是否存在
if [ -f "$SCHEDULER_PATH" ]; then
  if check_process_running "$SCHEDULER_PATH"; then
      echo "The scheduler is already running. Restarting scheduler ..."
      killall scheduler
  fi
  # 使用 nohup 将 scheduler 放到后台运行，并将输出重定向到 scheduler_log.log 文件
  nohup "$SCHEDULER_PATH" > scheduler_log.log 2>&1 &
  echo "The scheduler output is redirected to scheduler_log.log."
else
    echo "The scheduler file at $SCHEDULER_PATH does not exist."
    exit 1
fi

# 休眠 2 秒
sleep 2

# 判断 nodelet 文件是否存在
if [ -f "$NODELET_PATH" ]; then
  if check_process_running "$NODELET_PATH"; then
      echo "The nodelet is already running. Restarting nodelet ..."
      killall nodelet
  fi
  # 使用 nohup 将 nodelet 放到后台运行，并将输出重定向到 nodelet_log.log 文件
  nohup "$NODELET_PATH" > nodelet_log.log 2>&1 &
  echo "The nodelet output is redirected to nodelet_log.log."
else
    echo "The nodelet file at $NODELET_PATH does not exist."
    exit 1
fi