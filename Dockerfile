# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dsp-server cmd/dsp-server/main.go

# 运行阶段
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/dsp-server .
COPY --from=builder /app/configs ./configs

# 创建日志目录
RUN mkdir -p /app/logs && chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080 9090

# 设置环境变量
ENV CONFIG_PATH=/app/configs/config.yaml

# 运行应用
CMD ["./dsp-server"]
