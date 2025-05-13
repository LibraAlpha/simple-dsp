# Simple DSP 系统

这是一个简单的DSP（需求方平台）系统，支持RTA对接和广告计划跟踪功能。

## 系统要求

- PostgreSQL 14+
- Redis 6+
- Go 1.19+

## 快速开始

### 1. 使用Docker Compose（推荐）

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看应用日志
docker-compose logs -f app
```

### 2. 手动安装

确保您的系统已安装以下工具：
- PostgreSQL 14+ 和客户端 (`psql`)
- Redis 6+ 和客户端 (`redis-cli`)
- Go 1.19 或更高版本

配置环境变量：

```bash
# 数据库配置
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=simple_dsp

# Redis配置
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=""
```

### 3. 初始化系统

如果使用Docker Compose，初始化会自动完成。
手动安装时，运行初始化脚本：

```bash
chmod +x scripts/init.sh
./scripts/init.sh
```

这个脚本会：
- 创建必要的数据库和表
- 初始化Redis配置
- 插入测试数据

### 4. 启动服务

使用Docker Compose时服务会自动启动。
手动启动：

```bash
go run cmd/main.go
```

## 系统功能

### RTA对接
- 支持单次和批量请求/响应
- 超时控制和重试机制
- 动态配置管理
- Mock服务器用于测试

### 广告计划跟踪
- 支持三种跟踪类型：DP、点击检测和曝光检测
- 计划级别的动态配置
- 使用JSONB存储跟踪配置
- HTTP接口管理
- 事件处理和监控

## 目录结构

```
.
├── cmd/                # 主程序入口
├── internal/          
│   ├── campaign/      # 广告计划相关
│   ├── models/        # 数据库模型
│   ├── handlers/      # HTTP处理器
│   ├── rta/           # RTA相关实现
│   └── tracking/      # 跟踪服务
├── migrations/        # 数据库迁移文件
├── scripts/          # 脚本文件
├── Dockerfile        # Docker构建文件
├── docker-compose.yml # Docker编排配置
└── test/             # 测试用例和文档
```

## 数据库设计

### advertisers 表
```sql
CREATE TABLE advertisers (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL,
    budget DECIMAL(20,4) NOT NULL,
    create_time TIMESTAMP NOT NULL,
    update_time TIMESTAMP NOT NULL
);
```

### campaigns 表
```sql
CREATE TABLE campaigns (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    advertiser_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    budget DECIMAL(20,4) NOT NULL,
    bid_strategy VARCHAR(32) NOT NULL,
    targeting JSONB,
    tracking_configs JSONB,
    update_time TIMESTAMP NOT NULL,
    create_time TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_campaigns_advertiser 
        FOREIGN KEY (advertiser_id) 
        REFERENCES advertisers (id) 
        ON DELETE CASCADE
);
```

## API文档

### 广告计划管理

- `POST /api/v1/campaigns` - 创建广告计划
- `GET /api/v1/campaigns` - 获取广告计划列表
- `GET /api/v1/campaigns/:id` - 获取单个广告计划
- `PUT /api/v1/campaigns/:id` - 更新广告计划
- `DELETE /api/v1/campaigns/:id` - 删除广告计划
- `PUT /api/v1/campaigns/:id/tracking` - 更新跟踪配置

## 测试数据

初始化脚本会创建以下测试数据：

### 广告主
- ID: adv_001, 名称: 测试广告主1
- ID: adv_002, 名称: 测试广告主2

### 广告计划
- ID: camp_001, 名称: 测试广告计划1 (CPC)
- ID: camp_002, 名称: 测试广告计划2 (CPM)

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request 