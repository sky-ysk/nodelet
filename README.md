# Adaptive Scheduling Framework
自适应调度框架

## 如何使用调度框架
TODO: 教程文档

### 日志模块
[日志模块文档](./docs/logs.md)

## 如何部署调度框架
TODO：部署文档

### 编译模块
examples
``` shell
make all WHAT=./cmd/proxy FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/apiserver FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/registry FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/nodelet FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/proxy FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/scheduler FRAMEWORK_VERBOSE=3
```
二进制文件在`_output/xxx/bin`目录下

### 部署API-Server
本地安etcd，启动etcd.service
然后执行启动命令
`apiserver --etcd-servers=127.0.0.1:2379`

### 部署client-go
参考pkg/client-go/examples中的用法
参考pkg/proxy/handlers/中的用法

### 部署Proxy
如果要使用Swagger-UI,则将`third_party/swagger-ui`目录拷贝到本机`/tmp`目录下
`cp -r ./third_party/swagger-ui /tmp`

### 部署Nodelet

### 部署Resourcelet

### 部署Scheduler
## 技术支持



## Roadmap



## Licences

