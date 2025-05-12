package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 监控指标集合
type Metrics struct {
	// 竞价相关指标
	BidRequests        prometheus.Counter
	BidResponses      prometheus.Counter
	BidErrors         prometheus.Counter
	BidLatency        prometheus.Histogram
	BidPrice          *prometheus.HistogramVec
	WinPrice          *prometheus.HistogramVec
	BidDuration       prometheus.Histogram

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
	CacheHits             prometheus.Counter
	CacheMisses           prometheus.Counter
	CacheErrors           prometheus.Counter
	CacheLatency          prometheus.Histogram

	// 存储相关指标
	StorageUploadTotal    prometheus.Counter
	StorageUploadErrors   prometheus.Counter
	StorageUploadLatency  prometheus.Histogram
	StorageDeleteTotal    prometheus.Counter
	StorageDeleteErrors   prometheus.Counter
	StorageDeleteLatency  prometheus.Histogram
}

// NewMetrics 创建监控指标集合
func NewMetrics() *Metrics {
	return &Metrics{
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
	}
} 