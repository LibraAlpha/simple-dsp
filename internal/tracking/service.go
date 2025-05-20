package tracking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"simple-dsp/internal/campaign"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// Service 跟踪服务
type Service struct {
	httpClient *http.Client
	logger     *logger.Logger
	metrics    *metrics.Metrics
	configMgr  *campaign.ConfigManager
}

// TrackingEvent 跟踪事件
type TrackingEvent struct {
	CampaignID string                `json:"campaign_id"`
	EventType  campaign.TrackingType `json:"event_type"`
	Timestamp  time.Time             `json:"timestamp"`
	DeviceID   string                `json:"device_id"`
	IP         string                `json:"ip"`
	UserAgent  string                `json:"user_agent"`
	ExtraData  map[string]string     `json:"extra_data"`
}

// NewService 创建新的跟踪服务
func NewService(configMgr *campaign.ConfigManager, logger *logger.Logger, metrics *metrics.Metrics) *Service {
	return &Service{
		httpClient: &http.Client{},
		logger:     logger,
		metrics:    metrics,
		configMgr:  configMgr,
	}
}

// Track 处理跟踪事件
func (s *Service) Track(ctx context.Context, event *TrackingEvent) error {
	startTime := time.Now()
	defer func() {
		s.metrics.Tracking.Duration.WithLabelValues(string(event.EventType)).Observe(time.Since(startTime).Seconds())
	}()

	// 获取计划配置
	config, exists := s.configMgr.GetConfig(event.CampaignID)
	if !exists {
		return fmt.Errorf("campaign config not found: %s", event.CampaignID)
	}

	// 获取跟踪配置
	trackingConfig, exists := config.TrackingConfigs[event.EventType]
	if !exists || !trackingConfig.Enabled {
		return nil // 跟踪未启用，直接返回
	}

	// 创建HTTP请求
	req, err := s.createTrackingRequest(ctx, trackingConfig, event)
	if err != nil {
		return err
	}

	// 设置超时
	client := &http.Client{
		Timeout: trackingConfig.Timeout,
	}

	// 发送请求（带重试）
	var lastErr error
	for i := 0; i <= trackingConfig.RetryCount; i++ {
		if i > 0 {
			time.Sleep(trackingConfig.RetryInterval)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			s.logger.Error("跟踪请求失败",
				"campaign_id", event.CampaignID,
				"event_type", event.EventType,
				"attempt", i+1,
				"error", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			s.metrics.Tracking.Success.WithLabelValues(string(event.EventType)).Inc()
			return nil
		}

		lastErr = fmt.Errorf("tracking request failed with status code: %d", resp.StatusCode)
		s.logger.Error("跟踪请求返回错误状态码",
			"campaign_id", event.CampaignID,
			"event_type", event.EventType,
			"attempt", i+1,
			"status_code", resp.StatusCode)
	}

	s.metrics.Tracking.Failure.WithLabelValues(string(event.EventType)).Inc()
	return lastErr
}

// createTrackingRequest 创建跟踪请求
func (s *Service) createTrackingRequest(ctx context.Context, config *campaign.TrackingConfig, event *TrackingEvent) (*http.Request, error) {
	// 准备请求数据
	data := map[string]interface{}{
		"campaign_id": event.CampaignID,
		"event_type":  event.EventType,
		"timestamp":   event.Timestamp.Unix(),
		"device_id":   event.DeviceID,
		"ip":          event.IP,
		"user_agent":  event.UserAgent,
	}

	// 添加额外数据
	for k, v := range event.ExtraData {
		data[k] = v
	}

	// 序列化请求体
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 创建请求
	method := config.Method
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, method, config.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
