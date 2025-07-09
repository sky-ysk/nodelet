#!/bin/bash

selected_script="./test/nodelet/reference/workflow1/main.go"
script_dir=$(dirname "$selected_script")
script_file=$(basename "$selected_script")

# 执行信息输出
echo -e "\n\033[34m[$(date +'%T')] 运行程序: ${selected_script}\033[0m"
echo -e "分配权重: \033[33m${workflow_weights[$selected_script]}%\033[0m | 下次执行等待: \033[33m${delay}秒\033[0m"

# 执行程序
if (cd "$script_dir" && go run "$script_file"); then
    echo -e "\033[32m执行成功 ?\033[0m"
else
    echo -e "\033[31m执行失败 ?\033[0m" >&2
fi