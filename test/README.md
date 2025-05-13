# DSP系统测试用例说明

本目录包含了DSP系统的各个模块的测试用例，主要分为以下几个部分：

## 目录结构

```
test/
├── bidding/        # 竞价引擎测试
├── grpc/           # gRPC服务测试
├── http/           # HTTP接口测试
├── rta/            # RTA服务测试
└── README.md       # 本说明文件
```

## 测试模块说明

### 1. 竞价引擎测试 (bidding/)

位于 `test/bidding/engine_test.go`，主要测试竞价引擎的核心功能：

- 正常竞价请求处理
- 无效请求处理
- 边界条件测试

运行测试：
```bash
go test -v ./test/bidding
```

### 2. gRPC服务测试 (grpc/)

位于 `test/grpc/bid_service_test.go`，测试gRPC服务接口：

- 基本竞价请求响应
- 服务端流式处理
- 错误处理
- 超时处理

运行测试：
```bash
go test -v ./test/grpc
```

### 3. HTTP接口测试 (http/)

位于 `test/http/bid_handler_test.go`，测试HTTP REST接口：

- 正常竞价请求
- 参数验证
- 错误处理
- HTTP状态码验证

运行测试：
```bash
go test -v ./test/http
```

### 4. RTA服务测试 (rta/)

位于 `test/rta/` 目录下，包含以下测试文件：

#### 4.1 client_test.go
测试RTA客户端功能：
- 单次查询接口
- 批量查询接口
- 参数验证
- 错误处理
- 超时控制

#### 4.2 config_test.go
测试RTA配置管理功能：
- 配置的增删改查
- 配置验证
- 并发安全性
- 默认配置处理

#### 4.3 mock_server.go
模拟RTA服务器：
- 请求参数验证
- 响应模拟
- 错误场景模拟

运行RTA测试：
```bash
go test -v ./test/rta
```

## RTA配置示例

```json
{
  "task_id": "test_task_1",
  "channel": "test_channel",
  "advertising_space_id": "test_ad_space",
  "timeout": "100ms",
  "enabled": true,
  "priority": 1,
  "retry_count": 2,
  "retry_interval": "50ms",
  "cache_expiration": "5m"
}
```

## 运行所有测试

要运行所有测试用例：

```bash
go test -v ./test/...
```

## 测试数据说明

### RTA请求示例

```json
{
  "channel": "test_channel",
  "advertising_space_id": "test_ad_space",
  "imei": "123456789012345",
  "os": "0"
}
```

### RTA批量请求示例

```json
{
  "channel": "test_channel",
  "advertising_space_id": "test_ad_space",
  "imei_md5_list": "abc123,def456,ghi789"
}
```

## 注意事项

1. 运行测试前确保已安装所有依赖：
```bash
go mod download
```

2. 确保环境变量配置正确：
```bash
export DSP_ENV=test
```

3. RTA配置管理：
   - 配置修改是并发安全的
   - 缓存策略可以通过配置调整
   - 超时控制可以针对每个任务单独设置

4. 测试可能需要网络连接，请确保网络通畅

## 常见问题

1. 测试失败时的排查步骤
2. 如何添加新的测试用例
3. 如何修改测试配置
4. RTA配置动态更新的最佳实践

## 贡献指南

1. 提交新测试用例时请遵循现有的代码风格
2. 确保添加了足够的测试注释
3. 更新本README文件以反映新增的测试用例 