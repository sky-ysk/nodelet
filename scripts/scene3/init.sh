#!/bin/bash
# 这个脚本用于流程1的init操作 有两个参数 -i 控制up/down -r 指定哪台机器人

# 定义函数，用于打印用法信息
usage() {
    echo "Usage: $0 -i {u|d} -r {1|2}"
    exit 1
}

# 检查参数数量
if [ $# -lt 4 ]; then
    usage
fi

# 解析参数
init_script=""
while getopts ":i:r:" opt; do
    case $opt in
        i)
            if [ "$OPTARG" == "u" ]; then
                init_script="initUp"
            elif [ "$OPTARG" == "d" ]; then
                init_script="initDown"
            else
                echo "Invalid argument for -i: $OPTARG"
                usage
            fi
            ;;
        r)
            if [ "$OPTARG" == "1" ]; then
                init_script+="1"
            elif [ "$OPTARG" == "2" ]; then
                init_script+="2"
            else
                echo "Invalid argument for -r: $OPTARG"
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

# 检查是否提供了所有必需的参数
if [ -z "$init_script" ]; then
    usage
fi

# 执行相应的脚本
echo "Executing ${init_script}.go..."
go run ../../test/scene3/${init_script}.go