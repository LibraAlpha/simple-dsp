# 设置全局参数
ARG BASE_IMAGE=alpine
ARG BASE_IMAGE_VERSION=3.16
ARG GO_VERSION=1.23
ARG POSTGRES_VERSION=13.9-r0
ARG REDIS_VERSION=6.2.7-r0
ARG TARGETPLATFORM=linux/amd64

# 构建阶段
FROM --platform=${TARGETPLATFORM} golang:${GO_VERSION} AS builder

# 设置构建环境
ARG TARGETPLATFORM
ARG BASE_IMAGE
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# 根据目标平台和基础镜像选择构建依赖
RUN if [ "$BASE_IMAGE" = "alpine" ]; then \
        # Alpine 环境
        apk add --no-cache gcc musl-dev && \
        (apk add --no-cache "postgresql-dev~=${POSTGRES_VERSION%%-*}" || \
         (echo "指定版本不存在，将安装最新版本的postgresql-dev" && \
          apk add --no-cache postgresql-dev)); \
    elif [ "$BASE_IMAGE" = "debian" ]; then \
        # Debian/Ubuntu 环境
        apt-get update && apt-get install -y gcc libpq-dev; \
    elif [ "$BASE_IMAGE" = "centos" ]; then \
        # CentOS 环境
        yum update -y && yum install -y gcc postgresql-devel; \
    fi

# 处理依赖
COPY go.mod go.sum ./
RUN go mod download

# 构建应用
COPY . .
RUN go build -a -o main ./cmd/main.go

# 运行阶段
FROM --platform=${TARGETPLATFORM} ${BASE_IMAGE}:${BASE_IMAGE_VERSION}

# 设置运行环境
ARG BASE_IMAGE
ARG POSTGRES_VERSION
ARG REDIS_VERSION
ARG TARGETPLATFORM
WORKDIR /app
ENV TZ=Asia/Shanghai \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=simple_dsp \
    REDIS_HOST=redis \
    REDIS_PORT=6379

# 根据基础镜像安装运行时依赖
RUN if [ "$BASE_IMAGE" = "alpine" ]; then \
        # Alpine 环境
        apk add --no-cache tzdata ca-certificates && \
        (apk add --no-cache "postgresql-client~=${POSTGRES_VERSION%%-*}" || \
         (echo "指定版本不存在，将安装最新版本的postgresql-client" && \
          apk add --no-cache postgresql-client)) && \
        (apk add --no-cache "redis~=${REDIS_VERSION%%-*}" || \
         (echo "指定版本不存在，将安装最新版本的redis" && \
          apk add --no-cache redis)) && \
        echo "Installed versions:" && \
        apk info -v postgresql-client redis; \
    elif [ "$BASE_IMAGE" = "debian" ]; then \
        # Debian/Ubuntu 环境
        apt-get update && \
        apt-get install -y tzdata ca-certificates postgresql-client redis-server && \
        echo "Installed versions:" && \
        dpkg -l | grep -E 'postgresql-client|redis-server'; \
    elif [ "$BASE_IMAGE" = "centos" ]; then \
        # CentOS 环境
        yum update -y && \
        yum install -y tzdata ca-certificates postgresql redis && \
        echo "Installed versions:" && \
        rpm -qa | grep -E 'postgresql|redis'; \
    fi

# 复制应用文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY migrations ./migrations
COPY scripts/init.sh ./init.sh
RUN chmod +x ./init.sh

# 暴露端口并设置启动命令
EXPOSE 8080
CMD ["sh", "-c", "./init.sh && ./main"]
