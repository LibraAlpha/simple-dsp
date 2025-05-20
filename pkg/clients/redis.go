/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: redis.go
 * Project: simple-dsp
 * Description: Redis客户端封装，提供统一的Redis操作接口
 *
 * 主要功能:
 * - 提供统一的Redis客户端接口定义
 * - 支持单机和集群模式的Redis连接
 * - 封装常用的Redis操作方法
 * - 提供连接池管理和错误处理
 *
 * 实现细节:
 * - 使用go-redis库实现底层连接
 * - 通过适配器模式封装原生客户端
 * - 支持自动识别单机/集群模式
 * - 实现标准的Redis操作接口
 *
 * 依赖关系:
 * - github.com/go-redis/redis/v8
 * - simple-dsp/pkg/config
 * - simple-dsp/pkg/logger
 *
 * 注意事项:
 * - 需要正确配置Redis连接参数
 * - 集群模式下不支持DB选择
 * - 所有操作都需要传入context
 * - 注意处理连接池资源释放
 */

package clients

import (
	"context"
	"fmt"
	"time"

	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"

	"github.com/go-redis/redis/v8"
)

// InitRedis initRedis 初始化Redis客户端
func InitRedis(cfg *config.Config, log *logger.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addresses[0],
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis连接失败", "error", err)
	}

	return rdb, nil
}

// NewRedisClient 初始化 Redis 客户端（自动适配单机/集群）
func NewRedisClient(cfg config.RedisConfig, log *logger.Logger) (redis.UniversalClient, error) {
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

	return baseClient, nil
}
