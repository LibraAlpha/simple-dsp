package rta

import (
    "sync"
    "time"
)

// TaskConfig RTA任务配置
type TaskConfig struct {
    TaskID            string        `json:"task_id"`             // 任务ID
    Channel           string        `json:"channel"`             // 渠道ID
    AdvertisingSpaceID string       `json:"advertising_space_id"` // 广告位ID
    Timeout           time.Duration `json:"timeout"`             // 超时时间
    Enabled           bool          `json:"enabled"`             // 是否启用
    Priority          int           `json:"priority"`            // 优先级
    RetryCount        int           `json:"retry_count"`         // 重试次数
    RetryInterval     time.Duration `json:"retry_interval"`      // 重试间隔
    CacheExpiration   time.Duration `json:"cache_expiration"`    // 缓存过期时间
}

// ConfigManager RTA配置管理器
type ConfigManager struct {
    configs map[string]*TaskConfig  // 任务配置映射
    mu      sync.RWMutex           // 读写锁
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
    return &ConfigManager{
        configs: make(map[string]*TaskConfig),
    }
}

// SetConfig 设置任务配置
func (m *ConfigManager) SetConfig(config *TaskConfig) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.configs[config.TaskID] = config
}

// GetConfig 获取任务配置
func (m *ConfigManager) GetConfig(taskID string) (*TaskConfig, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    config, exists := m.configs[taskID]
    return config, exists
}

// RemoveConfig 移除任务配置
func (m *ConfigManager) RemoveConfig(taskID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.configs, taskID)
}

// ListConfigs 列出所有任务配置
func (m *ConfigManager) ListConfigs() []*TaskConfig {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    configs := make([]*TaskConfig, 0, len(m.configs))
    for _, config := range m.configs {
        configs = append(configs, config)
    }
    return configs
}

// DefaultConfig 默认配置
var DefaultConfig = &TaskConfig{
    Timeout:         time.Second * 100,  // 默认100ms超时
    Enabled:         true,
    Priority:        1,
    RetryCount:      2,
    RetryInterval:   time.Millisecond * 50,
    CacheExpiration: time.Minute * 5,
} 