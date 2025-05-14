package campaign

import (
	"fmt"
	"sync"
	"time"
)

// TrackingType 跟踪类型
type TrackingType string

const (
	TrackingTypeClick      TrackingType = "click"      // 点击跟踪
	TrackingTypeImpression TrackingType = "impression" // 曝光跟踪
	TrackingTypeDP         TrackingType = "dp"         // DP跟踪
)

// TrackingConfig 跟踪配置
type TrackingConfig struct {
	URL           string            `json:"url"`            // 跟踪URL
	Method        string            `json:"method"`         // HTTP方法
	Headers       map[string]string `json:"headers"`        // 自定义请求头
	Timeout       time.Duration     `json:"timeout"`        // 超时时间
	RetryCount    int               `json:"retry_count"`    // 重试次数
	RetryInterval time.Duration     `json:"retry_interval"` // 重试间隔
	Enabled       bool              `json:"enabled"`        // 是否启用
}

// Config CampaignConfig 广告计划配置
type Config struct {
	CampaignID      string                           `json:"campaign_id"`      // 广告计划ID
	Name            string                           `json:"name"`             // 计划名称
	AdvertiserID    string                           `json:"advertiser_id"`    // 广告主ID
	Status          string                           `json:"status"`           // 状态
	StartTime       time.Time                        `json:"start_time"`       // 开始时间
	EndTime         time.Time                        `json:"end_time"`         // 结束时间
	Budget          float64                          `json:"budget"`           // 预算
	BidStrategy     string                           `json:"bid_strategy"`     // 出价策略
	Targeting       *TargetingConfig                 `json:"targeting"`        // 定向配置
	TrackingConfigs map[TrackingType]*TrackingConfig `json:"tracking_configs"` // 跟踪配置
	UpdateTime      time.Time                        `json:"update_time"`      // 更新时间
	CreateTime      time.Time                        `json:"create_time"`      // 创建时间
}

// TargetingConfig 定向配置
type TargetingConfig struct {
	Locations    []string          `json:"locations"`     // 地域定向
	Ages         []string          `json:"ages"`          // 年龄定向
	Genders      []string          `json:"genders"`       // 性别定向
	Interests    []string          `json:"interests"`     // 兴趣定向
	OSTypes      []string          `json:"os_types"`      // 操作系统定向
	NetworkTypes []string          `json:"network_types"` // 网络类型定向
	CustomRules  map[string]string `json:"custom_rules"`  // 自定义规则
}

// ConfigManager 配置管理器
type ConfigManager struct {
	configs map[string]*Config // 计划配置映射
	mu      sync.RWMutex       // 读写锁
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs: make(map[string]*Config),
	}
}

// SetConfig 设置计划配置
func (m *ConfigManager) SetConfig(config *Config) error {
	if err := validateConfig(config); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	config.UpdateTime = time.Now()
	if _, exists := m.configs[config.CampaignID]; !exists {
		config.CreateTime = config.UpdateTime
	}

	m.configs[config.CampaignID] = config
	return nil
}

// GetConfig 获取计划配置
func (m *ConfigManager) GetConfig(campaignID string) (*Config, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, exists := m.configs[campaignID]
	return config, exists
}

// RemoveConfig 移除计划配置
func (m *ConfigManager) RemoveConfig(campaignID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.configs, campaignID)
}

// ListConfigs 列出所有计划配置
func (m *ConfigManager) ListConfigs() []*Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configs := make([]*Config, 0, len(m.configs))
	for _, config := range m.configs {
		configs = append(configs, config)
	}
	return configs
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.CampaignID == "" {
		return fmt.Errorf("campaign_id is required")
	}
	if config.AdvertiserID == "" {
		return fmt.Errorf("advertiser_id is required")
	}

	// 验证跟踪配置
	for trackingType, trackingConfig := range config.TrackingConfigs {
		if trackingConfig.Enabled {
			if trackingConfig.URL == "" {
				return fmt.Errorf("%s tracking URL is required", trackingType)
			}
			if trackingConfig.Timeout <= 0 {
				trackingConfig.Timeout = time.Second * 1 // 默认1秒超时
			}
			if trackingConfig.RetryCount < 0 {
				trackingConfig.RetryCount = 0
			}
			if trackingConfig.RetryInterval <= 0 {
				trackingConfig.RetryInterval = time.Millisecond * 100 // 默认100ms重试间隔
			}
		}
	}

	return nil
}
