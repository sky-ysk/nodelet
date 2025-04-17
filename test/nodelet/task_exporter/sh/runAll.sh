#!/bin/bash
# 定义相对路径下 apiserver、scheduler 和 nodelet 文件所在的目录
BIN_DIR="./_output/local/go/bin"
APISERVER_PATH="$BIN_DIR/apiserver"
SCHEDULER_PATH="$BIN_DIR/scheduler"
NODELET_PATH="$BIN_DIR/nodelet"
PROXY_PATH="$BIN_DIR/proxy"

# 删除上一次运行留下来的日志文件
rm -f nohup.out

# 函数用于检查进程是否存在
check_process_running() {
    local process_name="$1"
    if pgrep -f "$process_name" > /dev/null; then
        return 0
    else
        return 1
    fi
}

# 判断 apiserver 文件是否存在
if [ -f "$APISERVER_PATH" ]; then
  if check_process_running "$APISERVER_PATH"; then
      echo "The apiserver is already running."
      killall apiserver
      sleep 1
  fi
    # 使用 nohup 将 apiserver 放到后台运行，并将输出重定向到 apiserver_log.log 文件，带上相应参数
    nohup "$APISERVER_PATH" --etcd-servers=127.0.0.1:2379 > apiserver_log.log 2>&1 &   #如果文件已存在，内容会被覆盖  > scheduler_log.log:表示将标准输出（stdout）重定向到 apiserver_log.log 文件    2>&1：将标准错误（stderr）重定向到标准输出（stdout）  &：将进程放入后台运行，释放当前终端供其他操作使用
    echo "Started apiserver and redirected output to apiserver_log.log."
else
    echo "The apiserver file at $APISERVER_PATH does not exist."
    exit 1
fi
# 休眠 2 秒
sleep 2

# 判断 scheduler 文件是否存在
if [ -f "$SCHEDULER_PATH" ]; then
  if check_process_running "$SCHEDULER_PATH"; then
      echo "The scheduler is already running. Restarting scheduler ..."
      killall scheduler
      sleep 1
  fi
  # 使用 nohup 将 scheduler 放到后台运行，并将输出重定向到 scheduler_log.log 文件
  nohup "$SCHEDULER_PATH" > scheduler_log.log 2>&1 &
  echo "Started scheduler and redirected output to scheduler_log.log."
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
      sleep 1
  fi
  # 将 nodelet 放到前台运行
  echo "Starting nodelet in the foreground ..."
  nohup "$NODELET_PATH" &
else
    echo "The nodelet file at $NODELET_PATH does not exist."
    exit 1
fi

# 判断 proxy 文件是否存在
if [ -f "$PROXY_PATH" ]; then
  if check_process_running "$PROXY_PATH"; then
      echo "The proxy is already running. Restarting proxy ..."
      killall proxy
      sleep 1
  fi
  # 将 proxy 放到后台运行
  echo "Starting proxy and redirected output to proxy_log.log."
  nohup "$PROXY_PATH" > proxy_log.log 2>&1 &
else
    echo "The proxy file at $PROXY_PATH does not exist."
    exit 1
fi
