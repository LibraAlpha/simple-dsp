package frequency

import "errors"

var (
	// ErrInvalidUserID 无效的用户ID
	ErrInvalidUserID = errors.New("无效的用户ID")

	// ErrInvalidAdID 无效的广告ID
	ErrInvalidAdID = errors.New("无效的广告ID")

	// ErrInvalidConfig 无效的配置
	ErrInvalidConfig = errors.New("无效的配置")

	// ErrImpressionLimitExceeded 曝光频次超限
	ErrImpressionLimitExceeded = errors.New("曝光频次超限")

	// ErrClickLimitExceeded 点击频次超限
	ErrClickLimitExceeded = errors.New("点击频次超限")

	// ErrQPSLimitExceeded QPS超限
	ErrQPSLimitExceeded = errors.New("QPS超限")

	// ErrRedisOperationFailed Redis操作失败
	ErrRedisOperationFailed = errors.New("Redis操作失败")

	// ErrConfigNotFound 配置不存在
	ErrConfigNotFound = errors.New("配置不存在")
) 