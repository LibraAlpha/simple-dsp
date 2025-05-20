#!/bin/bash

# 设置错误时停止执行
set -e

# 定义颜色函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_color() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# 检查必要工具
check_requirements() {
    print_color "$GREEN" "正在检查环境要求..."
    
    # 检查 Go 版本
    if ! command -v go &> /dev/null; then
        print_color "$RED" "✗ 未安装 Go，请先安装 Go 1.19 或更高版本"
        exit 1
    fi
    go_version=$(go version)
    print_color "$GREEN" "✓ Go 已安装: $go_version"
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        print_color "$RED" "✗ 未安装 Docker，请先安装 Docker"
        exit 1
    fi
    print_color "$GREEN" "✓ Docker 已安装"
    
    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_color "$RED" "✗ 未安装 Docker Compose，请先安装 Docker Compose"
        exit 1
    fi
    print_color "$GREEN" "✓ Docker Compose 已安装"
}

# 运行测试
run_tests() {
    print_color "$GREEN" "正在运行测试..."
    go test ./... -v
    if [ $? -ne 0 ]; then
        print_color "$RED" "✗ 测试失败"
        exit 1
    fi
    print_color "$GREEN" "✓ 测试通过"
}

# 构建项目
build_project() {
    print_color "$GREEN" "正在构建项目..."
    
    # 清理旧的构建文件
    rm -rf bin
    mkdir -p bin
    
    # 构建 DSP 服务
    print_color "$GREEN" "构建 DSP 服务..."
    go build -o bin/dsp-server ./cmd/dsp-server
    if [ $? -ne 0 ]; then
        print_color "$RED" "✗ DSP 服务构建失败"
        exit 1
    fi
    
    # 构建管理后台服务
    print_color "$GREEN" "构建管理后台服务..."
    go build -o bin/admin-server ./cmd/admin-server
    if [ $? -ne 0 ]; then
        print_color "$RED" "✗ 管理后台服务构建失败"
        exit 1
    fi
    
    print_color "$GREEN" "✓ 项目构建完成"
}

# 启动服务
start_services() {
    print_color "$GREEN" "正在启动服务..."
    
    # 停止并删除旧容器
    docker-compose down
    
    # 构建并启动新容器
    docker-compose up -d --build
    
    # 等待服务启动
    print_color "$GREEN" "等待服务启动..."
    sleep 10
    
    # 检查服务健康状态
    services=("postgres" "redis" "app" "web")
    for service in "${services[@]}"; do
        if docker-compose ps $service | grep -q "Up"; then
            print_color "$GREEN" "✓ $service 服务已启动"
        else
            print_color "$RED" "✗ $service 服务启动失败"
            docker-compose logs $service
            exit 1
        fi
    done
}

# 显示服务状态
show_status() {
    print_color "$GREEN" "\n服务状态："
    docker-compose ps
    
    print_color "$GREEN" "\n服务访问地址："
    print_color "$YELLOW" "DSP 服务: http://localhost:8080"
    print_color "$YELLOW" "管理后台: http://localhost:8081"
    print_color "$YELLOW" "前端页面: http://localhost:80"
}

# 主函数
main() {
    print_color "$GREEN" "=== Simple DSP 一键启动脚本 ==="
    
    # 检查环境
    check_requirements
    
    # 运行测试
    run_tests
    
    # 构建项目
    build_project
    
    # 启动服务
    start_services
    
    # 显示状态
    show_status
    
    print_color "$GREEN" "\n=== 服务启动完成 ==="
}

# 执行主函数
main 