# Simple DSP (需求方平台)

一个简单的需求方平台(DSP)系统，用于广告投放和竞价。

## 功能特性

- 流量接入：处理广告请求，支持RTA和实时竞价
- 预算控制：实时预算管理和控制
- 数据统计：收集和处理展示、点击、转化等事件数据
- 事件处理：处理各类广告事件，支持实时统计
- 监控指标：支持Prometheus指标收集和监控

## 系统要求

- Go 1.21+
- Redis 6.0+
- Kafka 2.8+
- Prometheus (可选，用于监控)

## 快速开始

1. 克隆项目
```bash
git clone https://github.com/your-username/simple-dsp.git
cd simple-dsp
```

2. 安装依赖
```bash
go mod download
```

3. 配置
复制示例配置文件并根据需要修改：
```bash
cp configs/config.yaml.example configs/config.yaml
```

4. 编译
```bash
go build -o bin/dsp-server cmd/dsp-server/main.go
```

5. 运行
```bash
./bin/dsp-server
```

## 配置说明

配置文件位于`configs/config.yaml`，主要包含以下配置项：

- 服务器配置：端口、超时时间等
- 流量配置：QPS限制、超时时间等
- RTA配置：服务地址、重试策略等
- 竞价配置：并发数、价格限制等
- 预算配置：检查间隔、告警阈值等
- 数据统计配置：Kafka主题、Redis前缀等
- 事件处理配置：重试策略、队列大小等
- Redis配置：连接信息、连接池等
- Kafka配置：代理地址、消费者组等
- 日志配置：日志级别、文件管理等
- 监控配置：指标端口、推送网关等

## API接口

### 流量接入
- POST `/api/v1/traffic` - 处理广告请求

### 事件处理
- POST `/api/v1/events/impression` - 处理展示事件
- POST `/api/v1/events/click` - 处理点击事件
- POST `/api/v1/events/conversion` - 处理转化事件
- GET `/api/v1/events/stats` - 获取事件统计

### 健康检查
- GET `/health` - 健康检查接口

## 监控指标

系统提供以下监控指标：

- 请求处理延迟
- 请求成功率
- 竞价成功率
- 预算使用率
- 事件处理延迟
- 系统资源使用情况

## 开发说明

### 项目结构
```
.
├── cmd/                    # 命令行工具
│   └── dsp-server/        # 主程序入口
├── configs/               # 配置文件
├── internal/              # 内部包
│   ├── bidding/          # 竞价引擎
│   ├── budget/           # 预算管理
│   ├── event/            # 事件处理
│   ├── rta/              # RTA服务
│   ├── stats/            # 数据统计
│   └── traffic/          # 流量接入
├── pkg/                   # 公共包
│   ├── config/           # 配置管理
│   ├── logger/           # 日志工具
│   └── metrics/          # 监控指标
└── scripts/              # 脚本工具
```

### 开发环境设置

1. 安装开发工具
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

2. 运行代码检查
```bash
golangci-lint run
```

3. 运行测试
```bash
go test ./...
```

## 部署说明

### Docker部署

1. 构建镜像
```bash
docker build -t simple-dsp:latest .
```

2. 运行容器
```bash
docker run -d \
  --name dsp-server \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  simple-dsp:latest
```

### Kubernetes部署

1. 创建命名空间
```bash
kubectl create namespace dsp
```

2. 创建配置
```bash
kubectl create configmap dsp-config --from-file=configs/config.yaml -n dsp
```

3. 部署应用
```bash
kubectl apply -f k8s/deployment.yaml
```

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交变更
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License 