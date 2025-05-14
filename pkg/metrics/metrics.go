package metrics

import (
	"fmt"
	"net/http"
	"time"

	"simple-dsp/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

// Metrics 监控指标集合
type Metrics struct {
	// HTTP相关指标
	RequestTotal    *prometheus.CounterVec
	RequestDuration prometheus.Histogram

	// gRPC相关指标
	GRPCRequestTotal    *prometheus.CounterVec
	GRPCRequestDuration prometheus.Histogram

	// 竞价相关指标
	BidRequests  prometheus.Counter
	BidResponses prometheus.Counter
	BidErrors    prometheus.Counter
	BidLatency   prometheus.Histogram
	BidPrice     *prometheus.HistogramVec
	WinPrice     *prometheus.HistogramVec
	BidDuration  prometheus.Histogram

	// 频次控制相关指标
	FrequencyCheckTotal     prometheus.Counter
	FrequencyLimitExceeded  prometheus.Counter
	FrequencyCheckDuration  prometheus.Histogram
	FrequencyRecordTotal    prometheus.Counter
	FrequencyRecordDuration prometheus.Histogram

	// 素材管理相关指标
	CreativeUploaded       prometheus.Counter
	CreativeDeleted        prometheus.Counter
	CreativeSize           prometheus.Histogram
	CreativeGroupCreated   prometheus.Counter
	CreativeGroupDeleted   prometheus.Counter
	CreativeUploadDuration prometheus.Histogram
	CreativeAuditTotal     prometheus.Counter
	CreativeAuditApproved  prometheus.Counter
	CreativeAuditRejected  prometheus.Counter

	// 缓存相关指标
	CacheHits    prometheus.Counter
	CacheMisses  prometheus.Counter
	CacheErrors  prometheus.Counter
	CacheLatency prometheus.Histogram

	// 存储相关指标
	StorageUploadTotal   prometheus.Counter
	StorageUploadErrors  prometheus.Counter
	StorageUploadLatency prometheus.Histogram
	StorageDeleteTotal   prometheus.Counter
	StorageDeleteErrors  prometheus.Counter
	StorageDeleteLatency prometheus.Histogram

	// RTA检查时间
	RTACheckDuration prometheus.Histogram

	// 事件处理时间
	EventHandleDuration prometheus.Histogram

	// 预算检查时间
	BudgetCheckDuration prometheus.Histogram

	// HTTP服务器
	server *http.Server
}

// NoopMetrics 是一个不执行任何操作的指标收集器
type NoopMetrics struct{}

// NewNoopMetrics 创建一个不执行任何操作的指标收集器
func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{}
}

// Observe 为 NoopMetrics 实现所有指标方法
func (m *NoopMetrics) Observe(float64)                               {}
func (m *NoopMetrics) Inc()                                          {}
func (m *NoopMetrics) Add(float64)                                   {}
func (m *NoopMetrics) WithLabelValues(...string) prometheus.Observer { return m }
func (m *NoopMetrics) With(prometheus.Labels) prometheus.Observer    { return m }

// NewMetrics 创建监控指标集合
func NewMetrics(port int, path string) *Metrics {
	cfg := config.MetricsConfig{
		Enabled: true,
		Port:    port,
		Path:    path,
	}

	m, err := newMetrics(cfg)
	if err != nil {
		// 如果创建失败，返回 NoopMetrics
		return &Metrics{}
	}
	return m
}

