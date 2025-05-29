#!/bin/bash

# 默认参数
MAX_LOOPS=10       # 最大执行次数
SLEEP_TIME=60      # 每次执行后的休眠时间（秒）

# 帮助函数
show_help() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -c, --count N      执行N次后退出（默认: 10）"
    echo "  -s, --sleep N      每次执行后休眠N秒（默认: 60秒）"
    echo "  -h, --help         显示此帮助信息"
    exit 0
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case "$1" in
        -c|--count)
            MAX_LOOPS="$2"
            shift 2
            ;;
        -s|--sleep)
            SLEEP_TIME="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            ;;
        *)
            echo "未知参数: $1" >&2
            show_help
            ;;
    esac
done

# 格式化时间显示
format_time() {
    local seconds=$1
    if (( seconds < 60 )); then
        echo "${seconds}秒"
    elif (( seconds < 3600 )); then
        local minutes=$(( seconds / 60 ))
        local secs=$(( seconds % 60 ))
        echo "${minutes}分${secs}秒"
    else
        local hours=$(( seconds / 3600 ))
        local minutes=$(( (seconds % 3600) / 60 ))
        local secs=$(( seconds % 60 ))
        echo "${hours}时${minutes}分${secs}秒"
    fi
}

echo "=== 任务开始执行 ==="
echo "执行间隔: $(format_time $SLEEP_TIME)"
echo "总执行次数: $MAX_LOOPS"
echo "====================="

# 初始化计数器
COUNTER=0

# 主循环
while (( COUNTER < MAX_LOOPS )); do
    # 增加计数器
    ((COUNTER++))

    # 显示当前执行轮次
    echo -e "\n[$(date '+%Y-%m-%d %H:%M:%S')] 开始执行第 ${COUNTER}/${MAX_LOOPS} 轮..."

    # 执行程序
    go run ./test/scene3/createTaskForScene3.go

    # 检查是否达到最大循环次数
    if (( COUNTER >= MAX_LOOPS )); then
        echo -e "\n[$(date '+%Y-%m-%d %H:%M:%S')] 已完成所有 $MAX_LOOPS 次执行，任务结束。"
        exit 0
    fi

    # 显示休眠倒计时
    echo -e "\n[$(date '+%Y-%m-%d %H:%M:%S')] 开始休眠 $(format_time $SLEEP_TIME)，按 Ctrl+C 终止..."
    for (( i=SLEEP_TIME; i>0; i-- )); do
        echo -ne "\r剩余休眠时间: $(format_time $i)   "
        sleep 1
    done
    echo -e "\r剩余休眠时间: 0秒      "  # 清除倒计时显示
done