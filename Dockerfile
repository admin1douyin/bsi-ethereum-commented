# -----------------------------------------------------------------------------------
# Dockerfile 用于构建 go-ethereum (Geth) 的 Docker 镜像。
# 这个构建过程分为两个阶段：
# 1. 构建阶段 (builder)：在一个包含 Go 编译环境的容器中编译源码。
# 2. 运行阶段 (final)：将编译好的二进制文件复制到一个轻量级的 Alpine Linux 容器中。
# 这样可以确保最终的镜像非常小，只包含运行 Geth 所需的最小环境。
# -----------------------------------------------------------------------------------

# 定义构建参数，这些参数可以在 'docker build' 时通过 '--build-arg' 传入
# 用于为最终镜像添加元数据标签（如 Git 提交哈希、版本号、构建编号）
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# ==============================================================================
# 第一阶段：构建阶段 (Builder Stage)
# 使用官方的 Go 语言 Alpine 镜像作为基础镜像，版本为 1.24
# AS builder 给这个阶段起个别名，方便后面引用
# ==============================================================================
FROM golang:1.24-alpine AS builder

# 安装构建 Geth 所需的系统依赖：
# - gcc, musl-dev: 用于编译包含 C 代码的部分 (CGO)
# - linux-headers: 某些底层系统调用可能需要
# - git: 用于获取依赖或版本信息
# --no-cache 表示不缓存安装包索引，减小镜像体积
RUN apk add --no-cache gcc musl-dev linux-headers git

# 复制 go.mod 和 go.sum 文件到容器的工作目录 /go-ethereum/
# 这是一个 Docker 最佳实践：先复制依赖描述文件，再下载依赖。
# 只要这两个文件没变，Docker 就会使用缓存的依赖层，大大加快后续构建速度。
COPY go.mod /go-ethereum/
COPY go.sum /go-ethereum/

# 进入工作目录并下载所有 Go 模块依赖
RUN cd /go-ethereum && go mod download

# 将当前目录下的所有源代码添加到容器的 /go-ethereum/ 目录
# 这一步通常在下载依赖之后，因为源代码变化很频繁，放在后面可以利用前面的缓存
ADD . /go-ethereum

# 编译 Geth：
# - cd /go-ethereum: 进入源码目录
# - go run build/ci.go install: 运行 Geth 项目自带的构建脚本 (ci.go)
# - -static: 告诉编译器生成静态链接的二进制文件（不依赖外部动态库，方便移植）
# - ./cmd/geth: 指定要编译的目标程序是 cmd/geth
RUN cd /go-ethereum && go run build/ci.go install -static ./cmd/geth

# ==============================================================================
# 第二阶段：运行阶段 (Runtime Stage)
# 使用轻量级的 Alpine Linux 最新版作为基础镜像
# 这是最终发布镜像的基础，体积非常小
# ==============================================================================
FROM alpine:latest

# 安装 ca-certificates 根证书库
# 这是必须的，因为 Geth 需要通过 HTTPS 与其他节点通信或进行数据同步
RUN apk add --no-cache ca-certificates

# 从第一阶段 (builder) 中复制编译好的 geth 二进制文件
# 将其放到 /usr/local/bin/ 目录下，这样就可以直接在命令行运行 'geth'
COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

# 声明容器运行时监听的端口：
# - 8545: HTTP JSON-RPC 接口（默认端口）
# - 8546: WebSocket 接口（默认端口）
# - 30303: P2P 网络监听端口（TCP）
# - 30303/udp: P2P 网络发现端口（UDP）
EXPOSE 8545 8546 30303 30303/udp

# 设置容器的入口点
# 当容器启动时，默认执行 'geth' 命令
# 用户在 'docker run' 后面追加的参数会传递给 geth
ENTRYPOINT ["geth"]

# ==============================================================================
# 添加元数据标签
# ==============================================================================
# 重新声明参数，因为 ARG 在多阶段构建中不会自动传递到新阶段
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# 将构建参数的值写入镜像的 Label 中，方便通过 'docker inspect' 查看镜像信息
LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
