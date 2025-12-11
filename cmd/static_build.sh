#!/bin/bash

set -e  # 遇到任何错误立即退出

# 输出目录（相对于当前脚本所在目录）
OUTPUT_DIR="../_output/local/go/bin"

# 确保输出目录存在
mkdir -p "$OUTPUT_DIR"

# 定义要构建的二进制列表：格式为 "输出名:源文件"
BINARIES=(
    "nodelet:./nodelet/nodelet.go"
    "scheduler:./scheduler/scheduler.go"
    "proxy:./proxy/proxy.go"
    "apiserver:./apiserver/apiserver.go"
)

echo "开始静态编译 Go 二进制文件（CGO_ENABLED=0）..."

for bin in "${BINARIES[@]}"; do
    IFS=':' read -r name src <<< "$bin"
    output_path="$OUTPUT_DIR/$name"
    
    echo "正在编译 $name <- $src ..."
    
    CGO_ENABLED=0 GOOS=linux go build \
        -a \
        -ldflags '-extldflags "-static"' \
        -o "$output_path" \
        "$src"
        
    # 添加可执行权限
    chmod +x "$output_path"
    echo "✅ 已赋予可执行权限: $output_path"
done

echo "🎉 所有二进制已成功编译并设置为可执行！"