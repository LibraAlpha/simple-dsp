package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"simple-dsp/internal/campaign"
)

// Campaign 广告计划数据库模型
type Campaign struct {
	ID              string    `gorm:"column:id;primary_key"`
	Name            string    `gorm:"column:name"`
	AdvertiserID    string    `gorm:"column:advertiser_id"`
	Status          string    `gorm:"column:status"`
	StartTime       time.Time `gorm:"column:start_time"`
	EndTime         time.Time `gorm:"column:end_time"`
	Budget          float64   `gorm:"column:budget"`
	BidStrategy     string    `gorm:"column:bid_strategy"`
	Targeting       JSON      `gorm:"column:targeting"`
	TrackingConfigs JSON      `gorm:"column:tracking_configs"`
	UpdateTime      time.Time `gorm:"column:update_time"`
	CreateTime      time.Time `gorm:"column:create_time"`
}

// TableName 返回表名
func (Campaign) TableName() string {
	return "campaigns"
}

// JSON 自定义JSON类型
type JSON []byte

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid scan source")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

// IsNull 检查是否为空
func (j JSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

// ToCampaignConfig 转换为广告计划配置
func (c *Campaign) ToCampaignConfig() (*campaign.Config, error) {
	config := &campaign.Config{
		CampaignID:   c.ID,
		Name:         c.Name,
		AdvertiserID: c.AdvertiserID,
		Status:       c.Status,
		StartTime:    c.StartTime,
		EndTime:      c.EndTime,
		Budget:       c.Budget,
		BidStrategy:  c.BidStrategy,
		UpdateTime:   c.UpdateTime,
		CreateTime:   c.CreateTime,
	}

	// 解析定向配置
	if !c.Targeting.IsNull() {
		var targeting campaign.TargetingConfig
		if err := json.Unmarshal(c.Targeting, &targeting); err != nil {
			return nil, err
		}
		config.Targeting = &targeting
	}

	// 解析跟踪配置
	if !c.TrackingConfigs.IsNull() {
		var trackingConfigs map[campaign.TrackingType]*campaign.TrackingConfig
		if err := json.Unmarshal(c.TrackingConfigs, &trackingConfigs); err != nil {
			return nil, err
		}
		config.TrackingConfigs = trackingConfigs
	}

	return config, nil
}

// FromCampaignConfig 从广告计划配置转换
func (c *Campaign) FromCampaignConfig(config *campaign.Config) error {
	c.ID = config.CampaignID
	c.Name = config.Name
	c.AdvertiserID = config.AdvertiserID
	c.Status = config.Status
	c.StartTime = config.StartTime
	c.EndTime = config.EndTime
	c.Budget = config.Budget
	c.BidStrategy = config.BidStrategy
	c.UpdateTime = config.UpdateTime
	c.CreateTime = config.CreateTime

	// 序列化定向配置
	if config.Targeting != nil {
		targeting, err := json.Marshal(config.Targeting)
		if err != nil {
			return err
		}
		c.Targeting = targeting
	}

	// 序列化跟踪配置
	if config.TrackingConfigs != nil {
		trackingConfigs, err := json.Marshal(config.TrackingConfigs)
		if err != nil {
			return err
		}
		c.TrackingConfigs = trackingConfigs
	}

	return nil
}
