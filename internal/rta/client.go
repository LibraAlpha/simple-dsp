package rta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// Client RTA服务客户端
type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *logger.Logger
	metrics    *metrics.Metrics
}

// NewClient 创建新的RTA客户端
func NewClient(baseURL string, logger *logger.Logger, metrics *metrics.Metrics) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 100 * time.Millisecond,
		},
		baseURL: baseURL,
		logger:  logger,
		metrics: metrics,
	}
}

// CheckTargeting 检查用户是否符合RTA定向要求
func (c *Client) CheckTargeting(ctx context.Context, userID string) (bool, error) {
	startTime := time.Now()
	defer func() {
		c.metrics.RTACheckDuration.Observe(time.Since(startTime).Seconds())
	}()

	// 构造请求URL
	url := fmt.Sprintf("%s/api/v1/rta/check?user_id=%s", c.baseURL, userID)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("创建RTA请求失败", "error", err)
		return false, err
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("RTA请求失败", "error", err)
		return false, err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("RTA服务返回错误状态码", "status_code", resp.StatusCode)
		return false, fmt.Errorf("RTA服务返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var result struct {
		Code    int  `json:"code"`
		Message string `json:"message"`
		Data    struct {
			IsTargeted bool `json:"is_targeted"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("解析RTA响应失败", "error", err)
		return false, err
	}

	// 检查业务状态码
	if result.Code != 0 {
		c.logger.Error("RTA服务返回业务错误", "code", result.Code, "message", result.Message)
		return false, fmt.Errorf("RTA服务返回业务错误: %s", result.Message)
	}

	return result.Data.IsTargeted, nil
}

// BatchCheckTargeting 批量检查用户是否符合RTA定向要求
func (c *Client) BatchCheckTargeting(ctx context.Context, userIDs []string) (map[string]bool, error) {
	startTime := time.Now()
	defer func() {
		c.metrics.RTABatchCheckDuration.Observe(time.Since(startTime).Seconds())
	}()

	// 构造请求体
	reqBody := struct {
		UserIDs []string `json:"user_ids"`
	}{
		UserIDs: userIDs,
	}

	// 序列化请求体
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("序列化RTA批量请求失败", "error", err)
		return nil, err
	}

	// 构造请求URL
	url := fmt.Sprintf("%s/api/v1/rta/batch_check", c.baseURL)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		c.logger.Error("创建RTA批量请求失败", "error", err)
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("RTA批量请求失败", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("RTA服务返回错误状态码", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("RTA服务返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Results map[string]bool `json:"results"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("解析RTA批量响应失败", "error", err)
		return nil, err
	}

	// 检查业务状态码
	if result.Code != 0 {
		c.logger.Error("RTA服务返回业务错误", "code", result.Code, "message", result.Message)
		return nil, fmt.Errorf("RTA服务返回业务错误: %s", result.Message)
	}

	return result.Data.Results, nil
}
