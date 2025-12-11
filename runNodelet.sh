#!/bin/bash
# 定义相对路径下 apiserver、scheduler 和 nodelet 文件所在的目录
BIN_DIR="./_output/local/go/bin"
NODELET_PATH="$BIN_DIR/nodelet"

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

# 判断 nodelet 文件是否存在
if [ -f "$NODELET_PATH" ]; then
  if check_process_running "$NODELET_PATH"; then
      echo "The nodelet is already running. Restarting nodelet ..."
      killall nodelet
      sleep 1
  fi
  # 将 nodelet 放到前台运行
  echo "Starting nodelet in the foreground ..."
  nohup "$NODELET_PATH" --framework-conf ./frameworkConf.yaml &
else
    echo "The nodelet file at $NODELET_PATH does not exist."
    exit 1
fi