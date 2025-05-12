package bidding

import "errors"

var (
	// ErrInvalidBidRequest 表示无效的竞价请求
	ErrInvalidBidRequest = errors.New("无效的竞价请求")

	// ErrNoAvailableAds 表示没有可用的广告
	ErrNoAvailableAds = errors.New("没有可用的广告")

	// ErrBudgetExceeded 表示预算超限
	ErrBudgetExceeded = errors.New("预算已超限")

	// ErrInvalidAdSlot 表示无效的广告位
	ErrInvalidAdSlot = errors.New("无效的广告位")

	// ErrInvalidBidPrice 表示无效的竞价价格
	ErrInvalidBidPrice = errors.New("无效的竞价价格")

	// ErrAdNotFound 表示广告不存在
	ErrAdNotFound = errors.New("广告不存在")

	// ErrAdInactive 表示广告未激活
	ErrAdInactive = errors.New("广告未激活")

	// ErrCTRPredictionFailed 表示CTR预测失败
	ErrCTRPredictionFailed = errors.New("CTR预测失败")

	// ErrECPMCalculationFailed 表示eCPM计算失败
	ErrECPMCalculationFailed = errors.New("eCPM计算失败")
) 