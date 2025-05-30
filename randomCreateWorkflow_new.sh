#!/bin/bash

# 定义Workflow名称及其权重（百分比）
declare -A workflow_weights=(
    ["./test/nodelet/reference/workflow1/main.go"]=20    # Workflow1占15%
    ["./test/nodelet/reference/workflow6/main.go"]=20  # Workflow6占25%
    ["./test/nodelet/reference/workflow3/main.go"]=20    # Workflow3占30%
    ["./test/nodelet/reference/workflow4/main.go"]=20    # Workflow4占20%
    ["./test/nodelet/reference/workflow5/main.go"]=20    # Workflow5占10%
)

# 固定总轮次（建议设置为100的倍数）
total_rounds=25

# ---------- 生成符合权重的任务序列 ----------
generate_tasks() {
    local tasks=()
    # 计算每个Workflow的精确次数
    for key in "${!workflow_weights[@]}"; do
        count=$(( total_rounds * workflow_weights[$key] / 100 ))
        for ((i=0; i<count; i++)); do
            tasks+=("$key")
        done
    done

    # 处理余数（按权重从高到低补足）
    local remainder=$(( total_rounds - ${#tasks[@]} ))
    if [ $remainder -gt 0 ]; then
        # 按权重从高到低排序
        local sorted_keys=$(for key in "${!workflow_weights[@]}"; do 
            echo "$key ${workflow_weights[$key]}"
        done | sort -k2,2nr | cut -d' ' -f1)

        # 补足余数
        while [ $remainder -gt 0 ]; do
            for key in $sorted_keys; do
                tasks+=("$key")
                ((remainder--))
                if [ $remainder -le 0 ]; then break 2; fi
            done
        done
    fi

    # 打乱顺序（不固定随机种子，保证每次顺序不同）
    shuf -e "${tasks[@]}"
}

# ---------- 执行任务（仅执行一次100轮） ----------
echo -e "\033[34m[INFO] 开始执行任务，总轮次: ${total_rounds}\033[0m"

# 生成任务序列并打乱
mapfile -t tasks < <(generate_tasks)

# 遍历执行任务
for selected_script in "${tasks[@]}"; do
    # 生成随机延迟（40-70秒）
    delay=$(( RANDOM % 31  + 30 ))  # 修正为40-70秒

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

    # 倒计时显示
    for ((i=delay; i>0; i--)); do
        echo -ne "剩余等待: \033[35m${i}s\033[0m\r"
        sleep 1
    done
done

# 脚本结束提示
echo -e "\n\033[34m[INFO] 所有任务执行完成，总轮次: ${total_rounds}\033[0m"


