package rta

import (
	"time"

	"go.uber.org/zap"
)

// EvaluateUser 实时用户价值评估
func (c *Client) EvaluateUser(deviceID string) (bool, float64) {
	req := RTARequest{
		DeviceID:  deviceID,
		Timestamp: time.Now().Unix(),
	}

	resp, err := c.postRTA(req)
	if err != nil {
		c.logger.Error("RTA请求失败", zap.Error(err))
		return false, 0
	}

	// 根据RTA响应调整出价
	if resp.Participate {
		adjustedBid := resp.BaseBid * resp.BidMultiplier
		return true, adjustedBid
	}
	return false, 0
}
