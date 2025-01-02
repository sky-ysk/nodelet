#!/bin/bash

# 定义相对路径下apiserver和scheduler文件所在的目录
BIN_DIR="./_output/local/go/bin"
APISERVER_PATH="$BIN_DIR/apiserver"
SCHEDULER_PATH="$BIN_DIR/scheduler"
NODELET_PATH="$BIN_DIR/nodelet"

# 函数用于检查进程是否存在
check_process_running() {
    local process_name="$1"
    if pgrep -f "$process_name" > /dev/null; then
        return 0
    else
        return 1
    fi
}



# 判断apiserver文件是否存在
if [ -f "$APISERVER_PATH" ]; then
  if check_process_running "$APISERVER_PATH"; then
      echo "The apiserver is already running."
  else
    # 使用nohup将apiserver放到后台运行，并将输出重定向到apiserver_log.log文件，带上相应参数
    nohup "$APISERVER_PATH" --etcd-servers=127.0.0.1:2379 > apiserver_log.log 2>&1 &
  fi
else
    echo "The apiserver file at $APISERVER_PATH does not exist."
    exit 1
fi

# 休眠2秒
sleep 2



# 判断scheduler文件是否存在
if [ -f "$SCHEDULER_PATH" ]; then
  if check_process_running "$SCHEDULER_PATH"; then
      echo "The scheduler is already running. restart scheduler ..."
      killall scheduler
  fi
    "$SCHEDULER_PATH"
else
    echo "The scheduler file at $SCHEDULER_PATH does not exist."
    exit 1
fi

echo "The apiserver output is redirected to apiserver_log.log, and the scheduler output is shown in the foreground."

# 休眠2秒
sleep 2

# 判断 scheduler 文件是否存在
if [ -f "NODELET_PATH" ]; then
  if check_process_running "NODELET_PATH"; then
      echo "The scheduler is already running. restart scheduler ..."
      killall scheduler
  fi
    "NODELET_PATH"
else
    echo "The scheduler file at NODELET_PATH does not exist."
    exit 1
fi
