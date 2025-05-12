# Redis键值设计

## 1. 频次控制相关
### 1.1 曝光频次
- 键格式：`freq:imp:{user_id}:{creative_id}`
- 类型：Sorted Set
- 成员：时间戳(毫秒)
- 分数：时间戳(毫秒)
- 过期时间：24小时
- 说明：记录用户对特定素材的曝光时间

### 1.2 点击频次
- 键格式：`freq:click:{user_id}:{creative_id}`
- 类型：Sorted Set
- 成员：时间戳(毫秒)
- 分数：时间戳(毫秒)
- 过期时间：24小时
- 说明：记录用户对特定素材的点击时间

## 2. 素材管理相关
### 2.1 素材审核记录
- 键格式：`creative:audit:{creative_id}`
- 类型：String
- 值：审核记录JSON
- 过期时间：永久
- 说明：存储素材最新的审核状态

### 2.2 素材审核历史
- 键格式：`creative:audit:history:{creative_id}`
- 类型：List
- 成员：审核记录JSON
- 过期时间：永久
- 说明：存储素材的审核历史记录

### 2.3 素材版本信息
- 键格式：`creative:version:{creative_id}:{version}`
- 类型：String
- 值：版本信息JSON
- 过期时间：永久
- 说明：存储素材特定版本的信息

### 2.4 素材当前版本
- 键格式：`creative:current_version:{creative_id}`
- 类型：String
- 值：当前版本号
- 过期时间：永久
- 说明：记录素材的当前版本号

## 3. 分片上传相关
### 3.1 上传任务信息
- 键格式：`upload:{upload_id}`
- 类型：String
- 值：上传任务JSON
- 过期时间：24小时
- 说明：存储分片上传任务的信息

### 3.2 分片信息
- 键格式：`upload:{upload_id}:chunk:{chunk_index}`
- 类型：String
- 值：分片信息JSON
- 过期时间：24小时
- 说明：存储上传分片的信息

## 4. 配置管理相关
### 4.1 配置项
- 键格式：`config:{key}`
- 类型：String
- 值：配置JSON
- 过期时间：永久
- 说明：存储系统配置项

## 5. 缓存相关
### 5.1 素材缓存
- 键格式：`cache:creative:{creative_id}`
- 类型：String
- 值：素材JSON
- 过期时间：1小时
- 说明：缓存素材信息

### 5.2 广告缓存
- 键格式：`cache:ad:{ad_id}`
- 类型：String
- 值：广告JSON
- 过期时间：1小时
- 说明：缓存广告信息

## 6. 监控相关
### 6.1 实时指标
- 键格式：`metrics:{metric_name}:{timestamp}`
- 类型：String
- 值：指标值
- 过期时间：1小时
- 说明：存储实时监控指标

### 6.2 计数器
- 键格式：`counter:{counter_name}:{date}`
- 类型：String
- 值：计数值
- 过期时间：7天
- 说明：存储每日计数器

## 注意事项
1. 所有时间相关的值使用毫秒级时间戳
2. JSON数据需要进行压缩处理
3. 重要数据需要持久化
4. 合理设置过期时间，避免内存占用过大
5. 关键操作需要使用Pipeline或事务
6. 定期清理过期数据 