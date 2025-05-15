package frequency

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"simple-dsp/pkg/clients"
	"strconv"
	"time"

	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// Controller 频次控制器
type Controller struct {
	redis   *clients.GoRedisAdapter
	logger  *logger.Logger
	metrics *metrics.Metrics
}

// Config 频次控制配置
type Config struct {
	ImpressionLimit int           `json:"impression_limit"` // 曝光限制
	ClickLimit      int           `json:"click_limit"`      // 点击限制
	TimeWindow      time.Duration `json:"time_window"`      // 时间窗口
	QPS             float64       `json:"qps"`              // 每秒请求限制
}

// NewController 创建频次控制器
func NewController(redis *clients.GoRedisAdapter, logger *logger.Logger, metrics *metrics.Metrics) *Controller {
	return &Controller{
		redis:   redis,
		logger:  logger,
		metrics: metrics,
	}
}

// CheckImpression 检查曝光频次
func (c *Controller) CheckImpression(ctx context.Context, userID string, adID string) (bool, error) {
	// 获取配置
	config, err := c.getConfig(ctx, adID)
	if err != nil {
		return false, err
	}

	// 生成键名
	key := fmt.Sprintf("freq:imp:%s:%s:%s", userID, adID, time.Now().Format("20060102"))

	// 检查频次
	count, err := c.redis.Client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// 超过限制
	if count >= config.ImpressionLimit {
		c.metrics.Frequency.LimitExceeded.Inc()
		return false, nil
	}

	return true, nil
}

// RecordImpression 记录曝光
func (c *Controller) RecordImpression(ctx context.Context, userID string, adID string) error {
	// 生成键名
	key := fmt.Sprintf("freq:imp:%s:%s:%s", userID, adID, time.Now().Format("20060102"))

	// 增加计数
	_, err := c.redis.Client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	// 设置过期时间
	c.redis.Client.Expire(ctx, key, 24*time.Hour)

	return nil
}

// CheckClick 检查点击频次
func (c *Controller) CheckClick(ctx context.Context, userID string, adID string) (bool, error) {
	// 获取配置
	config, err := c.getConfig(ctx, adID)
	if err != nil {
		return false, err
	}

	// 生成键名
	key := fmt.Sprintf("freq:click:%s:%s:%s", userID, adID, time.Now().Format("20060102"))

	// 检查频次
	count, err := c.redis.Client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// 超过限制
	if count >= config.ClickLimit {
		c.metrics.Frequency.LimitExceeded.Inc()
		return false, nil
	}

	return true, nil
}

// RecordClick 记录点击
func (c *Controller) RecordClick(ctx context.Context, userID string, adID string) error {
	// 生成键名
	key := fmt.Sprintf("freq:click:%s:%s:%s", userID, adID, time.Now().Format("20060102"))

	// 增加计数
	_, err := c.redis.Client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	// 设置过期时间
	c.redis.Client.Expire(ctx, key, 24*time.Hour)

	return nil
}

// UpdateConfig 更新频次控制配置
func (c *Controller) UpdateConfig(ctx context.Context, adID string, config *Config) error {
	// 验证配置
	if err := c.validateConfig(config); err != nil {
		return err
	}

	// 生成键名
	key := fmt.Sprintf("freq:config:%s", adID)

	// 保存配置
	data := map[string]string{
		"impression_limit": strconv.Itoa(config.ImpressionLimit),
		"click_limit":      strconv.Itoa(config.ClickLimit),
		"time_window":      config.TimeWindow.String(),
		"qps":              fmt.Sprintf("%f", config.QPS),
	}

	// 使用 HSET 保存配置
	if err := c.redis.Client.HMSet(ctx, key, data).Err(); err != nil {
		return err
	}

	return nil
}

// GetConfig 获取频次控制配置
func (c *Controller) GetConfig(ctx context.Context, adID string) (*Config, error) {
	return c.getConfig(ctx, adID)
}

// 内部方法

func (c *Controller) getConfig(ctx context.Context, adID string) (*Config, error) {
	// 生成键名
	key := fmt.Sprintf("freq:config:%s", adID)

	// 获取配置
	data, err := c.redis.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 如果配置不存在，使用默认配置
	if len(data) == 0 {
		return &Config{
			ImpressionLimit: 10, // 默认每天最多曝光10次
			ClickLimit:      3,  // 默认每天最多点击3次
			TimeWindow:      24 * time.Hour,
			QPS:             100, // 默认QPS 100
		}, nil
	}

	// 解析配置
	impressionLimit, _ := strconv.Atoi(data["impression_limit"])
	clickLimit, _ := strconv.Atoi(data["click_limit"])
	timeWindow, _ := time.ParseDuration(data["time_window"])
	qps, _ := strconv.ParseFloat(data["qps"], 64)

	return &Config{
		ImpressionLimit: impressionLimit,
		ClickLimit:      clickLimit,
		TimeWindow:      timeWindow,
		QPS:             qps,
	}, nil
}

func (c *Controller) validateConfig(config *Config) error {
	if config.ImpressionLimit <= 0 {
		return fmt.Errorf("曝光限制必须大于0")
	}
	if config.ClickLimit <= 0 {
		return fmt.Errorf("点击限制必须大于0")
	}
	if config.TimeWindow <= 0 {
		return fmt.Errorf("时间窗口必须大于0")
	}
	if config.QPS <= 0 {
		return fmt.Errorf("QPS必须大于0")
	}
	return nil
}
