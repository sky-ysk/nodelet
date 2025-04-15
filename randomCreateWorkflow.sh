#!/bin/bash

# 定义Go程序入口文件路径数组 (注意检查文件名拼写应为 main.go 而非 mian.go)
declare -a scripts=(
    "./test/score/workflow1/main.go"
    "./test/score/workflow6/main.go"
    "./test/score/workflow3/main.go"
    "./test/score/workflow4/main.go"
    "./test/score/workflow5/main.go"
)

while true; do
    # 生成随机延迟 (40-70秒)
    delay=$(( RANDOM % 25 + 30 ))

    # 随机选择程序
    script_index=$(( RANDOM % 5 ))
    selected_script="${scripts[$script_index]}"
    script_dir=$(dirname "$selected_script")
    script_file=$(basename "$selected_script")

    # 执行信息输出
    echo -e "\n\033[34m[$(date +'%T')] 运行程序: ${selected_script}\033[0m"
    echo -e "下次执行等待: \033[33m${delay}秒\033[0m"

    # 执行程序（添加错误处理）
    if (cd "$script_dir" && go run "$script_file"); then
        echo -e "\033[32m执行成功 ✔\033[0m"
    else
        echo -e "\033[31m执行失败 ✘\033[0m" >&2
    fi

    # 倒计时显示
    for ((i=delay; i>0; i--)); do
        echo -ne "剩余等待: \033[35m${i}s\033[0m\r"
        sleep 1
    done
done