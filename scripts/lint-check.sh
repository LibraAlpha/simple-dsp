#!/bin/bash

# 依赖安装检查
if ! command -v golangci-lint &> /dev/null; then
    echo "安装golangci-lint v1.55.2..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
fi

# 代码规范检查
golangci-lint run --fix ./...

# 单元测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 竞态检测
go test -race ./...