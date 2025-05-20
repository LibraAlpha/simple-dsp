/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: collector.go
 * Project: simple-dsp
 * Description: 数据统计收集器，负责收集和处理广告数据
 *
 * 主要功能:
 * - 收集广告展示数据
 * - 收集广告点击数据
 * - 收集广告转化数据
 * - 提供数据统计接口
 *
 * 实现细节:
 * - 使用Kafka异步收集数据
 * - 实现数据聚合和统计
 * - 支持实时数据查询
 * - 提供数据导出功能
 *
 * 依赖关系:
 * - simple-dsp/pkg/clients
 * - simple-dsp/pkg/metrics
 * - simple-dsp/pkg/logger
 *
 * 注意事项:
 * - 注意数据收集的实时性
 * - 合理设置数据缓存
 * - 注意处理数据一致性
 * - 确保数据安全性
 */

package stats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"

	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// EventType 事件类型
type EventType string

const (
	// EventImpression 展示事件
	EventImpression EventType = "impression"
	// EventClick 点击事件
	EventClick EventType = "click"
	// EventConversion 转化事件
	EventConversion EventType = "conversion"
)

// Event 事件数据
type Event struct {
	EventType   EventType         `json:"event_type"`
	RequestID   string            `json:"request_id"`
	UserID      string            `json:"user_id"`
	AdID        string            `json:"ad_id"`
	SlotID      string            `json:"slot_id"`
	BidPrice    float64           `json:"bid_price"`
	WinPrice    float64           `json:"win_price"`
	Timestamp   time.Time         `json:"timestamp"`
	IP          string            `json:"ip"`
	UserAgent   string            `json:"user_agent"`
	ExtraParams map[string]string `json:"extra_params"`
}

// Collector 数据统计收集器
type Collector struct {
	logger      *logger.Logger
	metrics     *metrics.Metrics
	kafkaClient *kafka.Writer
	redisClient *redis.Client
}

// NewCollector 创建新的数据统计收集器
func NewCollector(kafkawriter *kafka.Writer, redisClient *redis.Client, logger *logger.Logger, metrics *metrics.Metrics) *Collector {
	return &Collector{
		logger:      logger,
		metrics:     metrics,
		kafkaClient: kafkawriter,
		redisClient: redisClient,
	}
}

// CollectEvent 收集事件数据
func (c *Collector) CollectEvent(ctx context.Context, event *Event) error {
	// 记录事件到Kafka
	eventBytes, err := json.Marshal(event)
	if err != nil {
		c.logger.Error("序列化事件数据失败", "error", err)
		return err
	}

	// 发送到Kafka
	topic := getEventTopic(event.EventType)
	if err := c.kafkaClient.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: eventBytes,
	}); err != nil {
		c.logger.Error("发送事件到Kafka失败", "error", err, "event_type", event.EventType)
		return err
	}

	// 更新实时计数器
	if err := c.updateRealtimeCounters(ctx, event); err != nil {
		c.logger.Error("更新实时计数器失败", "error", err)
		// 不返回错误，因为Kafka已经成功发送
	}

	// 更新监控指标
	c.updateMetrics(event)

	return nil
}

// GetRealtimeStats 获取实时统计数据
func (c *Collector) GetRealtimeStats(ctx context.Context, adID string) (*RealtimeStats, error) {
	now := time.Now()
	date := now.Format("2006-01-02")

	// 获取展示数
	impKey := getRealtimeKey(adID, date, EventImpression)
	impCount := c.redisClient.Get(ctx, impKey).String()

	// 获取点击数
	clickKey := getRealtimeKey(adID, date, EventClick)
	clickCount := c.redisClient.Get(ctx, clickKey).String()

	// 获取转化数
	convKey := getRealtimeKey(adID, date, EventConversion)
	convCount := c.redisClient.Get(ctx, convKey).String()

	// 获取消耗
	costKey := getRealtimeCostKey(adID, date)
	cost := c.redisClient.Get(ctx, costKey).String()

	return &RealtimeStats{
		AdID:        adID,
		Date:        date,
		Impressions: parseInt64(impCount),
		Clicks:      parseInt64(clickCount),
		Conversions: parseInt64(convCount),
		Cost:        parseFloat64(cost),
		CTR:         calculateCTR(parseInt64(impCount), parseInt64(clickCount)),
		CVR:         calculateCVR(parseInt64(clickCount), parseInt64(convCount)),
		UpdateTime:  now,
	}, nil
}

