## 阶段 1：构建阶段
#FROM golang:1.23 AS builder
#
## 设置工作目录为 /workspace
#WORKDIR /workspace
#
## 设置Go proxy
#ENV GOPROXY=https://goproxy.cn,direct
#
## 拷贝文件
#COPY ../../ /workspace
#
## 下载依赖，缓存依赖层
#RUN go mod download
#
## 设置 Go 环境变量
#ENV GO111MODULE=on \
#    CGO_ENABLED=0 \
#    GOOS=linux \
#    GOARCH=amd64
#
## 构建 nodelet 二进制文件
#RUN make all WHAT=./cmd/nodelet FRAMEWORK_VERBOSE=3
#
#
## 阶段 2：运行阶段
#FROM alpine:3.18
#
## 设置工作目录
#WORKDIR /workspace
#
## 从 builder 阶段中复制构建好的 nodelet 二进制文件
#COPY --from=builder /workspace/_output/local/go/bin/nodelet /workspace/nodelet
#
## 确保二进制文件具有执行权限
#RUN chmod +x /workspace/nodelet
#
## 设置容器启动命令
#CMD ["/workspace/nodelet"]


#方案2：
# 阶段 1：构建阶段
FROM golang:1.23 AS builder

# 设置工作目录为 /workspace
WORKDIR /workspace

# 设置 Go proxy
ENV GOPROXY=https://goproxy.cn,direct

# 1. 从项目根目录复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 2.下载依赖，缓存依赖层
RUN go mod download

# 3. 复制整个项目到容器（注意构建上下文为项目根目录）
COPY . .

# 设置 Go 环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 构建 nodelet 二进制文件
RUN make all WHAT=./cmd/nodelet FRAMEWORK_VERBOSE=3


# 阶段 2：运行阶段（基于 conda 镜像）
FROM continuumio/miniconda3:latest AS conda
#FROM continuumio/miniconda3:23.10.0-0 AS conda
# 设置工作目录
WORKDIR /workspace
RUN conda create -n yolo python=3.10 -y && \
    conda install -n yolo -c pytorch -c conda-forge \
    pytorch torchvision torchaudio cudatoolkit=11.8 \
    ultralytics -y && \
    conda clean --all -y && \
    find /opt/conda/ -type f -name '*.a' -delete && \
    rm -rf /opt/conda/pkgs

# 最终阶段
FROM debian:bullseye-slim
WORKDIR /app
# 从 builder 阶段中复制构建好的 nodelet 二进制文件
COPY --from=builder /workspace/_output/local/go/bin/nodelet .
# 从conda阶段复制环境
COPY --from=conda /opt/conda /opt/conda
# 确保二进制文件具有执行权限
RUN chmod +x /app/nodelet

# 设置环境变量
ENV PATH="/opt/conda/envs/yolo/bin:$PATH" \
    LD_LIBRARY_PATH="/opt/conda/envs/yolo/lib:$LD_LIBRARY_PATH"

# 创建非root用户
RUN useradd -m appuser && chown -R appuser /app
USER appuser

# 验证环境
RUN python -c "import torch; print(torch.__version__)" && \
    python -c "from ultralytics import YOLO; print(YOLO)"

CMD ["./nodelet"]