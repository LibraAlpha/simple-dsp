# 设置全局参数
ARG ALPINE_MIRROR=mirrors.aliyun.com
ARG POSTGRES_VERSION=13.9-r0
ARG REDIS_VERSION=6.2.7-r0

# 定义镜像源设置函数
FROM alpine:3.16 AS base
ARG ALPINE_MIRROR
RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories && \
    apk update

# 构建阶段
FROM golang:1.19-alpine3.16 AS builder
ARG ALPINE_MIRROR
COPY --from=base /etc/apk/repositories /etc/apk/repositories
WORKDIR /app

# 设置Go模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 安装构建依赖
RUN apk update && \
    apk add --no-cache \
        gcc \
        musl-dev \
        postgresql-dev

# 复制并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并构建
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -o main ./cmd/main.go

# 运行阶段
FROM alpine:3.16 AS runner
ARG ALPINE_MIRROR
ARG POSTGRES_VERSION
ARG REDIS_VERSION

# 复制预配置的镜像源
COPY --from=base /etc/apk/repositories /etc/apk/repositories
WORKDIR /app

# 安装运行时依赖
RUN apk update && \
    apk add --no-cache \
        ca-certificates \
        tzdata \
        postgresql-client=${POSTGRES_VERSION} \
        redis=${REDIS_VERSION}

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制所需文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY migrations ./migrations
COPY scripts/init.sh ./init.sh
RUN chmod +x ./init.sh

# 暴露应用端口
EXPOSE 8080

# 设置环境变量
ENV DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=simple_dsp \
    REDIS_HOST=redis \
    REDIS_PORT=6379

# 启动命令
CMD ["sh", "-c", "./init.sh && ./main"]
