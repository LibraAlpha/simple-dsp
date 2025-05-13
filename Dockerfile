# 构建阶段
FROM golang:1.19-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache gcc musl-dev protoc make

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
RUN CGO_ENABLED=1 GOOS=linux go build -a -o main ./cmd/server

# 运行阶段
FROM alpine:3.14

WORKDIR /app

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 暴露 HTTP 和 gRPC 端口
EXPOSE 8080 9090

CMD ["./main"]
