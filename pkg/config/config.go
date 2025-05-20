/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: config.go
 * Project: simple-dsp
 * Description: 配置管理模块，负责系统配置的加载和管理
 *
 * 主要功能:
 * - 加载系统配置
 * - 管理动态配置
 * - 提供配置访问接口
 * - 支持配置热更新
 *
 * 实现细节:
 * - 支持YAML配置文件
 * - 实现配置验证
 * - 支持环境变量覆盖
 * - 提供配置监听机制
 *
 * 依赖关系:
 * - gopkg.in/yaml.v3
 * - simple-dsp/pkg/clients
 * - simple-dsp/pkg/logger
 *
 * 注意事项:
 * - 注意配置的安全性
 * - 合理设置默认值
 * - 注意配置变更通知
 * - 确保配置一致性
 */

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Traffic  TrafficConfig  `mapstructure:"traffic"`
	RTA      RTAConfig      `mapstructure:"rta"`
	Bidding  BiddingConfig  `mapstructure:"bidding"`
	Budget   BudgetConfig   `mapstructure:"budget"`
	Stats    StatsConfig    `mapstructure:"stats"`
	Event    EventConfig    `mapstructure:"event"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Log      LogConfig      `mapstructure:"log"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// TrafficConfig 流量接入配置
type TrafficConfig struct {
	QPS           float64       `mapstructure:"qps"`
	Burst         int           `mapstructure:"burst"`
	RTATimeout    time.Duration `mapstructure:"rta_timeout"`
	BidTimeout    time.Duration `mapstructure:"bid_timeout"`
	MaxAdSlots    int           `mapstructure:"max_ad_slots"`
	MinAdSlotSize int           `mapstructure:"min_ad_slot_size"`
	MaxAdSlotSize int           `mapstructure:"max_ad_slot_size"`
}

// RTAConfig RTA服务配置
type RTAConfig struct {
	BaseURL    string        `mapstructure:"base_url"`
	AppKey     string        `mapstructure:"app_key"`
	AppSecret  string        `mapstructure:"app_secret"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryTimes int           `mapstructure:"retry_times"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
	CacheTTL   time.Duration `mapstructure:"cache_ttl"`
	BatchSize  int           `mapstructure:"batch_size"`
}

// BiddingConfig 竞价服务配置
type BiddingConfig struct {
	MaxConcurrentBids int           `mapstructure:"max_concurrent_bids"`
	BidTimeout        time.Duration `mapstructure:"bid_timeout"`
	MinBidPrice       float64       `mapstructure:"min_bid_price"`
	MaxBidPrice       float64       `mapstructure:"max_bid_price"`
	CTRModelPath      string        `mapstructure:"ctr_model_path"`
}

// BudgetConfig 预算管理配置
type BudgetConfig struct {
	CheckInterval    time.Duration `mapstructure:"check_interval"`
	WarningThreshold float64       `mapstructure:"warning_threshold"`
	AutoRenewal      bool          `mapstructure:"auto_renewal"`
	RenewalTime      string        `mapstructure:"renewal_time"`
}

// StatsConfig 数据统计配置
type StatsConfig struct {
	KafkaTopics struct {
		Impression string `mapstructure:"impression"`
		Click      string `mapstructure:"click"`
		Conversion string `mapstructure:"conversion"`
	} `mapstructure:"kafka_topics"`
	RedisPrefix   string        `mapstructure:"redis_prefix"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	RetentionDays int           `mapstructure:"retention_days"`
}

// EventConfig 事件处理配置
type EventConfig struct {
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
	ProcessTimeout time.Duration `mapstructure:"process_timeout"`
	QueueSize      int           `mapstructure:"queue_size"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addresses    []string      `mapstructure:"addresses"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type ClusterNode struct {
	ID    string
	Host  string
	Port  int
	Role  string // "master" 或 "slave"
	Slots []int  // 负责的哈希槽范围[2](@ref)
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers      []string      `mapstructure:"brokers"`
	Topic        string        `mapstructure:"topic"`
	GroupID      string        `mapstructure:"group_id"`
	Version      string        `mapstructure:"version"`
	MaxRetries   int           `mapstructure:"max_retries"`
	RetryBackoff time.Duration `mapstructure:"retry_backoff"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// MetricsConfig 监控指标配置
type MetricsConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Port        int    `mapstructure:"port"`
	Path        string `mapstructure:"path"`
	PushGateway string `mapstructure:"push_gateway"`
	HTTPEnabled bool   `mapstructure:"http_enabled"`
}

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

var (
	// GlobalConfig 全局配置实例
	GlobalConfig Config
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return validateConfig(&GlobalConfig)
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Port <= 0 {
		return fmt.Errorf("无效的服务器端口: %d", cfg.Server.Port)
	}

	// 验证流量配置
	if cfg.Traffic.QPS <= 0 {
		return fmt.Errorf("无效的QPS限制: %f", cfg.Traffic.QPS)
	}
	if cfg.Traffic.Burst <= 0 {
		return fmt.Errorf("无效的突发请求限制: %d", cfg.Traffic.Burst)
	}

	// 验证RTA配置
	if cfg.RTA.BaseURL == "" {
		return fmt.Errorf("RTA服务地址不能为空")
	}
	if cfg.RTA.Timeout <= 0 {
		return fmt.Errorf("无效的RTA超时时间: %v", cfg.RTA.Timeout)
	}

	// 验证Redis配置
	if len(cfg.Redis.Addresses) == 0 {
		return fmt.Errorf("Redis地址不能为空")
	}

	// 验证Kafka配置
	if len(cfg.Kafka.Brokers) == 0 {
		return fmt.Errorf("Kafka代理地址不能为空")
	}

	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return &GlobalConfig
}

