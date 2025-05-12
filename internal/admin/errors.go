package admin

import "errors"

var (
	// 广告相关错误
	ErrInvalidAdID          = errors.New("无效的广告ID")
	ErrInvalidAdTitle       = errors.New("无效的广告标题")
	ErrInvalidAdDescription = errors.New("无效的广告描述")
	ErrInvalidAdImageURL    = errors.New("无效的广告图片URL")
	ErrInvalidAdLandingURL  = errors.New("无效的广告落地页URL")
	ErrInvalidAdSize        = errors.New("无效的广告尺寸")
	ErrInvalidAdStatus      = errors.New("无效的广告状态")
	ErrAdNotFound           = errors.New("广告不存在")
	ErrAdAlreadyExists      = errors.New("广告已存在")
	ErrAdDeleted            = errors.New("广告已删除")

	// 预算相关错误
	ErrInvalidBudgetID      = errors.New("无效的预算ID")
	ErrInvalidBudgetName    = errors.New("无效的预算名称")
	ErrInvalidBudgetAmount  = errors.New("无效的预算金额")
	ErrInvalidBudgetTime    = errors.New("无效的预算时间")
	ErrInvalidBudgetStatus  = errors.New("无效的预算状态")
	ErrBudgetNotFound       = errors.New("预算不存在")
	ErrBudgetAlreadyExists  = errors.New("预算已存在")
	ErrBudgetExceeded       = errors.New("预算已超支")
	ErrBudgetExpired        = errors.New("预算已过期")
	ErrBudgetRenewalFailed  = errors.New("预算续费失败")

	// 统计相关错误
	ErrInvalidStatsTimeRange = errors.New("无效的统计时间范围")
	ErrStatsNotFound         = errors.New("统计数据不存在")
	ErrStatsCalculationFailed = errors.New("统计计算失败")

	// 系统相关错误
	ErrRedisConnectionFailed = errors.New("Redis连接失败")
	ErrMetricsCollectionFailed = errors.New("指标收集失败")
	ErrSystemUnavailable     = errors.New("系统不可用")

	// 通用错误
	ErrInvalidRequest       = errors.New("无效的请求")
	ErrInvalidResponse      = errors.New("无效的响应")
	ErrInternalServer       = errors.New("服务器内部错误")
	ErrUnauthorized         = errors.New("未授权访问")
	ErrForbidden            = errors.New("禁止访问")
	ErrServiceUnavailable   = errors.New("服务不可用")
	ErrRequestTimeout       = errors.New("请求超时")
	ErrTooManyRequests      = errors.New("请求过于频繁")
) 