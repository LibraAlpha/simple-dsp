package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"simple-dsp/pkg/logger"
)

// Service 配置管理服务
type Service struct {
	redis  *redis.Client
	logger *logger.Logger
}

// ConfigItem 配置项
type ConfigItem struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Version   int64       `json:"version"`
	UpdatedAt time.Time   `json:"updated_at"`
	UpdatedBy string      `json:"updated_by"`
}

// NewService 创建配置管理服务
func NewService(redis *redis.Client, logger *logger.Logger) *Service {
	return &Service{
		redis:  redis,
		logger: logger,
	}
}

// SetConfig 设置配置
func (s *Service) SetConfig(ctx context.Context, key string, value interface{}, updatedBy string) error {
	// 获取当前版本
	version, err := s.getCurrentVersion(ctx, key)
	if err != nil {
		version = 0
	}

	// 创建新的配置项
	item := &ConfigItem{
		Key:       key,
		Value:     value,
		Version:   version + 1,
		UpdatedAt: time.Now(),
		UpdatedBy: updatedBy,
	}

	// 序列化配置
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 使用Pipeline保存配置和版本历史
	pipe := s.redis.Pipeline()
	
	// 保存当前配置
	pipe.Set(ctx, s.getConfigKey(key), data, 0)
	
	// 保存历史版本
	historyKey := s.getHistoryKey(key, item.Version)
	pipe.Set(ctx, historyKey, data, 0)
	
	// 更新版本号
	pipe.Set(ctx, s.getVersionKey(key), item.Version, 0)

	// 执行Pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	// 发布配置变更事件
	event := map[string]interface{}{
		"type": "config_updated",
		"data": item,
	}
	eventData, _ := json.Marshal(event)
	s.redis.Publish(ctx, "config_changes", eventData)

	return nil
}

// GetConfig 获取配置
func (s *Service) GetConfig(ctx context.Context, key string) (*ConfigItem, error) {
	data, err := s.redis.Get(ctx, s.getConfigKey(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("配置不存在: %s", key)
		}
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	var item ConfigItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &item, nil
}

// GetConfigHistory 获取配置历史版本
func (s *Service) GetConfigHistory(ctx context.Context, key string, version int64) (*ConfigItem, error) {
	data, err := s.redis.Get(ctx, s.getHistoryKey(key, version)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("配置版本不存在: %s@%d", key, version)
		}
		return nil, fmt.Errorf("获取配置历史失败: %w", err)
	}

	var item ConfigItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &item, nil
}

// DeleteConfig 删除配置
func (s *Service) DeleteConfig(ctx context.Context, key string) error {
	// 获取当前版本
	version, err := s.getCurrentVersion(ctx, key)
	if err != nil {
		return err
	}

	// 使用Pipeline删除配置和版本历史
	pipe := s.redis.Pipeline()
	
	// 删除当前配置
	pipe.Del(ctx, s.getConfigKey(key))
	
	// 删除版本号
	pipe.Del(ctx, s.getVersionKey(key))
	
	// 删除所有历史版本
	for v := int64(1); v <= version; v++ {
		pipe.Del(ctx, s.getHistoryKey(key, v))
	}

	// 执行Pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("删除配置失败: %w", err)
	}

	// 发布配置删除事件
	event := map[string]interface{}{
		"type": "config_deleted",
		"data": map[string]string{"key": key},
	}
	eventData, _ := json.Marshal(event)
	s.redis.Publish(ctx, "config_changes", eventData)

	return nil
}

// ListConfigs 列出所有配置
func (s *Service) ListConfigs(ctx context.Context) ([]*ConfigItem, error) {
	// 获取所有配置键
	pattern := s.getConfigKey("*")
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("获取配置列表失败: %w", err)
	}

	// 获取所有配置
	var configs []*ConfigItem
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			s.logger.Error("获取配置失败", "key", key, "error", err)
			continue
		}

		var item ConfigItem
		if err := json.Unmarshal(data, &item); err != nil {
			s.logger.Error("解析配置失败", "key", key, "error", err)
			continue
		}

		configs = append(configs, &item)
	}

	return configs, nil
}

// 内部方法

func (s *Service) getCurrentVersion(ctx context.Context, key string) (int64, error) {
	version, err := s.redis.Get(ctx, s.getVersionKey(key)).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("获取版本号失败: %w", err)
	}
	return version, nil
}

func (s *Service) getConfigKey(key string) string {
	return fmt.Sprintf("config:%s", key)
}

func (s *Service) getVersionKey(key string) string {
	return fmt.Sprintf("config:%s:version", key)
}

func (s *Service) getHistoryKey(key string, version int64) string {
	return fmt.Sprintf("config:%s:history:%d", key, version)
} 