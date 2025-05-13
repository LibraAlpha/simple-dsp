package rta

// DeviceType 设备系统类型
type DeviceType int

const (
    Android DeviceType = 0
    IOS     DeviceType = 1
    WinPhone DeviceType = 2
    Other    DeviceType = 3
)

// SingleRequest 单次请求参数
type SingleRequest struct {
    Channel            string `json:"channel"`             // 大航海渠道ID
    AdvertisingSpaceID string `json:"advertisingSpaceId"` // 大航海广告位ID
    IMEI              string `json:"imei,omitempty"`      // IMEI原生值
    IMEIMD5           string `json:"imei_md5,omitempty"`  // IMEI的MD5值
    IDFA              string `json:"idfa,omitempty"`      // IDFA原生值
    IDFAMD5           string `json:"idfa_md5,omitempty"`  // IDFA的MD5值
    OAID              string `json:"oaid,omitempty"`      // OAID原生值
    OAIDMD5           string `json:"oaid_md5,omitempty"`  // OAID的MD5值
    OS                string `json:"os,omitempty"`        // 设备系统类型
    Profile           string `json:"profile,omitempty"`   // 预留JSON参数
}

// SingleResponse 单次请求响应
type SingleResponse struct {
    ErrCode int    `json:"errcode"` // 错误码，0：成功；1：限流；2：服务不可用
    Result  bool   `json:"result"`  // true: 目标用户；false: 非目标用户
    TaskID  string `json:"task_id"` // 大航海平台任务ID
}

// BatchRequest 批量请求参数
type BatchRequest struct {
    Channel            string `json:"channel"`             // 大航海渠道ID
    AdvertisingSpaceID string `json:"advertisingSpaceId"` // 大航海广告位ID
    IMEIMD5List       string `json:"imei_md5,omitempty"`  // IMEI的MD5值列表，逗号分隔
    IDFAMD5List       string `json:"idfa_md5,omitempty"`  // IDFA的MD5值列表，逗号分隔
    OAIDMD5List       string `json:"oaid_md5,omitempty"`  // OAID的MD5值列表，逗号分隔
}

// BatchResult 批量结果项
type BatchResult struct {
    TaskID   string `json:"task_id"`   // 任务ID
    IMEIMD5  string `json:"imei_md5"`  // 命中的IMEI MD5
    IDFAMD5  string `json:"idfa_md5"`  // 命中的IDFA MD5
    OAIDMD5  string `json:"oaid_md5"`  // 命中的OAID MD5
}

// BatchResponse 批量请求响应
type BatchResponse struct {
    ErrCode int           `json:"errcode"` // 错误码
    Results []BatchResult `json:"results"` // 结果数组
}

// Error codes
const (
    ErrCodeSuccess      = 0
    ErrCodeRateLimit    = 1
    ErrCodeUnavailable  = 2
) 