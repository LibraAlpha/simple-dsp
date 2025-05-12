package budget

import "errors"

var (
	// ErrBudgetNotFound 表示预算不存在
	ErrBudgetNotFound = errors.New("预算不存在")

	// ErrBudgetAlreadyExists 表示预算已存在
	ErrBudgetAlreadyExists = errors.New("预算已存在")

	// ErrBudgetExceeded 表示预算已超限
	ErrBudgetExceeded = errors.New("预算已超限")

	// ErrBudgetInactive 表示预算未激活
	ErrBudgetInactive = errors.New("预算未激活")

	// ErrBudgetExpired 表示预算已过期
	ErrBudgetExpired = errors.New("预算已过期")

	// ErrInvalidBudgetAmount 表示无效的预算金额
	ErrInvalidBudgetAmount = errors.New("无效的预算金额")

	// ErrInvalidBudgetTime 表示无效的预算时间
	ErrInvalidBudgetTime = errors.New("无效的预算时间")

	// ErrInvalidBudgetType 表示无效的预算类型
	ErrInvalidBudgetType = errors.New("无效的预算类型")

	// ErrRedisOperation 表示Redis操作失败
	ErrRedisOperation = errors.New("Redis操作失败")
) 