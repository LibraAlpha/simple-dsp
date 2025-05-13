#!/bin/bash

echo "开始初始化 DSP 系统..."

# 检查必要的命令是否存在
command -v psql >/dev/null 2>&1 || { echo "需要安装 PostgreSQL 客户端"; exit 1; }
command -v redis-cli >/dev/null 2>&1 || { echo "需要安装 Redis 客户端"; exit 1; }

# 数据库配置
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"postgres"}
DB_NAME=${DB_NAME:-"simple_dsp"}

# Redis配置
REDIS_HOST=${REDIS_HOST:-"localhost"}
REDIS_PORT=${REDIS_PORT:-"6379"}
REDIS_PASSWORD=${REDIS_PASSWORD:-""}

echo "正在初始化 PostgreSQL 数据库..."

# 创建数据库
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME;" || true

# 执行数据库迁移
echo "正在执行数据库迁移..."
for f in migrations/*.up.sql; do
    echo "执行迁移: $f"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$f"
done

echo "正在初始化 Redis..."

# 清理并初始化 Redis
redis-cli -h $REDIS_HOST -p $REDIS_PORT flushall

# 初始化频率控制相关的 Redis 键
redis-cli -h $REDIS_HOST -p $REDIS_PORT HSET frequency:limits user:daily 1000
redis-cli -h $REDIS_HOST -p $REDIS_PORT HSET frequency:limits campaign:daily 500
redis-cli -h $REDIS_HOST -p $REDIS_PORT HSET frequency:limits ip:hourly 100

# 插入测试数据
echo "正在插入测试数据..."

# 插入广告主数据
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
INSERT INTO advertisers (id, name, status, budget, create_time, update_time)
VALUES 
    ('adv_001', '测试广告主1', 'active', 10000.00, NOW(), NOW()),
    ('adv_002', '测试广告主2', 'active', 20000.00, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
EOF

# 插入广告计划数据
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
INSERT INTO campaigns (
    id, name, advertiser_id, status, start_time, end_time,
    budget, bid_strategy, targeting, tracking_configs,
    create_time, update_time
)
VALUES (
    'camp_001',
    '测试广告计划1',
    'adv_001',
    'active',
    NOW(),
    NOW() + INTERVAL '30 days',
    5000.00,
    'cpc',
    '{"locations": ["北京", "上海"], "ages": ["18-24", "25-34"]}',
    '{"click": {"url": "http://track.example.com/click", "method": "GET", "enabled": true}}',
    NOW(),
    NOW()
),
(
    'camp_002',
    '测试广告计划2',
    'adv_002',
    'active',
    NOW(),
    NOW() + INTERVAL '30 days',
    8000.00,
    'cpm',
    '{"locations": ["广州", "深圳"], "ages": ["25-34", "35-44"]}',
    '{"impression": {"url": "http://track.example.com/imp", "method": "POST", "enabled": true}}',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;
EOF

echo "初始化完成！"
echo "数据库: $DB_NAME"
echo "Redis: $REDIS_HOST:$REDIS_PORT"
echo "测试数据已插入" 