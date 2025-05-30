# 设置全局参数
ARG ALPINE_MIRROR=mirrors.aliyun.com
ARG GO_VERSION=1.23
ARG ALPINE_VERSION=3.16
ARG POSTGRES_VERSION=13.9-r0
ARG REDIS_VERSION=6.2.7-r0

# 构建阶段
FROM golang:${GO_VERSION} AS builder

# 设置构建环境
ARG ALPINE_MIRROR
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct

# 配置镜像源并安装构建依赖
RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev && \
    # 尝试安装指定版本的postgresql-dev，如果失败则安装最新版本
    (apk add --no-cache "postgresql-dev~=${POSTGRES_VERSION%%-*}" || \
     (echo "指定版本不存在，将安装最新版本的postgresql-dev" && \
      apk add --no-cache postgresql-dev))

# 处理依赖
COPY go.mod go.sum ./
RUN go mod download

# 构建应用
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -o main ./cmd/main.go

# 运行阶段
FROM alpine:${ALPINE_VERSION}

# 设置运行环境
ARG ALPINE_MIRROR
ARG POSTGRES_VERSION
ARG REDIS_VERSION
WORKDIR /app
ENV TZ=Asia/Shanghai \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=simple_dsp \
    REDIS_HOST=redis \
    REDIS_PORT=6379

# 配置镜像源并安装运行时依赖
RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories && \
    # 安装基础包
    apk add --no-cache tzdata ca-certificates && \
    # 尝试安装指定版本的postgresql-client，如果失败则安装最新版本
    (apk add --no-cache "postgresql-client~=${POSTGRES_VERSION%%-*}" || \
     (echo "指定版本不存在，将安装最新版本的postgresql-client" && \
      apk add --no-cache postgresql-client)) && \
    # 尝试安装指定版本的redis，如果失败则安装最新版本
    (apk add --no-cache "redis~=${REDIS_VERSION%%-*}" || \
     (echo "指定版本不存在，将安装最新版本的redis" && \
      apk add --no-cache redis)) && \
    # 打印安装的版本信息
    echo "Installed versions:" && \
    apk info -v postgresql-client redis

# 复制应用文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY migrations ./migrations
COPY scripts/init.sh ./init.sh
RUN chmod +x ./init.sh

# 暴露端口并设置启动命令
EXPOSE 8080
CMD ["sh", "-c", "./init.sh && ./main"]
