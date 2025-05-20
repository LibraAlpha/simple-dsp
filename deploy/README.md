# Simple DSP 分布式部署指南

## 目录结构
```
deploy/
├── configs/                 # 各环境配置文件
│   ├── dev/                # 开发环境
│   ├── test/               # 测试环境
│   └── prod/               # 生产环境
├── docker/                 # Docker 相关配置
│   ├── dsp-server/        # DSP 服务配置
│   ├── admin-server/      # 管理后台配置
│   └── web/               # 前端配置
├── nginx/                  # Nginx 配置
│   ├── conf.d/            # 站点配置
│   └── ssl/               # SSL 证书
└── scripts/               # 部署脚本
    ├── deploy.sh          # 通用部署脚本
    └── health-check.sh    # 健康检查脚本
```

## 部署架构

### 1. 服务组件
- DSP 服务集群 (dsp-server)
- 管理后台服务 (admin-server)
- 前端服务 (web)
- 数据库集群 (PostgreSQL)
- 缓存集群 (Redis)
- 消息队列 (Kafka)
- 负载均衡 (Nginx)

### 2. 网络架构
```
                    [负载均衡器]
                         |
        +----------------+----------------+
        |                |                |
    [DSP集群]       [管理后台]        [前端服务]
        |                |                |
    +---+---+        +---+---+        +---+---+
    |       |        |       |        |       |
[PostgreSQL]    [Redis集群]    [Kafka集群]
```

### 3. 配置说明

#### 3.1 环境变量
- 开发环境: `deploy/configs/dev/`
- 测试环境: `deploy/configs/test/`
- 生产环境: `deploy/configs/prod/`

每个环境目录下包含：
- `dsp-server.env` - DSP 服务配置
- `admin-server.env` - 管理后台配置
- `web.env` - 前端配置
- `database.env` - 数据库配置
- `redis.env` - Redis 配置
- `kafka.env` - Kafka 配置

#### 3.2 服务发现
- 使用 Consul 进行服务发现
- 配置在 `deploy/configs/{env}/consul.env`

#### 3.3 监控配置
- Prometheus 监控配置
- Grafana 仪表板配置
- 日志收集配置

## 部署步骤

1. 准备环境
```bash
# 安装必要工具
./deploy/scripts/prepare-env.sh
```

2. 配置环境变量
```bash
# 选择环境
export DEPLOY_ENV=prod  # 或 dev, test
```

3. 部署服务
```bash
# 部署所有服务
./deploy/scripts/deploy.sh all

# 或部署单个服务
./deploy/scripts/deploy.sh dsp-server
./deploy/scripts/deploy.sh admin-server
./deploy/scripts/deploy.sh web
```

4. 验证部署
```bash
# 运行健康检查
./deploy/scripts/health-check.sh
```

## 注意事项

1. 安全配置
   - 所有服务间通信使用 TLS
   - 使用 Vault 管理密钥
   - 配置防火墙规则

2. 高可用
   - 服务多实例部署
   - 数据库主从复制
   - Redis 集群模式

3. 监控告警
   - 配置服务健康检查
   - 设置资源使用告警
   - 配置日志收集

4. 备份策略
   - 数据库定时备份
   - 配置文件版本控制
   - 日志归档策略 