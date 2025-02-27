# 阶段 1：构建阶段
FROM golang:1.23 AS builder

# 设置工作目录为 /workspace
WORKDIR /workspace

# 设置Go proxy
ENV GOPROXY=https://goproxy.cn,direct

# 拷贝文件
COPY ../../ /workspace

# 下载依赖，缓存依赖层
RUN go mod download

# 设置 Go 环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 构建 nodelet 二进制文件
RUN make all WHAT=./cmd/scheduler FRAMEWORK_VERBOSE=3


# 阶段 2：运行阶段
FROM alpine:3.18

# 设置工作目录
WORKDIR /workspace

# 从 builder 阶段中复制构建好的 nodelet 二进制文件
COPY --from=builder /workspace/_output/local/go/bin/scheduler /workspace/scheduler

# 确保二进制文件具有执行权限
RUN chmod +x /workspace/scheduler

# 设置容器启动命令
CMD ["/workspace/scheduler"]