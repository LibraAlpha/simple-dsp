# 设置错误时停止执行
$ErrorActionPreference = "Stop"

# 定义颜色函数
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    else {
        $input | Write-Output
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

# 检查必要工具
function Test-Requirements {
    Write-ColorOutput Green "正在检查环境要求..."
    
    # 检查 Go 版本
    try {
        $goVersion = go version
        Write-ColorOutput Green "✓ Go 已安装: $goVersion"
    }
    catch {
        Write-ColorOutput Red "✗ 未安装 Go，请先安装 Go 1.19 或更高版本"
        exit 1
    }
    
    # 检查 Docker
    try {
        $dockerVersion = docker version
        Write-ColorOutput Green "✓ Docker 已安装"
    }
    catch {
        Write-ColorOutput Red "✗ 未安装 Docker，请先安装 Docker"
        exit 1
    }
    
    # 检查 Docker Compose
    try {
        $dockerComposeVersion = docker-compose version
        Write-ColorOutput Green "✓ Docker Compose 已安装"
    }
    catch {
        Write-ColorOutput Red "✗ 未安装 Docker Compose，请先安装 Docker Compose"
        exit 1
    }
}

# 运行测试
function Run-Tests {
    Write-ColorOutput Green "正在运行测试..."
    go test ./... -v
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput Red "✗ 测试失败"
        exit 1
    }
    Write-ColorOutput Green "✓ 测试通过"
}

# 构建项目
function Build-Project {
    Write-ColorOutput Green "正在构建项目..."
    
    # 清理旧的构建文件
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin"
    }
    New-Item -ItemType Directory -Force -Path "bin"
    
    # 构建 DSP 服务
    Write-ColorOutput Green "构建 DSP 服务..."
    go build -o bin/dsp-server.exe ./cmd/dsp-server
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput Red "✗ DSP 服务构建失败"
        exit 1
    }
    
    # 构建管理后台服务
    Write-ColorOutput Green "构建管理后台服务..."
    go build -o bin/admin-server.exe ./cmd/admin-server
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput Red "✗ 管理后台服务构建失败"
        exit 1
    }
    
    Write-ColorOutput Green "✓ 项目构建完成"
}

# 启动服务
function Start-Services {
    Write-ColorOutput Green "正在启动服务..."
    
    # 停止并删除旧容器
    docker-compose down
    
    # 构建并启动新容器
    docker-compose up -d --build
    
    # 等待服务启动
    Write-ColorOutput Green "等待服务启动..."
    Start-Sleep -Seconds 10
    
    # 检查服务健康状态
    $services = @("postgres", "redis", "app", "web")
    foreach ($service in $services) {
        $status = docker-compose ps $service
        if ($status -match "Up") {
            Write-ColorOutput Green "✓ $service 服务已启动"
        }
        else {
            Write-ColorOutput Red "✗ $service 服务启动失败"
            docker-compose logs $service
            exit 1
        }
    }
}

# 显示服务状态
function Show-Status {
    Write-ColorOutput Green "`n服务状态："
    docker-compose ps
    
    Write-ColorOutput Green "`n服务访问地址："
    Write-ColorOutput Yellow "DSP 服务: http://localhost:8080"
    Write-ColorOutput Yellow "管理后台: http://localhost:8081"
    Write-ColorOutput Yellow "前端页面: http://localhost:80"
}

# 主函数
function Main {
    Write-ColorOutput Green "=== Simple DSP 一键启动脚本 ==="
    
    # 检查环境
    Test-Requirements
    
    # 运行测试
    Run-Tests
    
    # 构建项目
    Build-Project
    
    # 启动服务
    Start-Services
    
    # 显示状态
    Show-Status
    
    Write-ColorOutput Green "`n=== 服务启动完成 ==="
}

# 执行主函数
Main 