.PHONY: all build clean proto

# 默认目标
all: proto build

# 构建应用
build:
	go build -o bin/dsp-server cmd/server/main.go

# 清理构建产物
clean:
	rm -rf bin/
	rm -rf api/proto/dsp/v1/*.pb.go

# 生成 protobuf 代码
proto:
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		api/proto/dsp/v1/*.proto

# 安装依赖工具
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 运行测试
test:
	go test -v ./... 