package admin

import (
	"net/url"
	"regexp"
	"simple-dsp/internal/bidding"
	"strings"
	"time"
)

// Validator 验证器接口
type Validator interface {
	ValidateAd(ad *Ad) error
	ValidateBudget(budget *Budget) error
	ValidateTimeRange(start, end time.Time) error
}

// validator 验证器实现
type validator struct{}

// NewValidator 创建验证器
func NewValidator() Validator {
	return &validator{}
}

// ValidateAd 验证广告信息
func (v *validator) ValidateAd(ad *Ad) error {
	// 验证广告ID
	if ad.ID != "" && !isValidID(ad.ID) {
		return ErrInvalidAdID
	}

	// 验证广告标题
	if ad.Title == "" || len(ad.Title) > 100 {
		return ErrInvalidAdTitle
	}

	// 验证广告描述
	if ad.Description != "" && len(ad.Description) > 500 {
		return ErrInvalidAdDescription
	}

	// 验证广告图片URL
	if ad.ImageURL != "" && !isValidURL(ad.ImageURL) {
		return ErrInvalidAdImageURL
	}

	// 验证广告落地页URL
	if ad.LandingURL != "" && !isValidURL(ad.LandingURL) {
		return ErrInvalidAdLandingURL
	}

	// 验证广告尺寸
	if ad.Width <= 0 || ad.Height <= 0 {
		return ErrInvalidAdSize
	}

	// 验证广告状态
	if ad.Status != "" && !isValidAdStatus(ad.Status) {
		return ErrInvalidAdStatus
	}

	return nil
}

// ValidateBudget 验证预算信息
func (v *validator) ValidateBudget(budget *Budget) error {
	// 验证预算ID
	if budget.ID != "" && !isValidID(budget.ID) {
		return ErrInvalidBudgetID
	}

	// 验证预算名称
	if budget.Name == "" || len(budget.Name) > 50 {
		return ErrInvalidBudgetName
	}

	// 验证预算金额
	if budget.Amount <= 0 {
		return ErrInvalidBudgetAmount
	}

	// 验证预算时间
	if budget.StartTime.IsZero() || budget.EndTime.IsZero() {
		return ErrInvalidBudgetTime
	}
	if budget.StartTime.After(budget.EndTime) {
		return ErrInvalidBudgetTime
	}

	// 验证预算状态
	if budget.Status != "" && !isValidBudgetStatus(budget.Status) {
		return ErrInvalidBudgetStatus
	}

	return nil
}

// ValidateTimeRange 验证时间范围
func (v *validator) ValidateTimeRange(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return ErrInvalidStatsTimeRange
	}
	if start.After(end) {
		return ErrInvalidStatsTimeRange
	}
	if end.Sub(start) > 30*24*time.Hour { // 最多查询30天
		return ErrInvalidStatsTimeRange
	}
	return nil
}

// 辅助验证函数

func isValidID(id string) bool {
	// 验证ID格式：时间戳(14位) + 随机字符串(6位)
	pattern := `^\d{14}[a-zA-Z0-9]{6}$`
	matched, _ := regexp.MatchString(pattern, id)
	return matched
}

func isValidURL(urlStr string) bool {
	// 验证URL格式
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func isValidAdStatus(status string) bool {
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
		"deleted":  true,
	}
	return validStatuses[status]
}

func isValidBudgetStatus(status string) bool {
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
		"expired":  true,
	}
	return validStatuses[status]
}

// 验证请求参数

func validateAdSlot(slot *bidding.AdSlot) error {
	if slot == nil {
		return ErrInvalidRequest
	}
	if slot.SlotID == "" {
		return ErrInvalidRequest
	}
	if slot.Width <= 0 || slot.Height <= 0 {
		return ErrInvalidRequest
	}
	if slot.MinPrice < 0 || slot.MaxPrice < 0 {
		return ErrInvalidRequest
	}
	return nil
}

func validateBidRequest(req *bidding.BidRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}
	if req.RequestID == "" {
		return ErrInvalidRequest
	}
	if req.UserID == "" {
		return ErrInvalidRequest
	}
	if req.DeviceID == "" {
		return ErrInvalidRequest
	}
	if req.IP == "" || !isValidIP(req.IP) {
		return ErrInvalidRequest
	}
	if len(req.AdSlots) == 0 {
		return ErrInvalidRequest
	}
	for _, slot := range req.AdSlots {
		if err := validateAdSlot(&slot); err != nil {
			return err
		}
	}
	return nil
}

func isValidIP(ip string) bool {
	// 简单验证IP格式
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		if len(part) > 3 {
			return false
		}
		if part[0] == '0' && len(part) > 1 {
			return false
		}
	}
	return true
}
