package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
)

// GoRedisAdapter 适配器实现标准 Redis 客户端
type GoRedisAdapter struct {
	Client redis.UniversalClient // 支持单机和集群模式
}

// RedisClient Redis客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Close() error
}

// NewRedisClient 初始化 Redis 客户端（自动适配单机/集群）
func NewRedisClient(cfg config.RedisConfig, log *logger.Logger) (*GoRedisAdapter, error) {
	var baseClient redis.UniversalClient

	// 根据配置选择部署模式
	if len(cfg.Addresses) > 1 {
		// 集群模式初始化（网页5）
		baseClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cfg.Addresses,
			Password: cfg.Password,
			PoolSize: cfg.PoolSize,
		})
	} else {
		// 单机模式初始化
		baseClient = redis.NewClient(&redis.Options{
			Addr:         cfg.Addresses[0],
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
		})
	}

	// 连接测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := baseClient.Ping(ctx).Err(); err != nil {
		log.Error("Redis 连接失败", "error", err)
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	log.Info("Redis 连接成功", "mode", func() string {
		if len(cfg.Addresses) > 1 {
			return "cluster"
		}
		return "standalone"
	}())

	return &GoRedisAdapter{Client: baseClient}, nil
}

func (a *GoRedisAdapter) Get(ctx context.Context, key string) (string, error) {
	val, err := a.Client.Get(ctx, key).Result()
	if err == redis.Nil { // 处理 key 不存在的情况（网页3）
		return "", nil
	}
	return val, err
}

func (a *GoRedisAdapter) Set(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	return a.Client.Set(ctx, key, value, exp).Err()
}

func (a *GoRedisAdapter) Del(ctx context.Context, keys ...string) (int64, error) {
	// 将字符串切片转换为[]interface{}
	redisKeys := make([]interface{}, len(keys))
	for i, k := range keys {
		redisKeys[i] = k
	}

	// 调用go-redis的Del方法（实际执行UNLINK或DEL命令）
	cmd := a.Client.Del(ctx, keys...)
	return cmd.Result() // 返回(int64, error)
}

func (a *GoRedisAdapter) Incr(ctx context.Context, key string) (int64, error) {
	return a.Client.Incr(ctx, key).Result()
}

func (a *GoRedisAdapter) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	// 必须使用String序列化器
	return a.Client.IncrBy(ctx, key, value).Result()
}

func (a *GoRedisAdapter) Close() error {
	return a.Client.Close()
}