// RealtimeStats 实时统计数据
type RealtimeStats struct {
	AdID        string    `json:"ad_id"`
	Date        string    `json:"date"`
	Impressions int64     `json:"impressions"`
	Clicks      int64     `json:"clicks"`
	Conversions int64     `json:"conversions"`
	Cost        float64   `json:"cost"`
	CTR         float64   `json:"ctr"`
	CVR         float64   `json:"cvr"`
	UpdateTime  time.Time `json:"update_time"`
}

// updateRealtimeCounters 更新实时计数器
func (c *Collector) updateRealtimeCounters(ctx context.Context, event *Event) error {
	date := event.Timestamp.Format("2006-01-02")

	// 更新事件计数
	eventKey := getRealtimeKey(event.AdID, date, event.EventType)
	_ = c.redisClient.IncrBy(ctx, eventKey, 1)

	// 如果是展示事件，更新消耗
	if event.EventType == EventImpression && event.WinPrice > 0 {
		costKey := getRealtimeCostKey(event.AdID, date)
		_ = c.redisClient.IncrBy(ctx, costKey, int64(event.WinPrice*100))
	}

	return nil
}

// updateMetrics 更新监控指标
func (c *Collector) updateMetrics(event *Event) {
	labels := map[string]string{
		"ad_id":   event.AdID,
		"slot_id": event.SlotID,
	}

	switch event.EventType {
	case EventImpression:
		c.metrics.Events.Impressions.With(prometheus.Labels{
			"ad_id":   event.AdID,
			"slot_id": event.SlotID,
		}).Inc()
		if event.WinPrice > 0 {
			costCents := event.WinPrice * 100 // 转换为整数单位避免浮点精度问题
			c.metrics.Budget.Cost.With(prometheus.Labels{
				"ad_id": event.AdID,
				"type":  "impression_cost", // 区分成本类型
			}).Add(float64(costCents))
		} else {
			//c.metrics.Errors.WithLabelValues("invalid_winprice").Inc()
		}
	case EventClick:
		c.metrics.Events.Clicks.WithLabelValues(labels["ad_id"], labels["slot_id"]).Inc()
	case EventConversion:
		c.metrics.Events.Conversions.WithLabelValues(labels["ad_id"], labels["slot_id"]).Inc()
	}
}

// getEventTopic 获取事件对应的Kafka主题
func getEventTopic(eventType EventType) string {
	return "dsp.events." + string(eventType)
}

// getRealtimeKey 获取实时统计的Redis键
func getRealtimeKey(adID, date string, eventType EventType) string {
	return "stats:realtime:" + adID + ":" + date + ":" + string(eventType)
}

// getRealtimeCostKey 获取实时消耗的Redis键
func getRealtimeCostKey(adID, date string) string {
	return "stats:realtime:" + adID + ":" + date + ":cost"
}

// parseInt64 解析字符串为int64
func parseInt64(s string) int64 {
	var i int64
	err := json.Unmarshal([]byte(s), &i)
	if err != nil {
		return 0
	}
	return i
}

// parseFloat64 解析字符串为float64
func parseFloat64(s string) float64 {
	var f float64
	err := json.Unmarshal([]byte(s), &f)
	if err != nil {
		return 0
	}
	return f
}

// calculateCTR 计算点击率
func calculateCTR(impressions, clicks int64) float64 {
	if impressions == 0 {
		return 0
	}
	return float64(clicks) / float64(impressions)
}

// calculateCVR 计算转化率
func calculateCVR(clicks, conversions int64) float64 {
	if clicks == 0 {
		return 0
	}
	return float64(conversions) / float64(clicks)
}
