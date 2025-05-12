# Simple DSP (需求方平台)

一个简单的广告需求方平台(DSP)实现，支持 CPC 和 CPM 双计费模式。

## 功能特点

- 支持 CPC(按点击付费)和 CPM(按千次展示付费)双计费模式
- 出价策略管理
  - 支持创建时锁定计费类型
  - 支持出价锁定功能
  - 支持多素材关联
- 完整的预算控制
- 实时竞价引擎
- 详细的数据统计

## 技术栈

### 后端
- Go 1.19
- MySQL 8.0
- Redis 6.2
- 依赖管理: Go Modules

### 前端
- Vue 3
- Element Plus
- Vite
- Pinia 状态管理

## 快速开始

### 使用 Docker Compose 部署

1. 克隆项目
```bash
git clone https://github.com/yourusername/simple-dsp.git
cd simple-dsp
```

2. 启动服务
```bash
docker-compose up -d
```

服务将在以下端口启动:
- 前端界面: http://localhost
- 后端 API: http://localhost:8080
- MySQL: localhost:3306
- Redis: localhost:6379

### 手动部署

1. 后端服务
```bash
# 安装依赖
go mod download

# 编译
go build -o main ./cmd/server

# 运行
./main
```

2. 前端服务
```bash
cd web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build
```

## 项目结构

```
.
├── cmd/                    # 入口程序
├── internal/              # 内部包
│   ├── bidding/          # 竞价相关
│   ├── budget/           # 预算控制
│   └── frequency/        # 频次控制
├── pkg/                   # 公共包
├── web/                   # 前端代码
│   ├── src/              # 源代码
│   └── public/           # 静态资源
├── migrations/           # 数据库迁移
├── configs/              # 配置文件
└── docker-compose.yml    # Docker 编排配置
```

## 数据库设计

### bid_strategies 表
```sql
CREATE TABLE bid_strategies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '策略名称',
    bid_type ENUM('CPC', 'CPM') NOT NULL COMMENT '计费类型',
    price DECIMAL(10,4) NOT NULL COMMENT '出价，CPC单位为元，CPM单位为分',
    daily_budget DECIMAL(10,2) NOT NULL COMMENT '日预算，单位为元',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
    is_price_locked TINYINT NOT NULL DEFAULT 1 COMMENT '出价是否锁定：0-未锁定，1-锁定',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### bid_strategy_creatives 表
```sql
CREATE TABLE bid_strategy_creatives (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    strategy_id BIGINT NOT NULL COMMENT '策略ID',
    creative_id BIGINT NOT NULL COMMENT '素材ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (strategy_id) REFERENCES bid_strategies(id) ON DELETE CASCADE
);
```

## API 文档

### 出价策略管理

#### 获取策略列表
```
GET /api/v1/bids
```

#### 创建策略
```
POST /api/v1/bids
```

#### 更新策略
```
PUT /api/v1/bids/:id
```

#### 删除策略
```
DELETE /api/v1/bids/:id
```

### 素材管理

#### 添加素材
```
POST /api/v1/bids/:id/creatives
```

#### 移除素材
```
DELETE /api/v1/bids/:id/creatives/:creativeId
```

### 统计数据

#### 获取策略统计
```
GET /api/v1/bids/:id/stats
```

## 开发团队

- 后端开发: [Your Name]
- 前端开发: [Your Name]

## 许可证

MIT License 