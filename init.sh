#!/bin/bash

# 定义函数，用于打印用法信息
usage() {
    echo "Usage: $0 -i {u|d}"
    exit 1
}

# 检查参数数量
if [ $# -ne 2 ]; then
    usage
fi

# 解析参数
while getopts ":i:" opt; do
    case $opt in
        i)
            if [ "$OPTARG" == "u" ]; then
                echo "Executing initUp.go..."
                go run ./test/scene3/initUp.go
            elif [ "$OPTARG" == "d" ]; then
                echo "Executing initDown.go..."
                go run ./test/scene3/initDown.go
            else
                echo "Invalid argument for -i: $OPTARG"
                usage
            fi
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            usage
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            usage
            ;;
    esac
done