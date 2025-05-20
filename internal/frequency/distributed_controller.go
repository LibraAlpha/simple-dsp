package frequency

import (
	"context"
	"fmt"
	"time"

	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"

	"github.com/go-redis/redis/v8"
)

// DistributedController 分布式频次控制器
type DistributedController struct {
	redis   *redis.Client
	logger  *logger.Logger
	metrics *metrics.Metrics
}

// NewDistributedController 创建分布式频次控制器
func NewDistributedController(redis *redis.Client, logger *logger.Logger, metrics *metrics.Metrics) *DistributedController {
	return &DistributedController{
		redis:   redis,
		logger:  logger,
		metrics: metrics,
	}
}

// CheckFrequency 检查频次限制
func (dc *DistributedController) CheckFrequency(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	start := time.Now()
	defer func() {
		dc.metrics.Frequency.CheckDuration.Observe(time.Since(start).Seconds())
	}()

	// 使用Redis的Sorted Set实现滑动窗口
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	// 使用Pipeline减少网络往返
	pipe := dc.redis.Pipeline()

	// 移除窗口外的记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 获取当前窗口内的记录数
	countCmd := pipe.ZCount(ctx, key, fmt.Sprintf("%d", windowStart), fmt.Sprintf("%d", now))

	// 执行Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		dc.logger.Error("频次检查失败", "error", err)
		return false, err
	}

	count := countCmd.Val()
	allowed := count < int64(limit)

	// 更新指标
	dc.metrics.Frequency.CheckTotal.Inc()
	if !allowed {
		dc.metrics.Frequency.LimitExceeded.Inc()
	}

	return allowed, nil
}

// RecordFrequency 记录频次
func (dc *DistributedController) RecordFrequency(ctx context.Context, key string, window time.Duration) error {
	start := time.Now()
	defer func() {
		dc.metrics.Frequency.RecordDuration.Observe(time.Since(start).Seconds())
	}()

	now := time.Now().UnixNano()

	// 添加记录并设置过期时间
	pipe := dc.redis.Pipeline()
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: now,
	})
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		dc.logger.Error("记录频次失败", "error", err)
		return err
	}

	// 更新指标
	dc.metrics.Frequency.RecordTotal.Inc()

	return nil
}

// GetFrequencyStats 获取频次统计
func (dc *DistributedController) GetFrequencyStats(ctx context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	count, err := dc.redis.ZCount(ctx, key,
		fmt.Sprintf("%d", windowStart),
		fmt.Sprintf("%d", now)).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ClearFrequency 清除频次记录
func (dc *DistributedController) ClearFrequency(ctx context.Context, key string) error {
	return dc.redis.Del(ctx, key).Err()
}
