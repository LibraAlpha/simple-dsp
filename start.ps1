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

# 检测操作系统并执行相应的启动脚本
function Start-Project {
    Write-ColorOutput Green "=== Simple DSP 启动器 ==="
    
    # 检测操作系统
    if ($IsLinux) {
        Write-ColorOutput Green "检测到 Linux 系统，使用 shell 脚本启动..."
        if (Test-Path "scripts/start.sh") {
            # 确保脚本有执行权限
            chmod +x scripts/start.sh
            # 执行 shell 脚本
            bash scripts/start.sh
        }
        else {
            Write-ColorOutput Red "未找到 Linux 启动脚本 (scripts/start.sh)"
            exit 1
        }
    }
    elseif ($IsWindows) {
        Write-ColorOutput Green "检测到 Windows 系统，使用 PowerShell 脚本启动..."
        if (Test-Path "scripts/start.ps1") {
            # 执行 PowerShell 脚本
            & "scripts/start.ps1"
        }
        else {
            Write-ColorOutput Red "未找到 Windows 启动脚本 (scripts/start.ps1)"
            exit 1
        }
    }
    else {
        Write-ColorOutput Red "不支持的操作系统"
        exit 1
    }
}

# 执行启动
Start-Project 