// DynamicConfig 动态配置管理器
type DynamicConfig struct {
	redis   *redis.Client
	configs map[string]interface{}
	mu      sync.RWMutex
}

// NewDynamicConfig 创建动态配置管理器
func NewDynamicConfig(redis *redis.Client) *DynamicConfig {
	dc := &DynamicConfig{
		redis:   redis,
		configs: make(map[string]interface{}),
	}

	// 启动配置监听
	go dc.watchConfigChanges()
	return dc
}

// Get 获取配置值
func (dc *DynamicConfig) Get(key string) interface{} {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if value, ok := dc.configs[key]; ok {
		return value
	}

	// 如果内存中没有，尝试从Redis获取
	if value, err := dc.loadFromRedis(key); err == nil {
		dc.configs[key] = value
		return value
	}

	// 返回默认配置
	return dc.getDefaultConfig(key)
}

// Set 设置配置值
func (dc *DynamicConfig) Set(key string, value interface{}) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// 保存到Redis
	if err := dc.saveToRedis(key, value); err != nil {
		return err
	}

	// 更新内存中的值
	dc.configs[key] = value
	return nil
}

// watchConfigChanges 监听配置变更
func (dc *DynamicConfig) watchConfigChanges() {
	pubsub := dc.redis.Subscribe(context.Background(), "config_changes")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var event struct {
			Type string                 `json:"type"`
			Data map[string]interface{} `json:"data"`
		}

		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			continue
		}

		dc.mu.Lock()
		switch event.Type {
		case "config_updated":
			if item, ok := event.Data["key"].(string); ok {
				if value, err := dc.loadFromRedis(item); err == nil {
					dc.configs[item] = value
				}
			}
		case "config_deleted":
			if item, ok := event.Data["key"].(string); ok {
				delete(dc.configs, item)
			}
		}
		dc.mu.Unlock()
	}
}

// loadFromRedis 从Redis加载配置
func (dc *DynamicConfig) loadFromRedis(key string) (interface{}, error) {
	data, err := dc.redis.Get(context.Background(), "config:"+key).Bytes()
	if err != nil {
		return nil, err
	}

	var item struct {
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	return item.Value, nil
}

// saveToRedis 保存配置到Redis
func (dc *DynamicConfig) saveToRedis(key string, value interface{}) error {
	data, err := json.Marshal(map[string]interface{}{
		"key":   key,
		"value": value,
	})
	if err != nil {
		return err
	}

	return dc.redis.Set(context.Background(), "config:"+key, data, 0).Err()
}

// getDefaultConfig 获取默认配置
func (dc *DynamicConfig) getDefaultConfig(key string) interface{} {
	switch key {
	case "server.port":
		return 8080
	case "server.read_timeout":
		return time.Second * 5
	case "server.write_timeout":
		return time.Second * 10
	case "redis.pool_size":
		return 100
	case "kafka.batch_size":
		return 100
	case "traffic.qps":
		return 1000.0
	case "bidding.max_concurrent_bids":
		return 100
	default:
		return nil
	}
}

// Service 配置服务
type Service struct {
	config        *Config
	dynamicConfig *DynamicConfig
	redis         *redis.Client
	mu            sync.RWMutex
}

// NewService 创建配置服务实例
func NewService(redis *redis.Client) *Service {
	return &Service{
		config:        &GlobalConfig,
		dynamicConfig: NewDynamicConfig(redis),
		redis:         redis,
	}
}

// GetConfig 获取全局配置
func (s *Service) GetConfig() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetDynamicConfig 获取动态配置管理器
func (s *Service) GetDynamicConfig() *DynamicConfig {
	return s.dynamicConfig
}

// UpdateConfig 更新配置
func (s *Service) UpdateConfig(newConfig *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateConfig(newConfig); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	s.config = newConfig
	return nil
}

// GetRedis 获取Redis客户端
func (s *Service) GetRedis() *redis.Client {
	return s.redis
}

// GetMetrics 获取监控指标配置
func (s *Service) GetMetrics() *MetricsConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Metrics
}

// GetRTA 获取RTA配置
func (s *Service) GetRTA() *RTAConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.RTA
}

// GetTraffic 获取流量配置
func (s *Service) GetTraffic() *TrafficConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Traffic
}

// GetBidding 获取竞价配置
func (s *Service) GetBidding() *BiddingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Bidding
}

// GetBudget 获取预算配置
func (s *Service) GetBudget() *BudgetConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Budget
}

// GetStats 获取统计配置
func (s *Service) GetStats() *StatsConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Stats
}

// GetEvent 获取事件配置
func (s *Service) GetEvent() *EventConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Event
}

// GetKafka 获取Kafka配置
func (s *Service) GetKafka() *KafkaConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Kafka
}

// GetLog 获取日志配置
func (s *Service) GetLog() *LogConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Log
}

// GetPostgres 获取PostgreSQL配置
func (s *Service) GetPostgres() *PostgresConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Postgres
}
