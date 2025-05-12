package config

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"simple-dsp/pkg/logger"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	redis      *redis.Client
	logger     *logger.Logger
	configs    map[string]interface{}
	watchers   map[string][]chan interface{}
	mu         sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewConfigManager 创建配置管理器
func NewConfigManager(redis *redis.Client, logger *logger.Logger) *ConfigManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConfigManager{
		redis:      redis,
		logger:     logger,
		configs:    make(map[string]interface{}),
		watchers:   make(map[string][]chan interface{}),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// Watch 监听配置变更
func (cm *ConfigManager) Watch(key string, callback chan interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if _, exists := cm.watchers[key]; !exists {
		cm.watchers[key] = make([]chan interface{}, 0)
	}
	cm.watchers[key] = append(cm.watchers[key], callback)
}

// StartWatch 开始监听配置变更
func (cm *ConfigManager) StartWatch() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-cm.ctx.Done():
				return
			case <-ticker.C:
				cm.checkConfigUpdates()
			}
		}
	}()
}

// Stop 停止配置监听
func (cm *ConfigManager) Stop() {
	cm.cancelFunc()
}

// checkConfigUpdates 检查配置更新
func (cm *ConfigManager) checkConfigUpdates() {
	cm.mu.RLock()
	keys := make([]string, 0, len(cm.watchers))
	for k := range cm.watchers {
		keys = append(keys, k)
	}
	cm.mu.RUnlock()

	for _, key := range keys {
		value, err := cm.redis.Get(cm.ctx, "config:"+key).Bytes()
		if err != nil {
			if err != redis.Nil {
				cm.logger.Error("获取配置失败", "key", key, "error", err)
			}
			continue
		}

		var newConfig interface{}
		if err := json.Unmarshal(value, &newConfig); err != nil {
			cm.logger.Error("解析配置失败", "key", key, "error", err)
			continue
		}

		cm.mu.Lock()
		oldConfig, exists := cm.configs[key]
		if !exists || !jsonEqual(oldConfig, newConfig) {
			cm.configs[key] = newConfig
			// 通知所有监听器
			for _, watcher := range cm.watchers[key] {
				select {
				case watcher <- newConfig:
				default:
					cm.logger.Warn("配置通知队列已满", "key", key)
				}
			}
		}
		cm.mu.Unlock()
	}
}

// jsonEqual 比较两个JSON对象是否相等
func jsonEqual(a, b interface{}) bool {
	aJson, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJson, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aJson) == string(bJson)
} 