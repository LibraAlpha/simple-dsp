package rta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

const (
	singleAPIURL = "https://open.taobao.com/api.htm?docId=48589"
	batchAPIURL  = "https://open.taobao.com/api.htm?docId=48588"
)

// Client RTA客户端
type Client struct {
	appKey       string
	appSecret    string
	httpClient   *http.Client
	logger       *logger.Logger
	metrics      *metrics.Metrics
	configMgr    *ConfigManager
	cache        *cache.Cache
	defaultTimeout time.Duration
}

// NewClient 创建新的RTA客户端
func NewClient(appKey, appSecret string, logger *logger.Logger, metrics *metrics.Metrics) *Client {
	return &Client{
		appKey:    appKey,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger:  logger,
		metrics: metrics,
	}
}

// SingleQuery 执行单次RTA查询
func (c *Client) SingleQuery(ctx context.Context, req *SingleRequest) (*SingleResponse, error) {
	// 参数验证
	if err := c.validateSingleRequest(req); err != nil {
		return nil, err
	}

	// 构建请求
	params := map[string]string{
		"app_key":    c.appKey,
		"channel":    req.Channel,
		"ad_space_id": req.AdvertisingSpaceID,
	}

	// 添加设备ID参数
	c.addDeviceParams(params, req)

	// 发送请求
	resp := &SingleResponse{}
	if err := c.doRequest(ctx, singleAPIURL, params, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// BatchQuery 执行批量RTA查询
func (c *Client) BatchQuery(ctx context.Context, req *BatchRequest) (*BatchResponse, error) {
	// 参数验证
	if err := c.validateBatchRequest(req); err != nil {
		return nil, err
	}

	// 构建请求
	params := map[string]string{
		"app_key":    c.appKey,
		"channel":    req.Channel,
		"ad_space_id": req.AdvertisingSpaceID,
	}

	// 添加设备ID列表
	if req.IMEIMD5List != "" {
		params["imei_md5"] = req.IMEIMD5List
	}
	if req.IDFAMD5List != "" {
		params["idfa_md5"] = req.IDFAMD5List
	}
	if req.OAIDMD5List != "" {
		params["oaid_md5"] = req.OAIDMD5List
	}

	// 发送请求
	resp := &BatchResponse{}
	if err := c.doRequest(ctx, batchAPIURL, params, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// validateSingleRequest 验证单次请求参数
func (c *Client) validateSingleRequest(req *SingleRequest) error {
	if req.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	if req.AdvertisingSpaceID == "" {
		return fmt.Errorf("advertising_space_id is required")
	}
	
	// 至少需要一个设备ID
	hasDeviceID := req.IMEI != "" || req.IMEIMD5 != "" ||
		req.IDFA != "" || req.IDFAMD5 != "" ||
		req.OAID != "" || req.OAIDMD5 != ""
	if !hasDeviceID {
		return fmt.Errorf("at least one device ID is required")
	}
	
	return nil
}

// validateBatchRequest 验证批量请求参数
func (c *Client) validateBatchRequest(req *BatchRequest) error {
	if req.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	if req.AdvertisingSpaceID == "" {
		return fmt.Errorf("advertising_space_id is required")
	}
	
	// 至少需要一个设备ID列表
	if req.IMEIMD5List == "" && req.IDFAMD5List == "" && req.OAIDMD5List == "" {
		return fmt.Errorf("at least one device ID list is required")
	}
	
	// 检查设备ID数量限制
	for _, list := range []string{req.IMEIMD5List, req.IDFAMD5List, req.OAIDMD5List} {
		if list != "" && len(strings.Split(list, ",")) > 20 {
			return fmt.Errorf("device ID list cannot contain more than 20 items")
		}
	}
	
	return nil
}

// addDeviceParams 添加设备相关参数
func (c *Client) addDeviceParams(params map[string]string, req *SingleRequest) {
	if req.IMEI != "" {
		params["imei"] = req.IMEI
	}
	if req.IMEIMD5 != "" {
		params["imei_md5"] = req.IMEIMD5
	}
	if req.IDFA != "" {
		params["idfa"] = req.IDFA
	}
	if req.IDFAMD5 != "" {
		params["idfa_md5"] = req.IDFAMD5
	}
	if req.OAID != "" {
		params["oaid"] = req.OAID
	}
	if req.OAIDMD5 != "" {
		params["oaid_md5"] = req.OAIDMD5
	}
	if req.OS != "" {
		params["os"] = req.OS
	}
	if req.Profile != "" {
		params["profile"] = req.Profile
	}
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, url string, params map[string]string, result interface{}) error {
	// TODO: 实现实际的HTTP请求逻辑
	// 1. 添加签名
	// 2. 发送请求
	// 3. 处理响应
	return nil
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