// newMetrics 内部函数，用于创建监控指标集合
func newMetrics(cfg config.MetricsConfig) (*Metrics, error) {
	m := &Metrics{
		// HTTP相关指标
		RequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "HTTP请求总数",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_request_duration_seconds",
			Help:    "请求处理时间",
			Buckets: prometheus.DefBuckets,
		}),

		// gRPC相关指标
		GRPCRequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_requests_total",
				Help: "gRPC请求总数",
			},
			[]string{"method", "status"},
		),
		GRPCRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC请求延迟分布",
			Buckets: prometheus.DefBuckets,
		}),

		// 竞价相关指标
		BidRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_bid_requests_total",
			Help: "竞价请求总数",
		}),
		BidResponses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_bid_responses_total",
			Help: "竞价响应总数",
		}),
		BidErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_bid_errors_total",
			Help: "竞价错误总数",
		}),
		BidLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_bid_latency_seconds",
			Help:    "竞价延迟分布",
			Buckets: prometheus.DefBuckets,
		}),
		BidPrice: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "dsp_bid_price",
			Help:    "竞价出价分布",
			Buckets: prometheus.LinearBuckets(0, 10, 10),
		}, []string{"ad_id"}),
		WinPrice: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "dsp_win_price",
			Help:    "竞价获胜价格分布",
			Buckets: prometheus.LinearBuckets(0, 10, 10),
		}, []string{"ad_id"}),
		BidDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_bid_duration_seconds",
			Help:    "竞价处理时间分布",
			Buckets: prometheus.DefBuckets,
		}),

		// 频次控制相关指标
		FrequencyCheckTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_frequency_check_total",
			Help: "频次检查总数",
		}),
		FrequencyLimitExceeded: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_frequency_limit_exceeded_total",
			Help: "频次超限总数",
		}),
		FrequencyCheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_frequency_check_duration_seconds",
			Help:    "频次检查耗时分布",
			Buckets: prometheus.DefBuckets,
		}),
		FrequencyRecordTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_frequency_record_total",
			Help: "频次记录总数",
		}),
		FrequencyRecordDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_frequency_record_duration_seconds",
			Help:    "频次记录耗时分布",
			Buckets: prometheus.DefBuckets,
		}),

		// 素材管理相关指标
		CreativeUploaded: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_creative_uploaded_total",
			Help: "素材上传总数",
		}),
		CreativeDeleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_creative_deleted_total",
			Help: "素材删除总数",
		}),
		CreativeSize: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_creative_size_bytes",
			Help:    "素材大小分布",
			Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
		}),
		CreativeGroupCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_creative_group_created_total",
			Help: "素材组创建总数",
		}),
		CreativeGroupDeleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_creative_group_deleted_total",
			Help: "素材组删除总数",
		}),
		CreativeUploadDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_creative_upload_duration_seconds",
			Help:    "素材上传耗时分布",
			Buckets: prometheus.DefBuckets,
		}),
		CreativeAuditTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_creative_audit_total",
			Help: "素材审核总数",
		}),
		CreativeAuditApproved: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_creative_audit_approved_total",
			Help: "素材审核通过总数",
		}),
		CreativeAuditRejected: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_creative_audit_rejected_total",
			Help: "素材审核拒绝总数",
		}),

		// 缓存相关指标
		CacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_cache_hits_total",
			Help: "缓存命中总数",
		}),
		CacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_cache_misses_total",
			Help: "缓存未命中总数",
		}),
		CacheErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_cache_errors_total",
			Help: "缓存错误总数",
		}),
		CacheLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_cache_latency_seconds",
			Help:    "缓存操作延迟分布",
			Buckets: prometheus.DefBuckets,
		}),

		// 存储相关指标
		StorageUploadTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_storage_upload_total",
			Help: "存储上传总数",
		}),
		StorageUploadErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_storage_upload_errors_total",
			Help: "存储上传错误总数",
		}),
		StorageUploadLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_storage_upload_latency_seconds",
			Help:    "存储上传延迟分布",
			Buckets: prometheus.DefBuckets,
		}),
		StorageDeleteTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_storage_delete_total",
			Help: "存储删除总数",
		}),
		StorageDeleteErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dsp_storage_delete_errors_total",
			Help: "存储删除错误总数",
		}),
		StorageDeleteLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_storage_delete_latency_seconds",
			Help:    "存储删除延迟分布",
			Buckets: prometheus.DefBuckets,
		}),

		// RTA检查时间
		RTACheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_rta_check_duration_seconds",
			Help:    "RTA检查时间",
			Buckets: prometheus.DefBuckets,
		}),

		// 事件处理时间
		EventHandleDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "dsp_event_handle_duration_seconds",
			Help:    "事件处理时间",
			Buckets: prometheus.DefBuckets,
		}, []string{"event_type"}),

		// 预算检查时间
		BudgetCheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "dsp_budget_check_duration_seconds",
			Help:    "预算检查时间",
			Buckets: prometheus.DefBuckets,
		}),
	}

	// 如果启用了指标收集
	if cfg.Enabled {
		// 创建HTTP服务器
		mux := http.NewServeMux()
		mux.Handle(cfg.Path, promhttp.Handler())

		m.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: mux,
		}

		// 启动HTTP服务器
		go func() {
			if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				// 这里不能使用logger，因为可能造成循环依赖
				fmt.Printf("metrics服务器错误: %v\n", err)
			}
		}()
	}

	return m, nil
}

// Close 关闭metrics服务器
func (m *Metrics) Close() error {
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}

// StartPushGateway 启动指标推送到 PushGateway
func (m *Metrics) StartPushGateway(pushGatewayURL string) {
	go func() {
		for {
			// 创建 pusher
			pusher := push.New(pushGatewayURL, "dsp_metrics")

			// 添加所有指标
			pusher.Collector(m.RequestTotal)
			pusher.Collector(m.RequestDuration)
			pusher.Collector(m.GRPCRequestTotal)
			pusher.Collector(m.GRPCRequestDuration)
			pusher.Collector(m.BidRequests)
			pusher.Collector(m.BidResponses)
			pusher.Collector(m.BidErrors)
			pusher.Collector(m.BidLatency)
			pusher.Collector(m.BidPrice)
			pusher.Collector(m.WinPrice)
			pusher.Collector(m.BidDuration)
			pusher.Collector(m.FrequencyCheckTotal)
			pusher.Collector(m.FrequencyLimitExceeded)
			pusher.Collector(m.FrequencyCheckDuration)
			pusher.Collector(m.FrequencyRecordTotal)
			pusher.Collector(m.FrequencyRecordDuration)
			pusher.Collector(m.CreativeUploaded)
			pusher.Collector(m.CreativeDeleted)
			pusher.Collector(m.CreativeSize)
			pusher.Collector(m.CreativeGroupCreated)
			pusher.Collector(m.CreativeGroupDeleted)
			pusher.Collector(m.CreativeUploadDuration)
			pusher.Collector(m.CreativeAuditTotal)
			pusher.Collector(m.CreativeAuditApproved)
			pusher.Collector(m.CreativeAuditRejected)
			pusher.Collector(m.CacheHits)
			pusher.Collector(m.CacheMisses)
			pusher.Collector(m.CacheErrors)
			pusher.Collector(m.CacheLatency)
			pusher.Collector(m.StorageUploadTotal)
			pusher.Collector(m.StorageUploadErrors)
			pusher.Collector(m.StorageUploadLatency)
			pusher.Collector(m.StorageDeleteTotal)
			pusher.Collector(m.StorageDeleteErrors)
			pusher.Collector(m.StorageDeleteLatency)
			pusher.Collector(m.RTACheckDuration)
			pusher.Collector(m.EventHandleDuration)
			pusher.Collector(m.BudgetCheckDuration)

			// 推送指标
			if err := pusher.Push(); err != nil {
				fmt.Printf("推送指标到 PushGateway 失败: %v\n", err)
			}

			// 等待下一次推送
			time.Sleep(15 * time.Second)
		}
	}()
}
