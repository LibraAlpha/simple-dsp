# 构建阶段
FROM golang:1.19-alpine3.16 AS builder

WORKDIR /app

# 设置Go模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 设置Alpine镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装依赖
RUN apk update && \
    apk add --no-cache \
    gcc \
    musl-dev \
    postgresql-client=13.9-r0 \
    postgresql-dev=13.9-r0

# 安装 protoc 插件
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 生成 protobuf 代码
RUN make proto

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -o main ./cmd/main.go

# 运行阶段
FROM alpine:3.16

WORKDIR /app

# 设置Alpine镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
    ca-certificates \
    tzdata \
    postgresql-client=13.9-r0 \
    redis=6.2.7-r0

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
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
