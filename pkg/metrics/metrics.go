/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: metrics.go
 * Project: simple-dsp
 * Description: 指标监控模块，提供系统性能指标收集和展示
 *
 * 主要功能:
 * - 收集系统性能指标
 * - 提供指标查询接口
 * - 支持指标告警
 * - 集成Prometheus
 *
 * 实现细节:
 * - 使用Prometheus客户端
 * - 实现自定义指标
 * - 支持指标聚合
 * - 提供PushGateway集成
 *
 * 依赖关系:
 * - github.com/prometheus/client_golang
 * - simple-dsp/pkg/config
 * - simple-dsp/pkg/logger
 *
 * 注意事项:
 * - 注意指标命名规范
 * - 合理设置采集周期
 * - 注意指标数据量
 * - 确保指标准确性
 */

// Package metrics metrics.go
package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"

	"simple-dsp/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 模块化指标定义（网页4）
type (
	HTTPMetrics struct {
		RequestTotal    *prometheus.CounterVec
		RequestDuration *prometheus.HistogramVec
	}

	GRPCMetrics struct {
		RequestTotal    *prometheus.CounterVec
		RequestDuration *prometheus.HistogramVec
	}

	BidMetrics struct {
		Requests  prometheus.Counter
		Responses prometheus.Counter
		Errors    prometheus.Counter
		Latency   prometheus.Histogram
		Price     *prometheus.HistogramVec
		WinPrice  *prometheus.HistogramVec
		Duration  prometheus.Histogram
	}

	FrequencyMetrics struct {
		CheckTotal     prometheus.Counter
		LimitExceeded  prometheus.Counter
		CheckDuration  prometheus.Histogram
		RecordTotal    prometheus.Counter
		RecordDuration prometheus.Histogram
	}

	CreativeMetrics struct {
		Uploaded       prometheus.Counter
		Deleted        prometheus.Counter
		Size           prometheus.Histogram
		GroupCreated   prometheus.Counter
		GroupDeleted   prometheus.Counter
		UploadDuration prometheus.Histogram
		AuditTotal     prometheus.Counter
		AuditApproved  prometheus.Counter
		AuditRejected  prometheus.Counter
	}

	CacheMetrics struct {
		Hits    prometheus.Counter
		Misses  prometheus.Counter
		Errors  prometheus.Counter
		Latency prometheus.Histogram
	}

	StorageMetrics struct {
		UploadTotal   prometheus.Counter
		UploadErrors  prometheus.Counter
		UploadLatency prometheus.Histogram
		DeleteTotal   prometheus.Counter
		DeleteErrors  prometheus.Counter
		DeleteLatency prometheus.Histogram
	}

	EventMetrics struct {
		Impressions *prometheus.CounterVec
		Clicks      *prometheus.CounterVec
		Conversions *prometheus.CounterVec
	}

	BudgetMetrics struct {
		Cost        *prometheus.CounterVec
		DailyBudget *prometheus.CounterVec
	}

	RTAMetrics struct {
		CheckDuration      prometheus.Histogram
		BatchCheckDuration prometheus.Histogram
		Requests           prometheus.Counter
		Errors             prometheus.Counter
	}

	TrackingMetrics struct {
		Duration *prometheus.HistogramVec
		Success  *prometheus.CounterVec
		Failure  *prometheus.CounterVec
	}
)

type Metrics struct {
	HTTP      *HTTPMetrics
	GRPC      *GRPCMetrics
	Bid       *BidMetrics
	Budget    *BudgetMetrics
	Frequency *FrequencyMetrics
	Creative  *CreativeMetrics
	Cache     *CacheMetrics
	Storage   *StorageMetrics
	Events    *EventMetrics
	RTA       *RTAMetrics
	Tracking  *TrackingMetrics
	server    *http.Server
}

// NoopMetrics NoopMetrics实现
type NoopMetrics struct {
	HTTP      *HTTPMetrics
	GRPC      *GRPCMetrics
	Bid       *BidMetrics
	Frequency *FrequencyMetrics
	Creative  *CreativeMetrics
	Cache     *CacheMetrics
	Storage   *StorageMetrics
	Events    *EventMetrics
	Budget    *BudgetMetrics
}

func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{
		HTTP:      &HTTPMetrics{},
		GRPC:      &GRPCMetrics{},
		Bid:       &BidMetrics{},
		Frequency: &FrequencyMetrics{},
		Creative:  &CreativeMetrics{},
		Cache:     &CacheMetrics{},
		Storage:   &StorageMetrics{},
		Events:    &EventMetrics{},
		Budget:    &BudgetMetrics{},
	}
}

// Observe 空操作方法
func (m *NoopMetrics) Observe(float64)                               {}
func (m *NoopMetrics) Inc()                                          {}
func (m *NoopMetrics) Add(float64)                                   {}
func (m *NoopMetrics) WithLabelValues(...string) prometheus.Observer { return m }
func (m *NoopMetrics) With(prometheus.Labels) prometheus.Observer    { return m }

func NewMetrics(cfg config.MetricsConfig) (*Metrics, error) {
	if !cfg.Enabled {
		return &Metrics{}, nil
	}

	registry := prometheus.NewRegistry()

	metrics := &Metrics{
		HTTP: &HTTPMetrics{
			RequestTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "http_requests_total",
					Help: "HTTP请求总数",
				},
				[]string{"method", "path", "status"},
			),
			RequestDuration: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "http_request_duration_seconds",
					Help:    "HTTP请求延迟分布",
					Buckets: prometheus.DefBuckets,
				},
				[]string{"method", "path"},
			),
		},

		GRPC: &GRPCMetrics{
			RequestTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "grpc_requests_total",
					Help: "gRPC请求总数",
				},
				[]string{"method", "status"},
			),
			RequestDuration: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "grpc_request_duration_seconds",
					Help:    "gRPC请求延迟分布",
					Buckets: prometheus.DefBuckets,
				},
				[]string{"method", "status"},
			),
		},

		Bid: &BidMetrics{
			Requests: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_bid_requests_total",
				Help: "竞价请求总数",
			}),
			Responses: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_bid_responses_total",
				Help: "竞价响应总数",
			}),
			Errors: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_bid_errors_total",
				Help: "竞价错误总数",
			}),
			Latency: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_bid_latency_seconds",
				Help:    "竞价延迟分布",
				Buckets: prometheus.DefBuckets,
			}),
			Price: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "dsp_bid_price",
				Help:    "竞价出价分布",
				Buckets: prometheus.LinearBuckets(0, 10, 10),
			}, []string{"ad_type", "campaign"}),
			WinPrice: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "dsp_win_price",
				Help:    "竞价获胜价格分布",
				Buckets: prometheus.LinearBuckets(0, 10, 10),
			}, []string{"ad_type", "campaign"}),
			Duration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_bid_duration_seconds",
				Help:    "竞价处理时间分布",
				Buckets: prometheus.DefBuckets,
			}),
		},

		Frequency: &FrequencyMetrics{
			CheckTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_frequency_check_total",
				Help: "频次检查总数",
			}),
			LimitExceeded: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_frequency_limit_exceeded_total",
				Help: "频次超限总数",
			}),
			CheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_frequency_check_duration_seconds",
				Help:    "频次检查耗时分布",
				Buckets: prometheus.DefBuckets,
			}),
			RecordTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_frequency_record_total",
				Help: "频次记录总数",
			}),
			RecordDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_frequency_record_duration_seconds",
				Help:    "频次记录耗时分布",
				Buckets: prometheus.DefBuckets,
			}),
		},

		Creative: &CreativeMetrics{
			Uploaded: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_creative_uploaded_total",
				Help: "素材上传总数",
			}),
			Deleted: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_creative_deleted_total",
				Help: "素材删除总数",
			}),
			Size: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_creative_size_bytes",
				Help:    "素材大小分布",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
			}),
			GroupCreated: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_creative_group_created_total",
				Help: "素材组创建总数",
			}),
			GroupDeleted: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_creative_group_deleted_total",
				Help: "素材组删除总数",
			}),
			UploadDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_creative_upload_duration_seconds",
				Help:    "素材上传耗时分布",
				Buckets: prometheus.DefBuckets,
			}),
			AuditTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_creative_audit_total",
				Help: "素材审核总数",
			}),
			AuditApproved: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_creative_audit_approved_total",
				Help: "素材审核通过总数",
			}),
			AuditRejected: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_creative_audit_rejected_total",
				Help: "素材审核拒绝总数",
			}),
		},

		Cache: &CacheMetrics{
			Hits: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_cache_hits_total",
				Help: "缓存命中总数",
			}),
			Misses: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_cache_misses_total",
				Help: "缓存未命中总数",
			}),
			Errors: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_cache_errors_total",
				Help: "缓存错误总数",
			}),
			Latency: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_cache_latency_seconds",
				Help:    "缓存操作延迟分布",
				Buckets: prometheus.DefBuckets,
			}),
		},

		Storage: &StorageMetrics{
			UploadTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_storage_upload_total",
				Help: "存储上传总数",
			}),
			UploadErrors: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_storage_upload_errors_total",
				Help: "存储上传错误总数",
			}),
			UploadLatency: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_storage_upload_latency_seconds",
				Help:    "存储上传延迟分布",
				Buckets: prometheus.DefBuckets,
			}),
			DeleteTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_storage_delete_total",
				Help: "存储删除总数",
			}),
			DeleteErrors: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_storage_delete_errors_total",
				Help: "存储删除错误总数",
			}),
			DeleteLatency: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_storage_delete_latency_seconds",
				Help:    "存储删除延迟分布",
				Buckets: prometheus.DefBuckets,
			}),
		},

		Events: &EventMetrics{
			Impressions: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "dsp_event_impression",
					Help: "曝光数",
				},
				[]string{"ad_id", "slot_id"},
			),
			Clicks: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "dsp_event_clicks",
					Help: "点击数",
				},
				[]string{"ad_id", "slot_id"},
			),
			Conversions: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "dsp_event_conversions",
					Help: "点击数",
				},
				[]string{"ad_id", "slot_id"},
			),
		},

		RTA: &RTAMetrics{
			CheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_rta_check_duration_seconds",
				Help:    "RTA检查耗时分布",
				Buckets: prometheus.DefBuckets,
			}),
			BatchCheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "dsp_rta_batch_check_duration_seconds",
				Help:    "RTA批量检查耗时分布",
				Buckets: prometheus.DefBuckets,
			}),
			Requests: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_rta_requests_total",
				Help: "RTA请求总数",
			}),
			Errors: promauto.NewCounter(prometheus.CounterOpts{
				Name: "dsp_rta_errors_total",
				Help: "RTA错误总数",
			}),
		},

		Tracking: &TrackingMetrics{
			Duration: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "dsp_tracking_duration_seconds",
				Help:    "跟踪请求耗时分布",
				Buckets: prometheus.DefBuckets,
			}, []string{"event_type"}),
			Success: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "dsp_tracking_success_total",
				Help: "跟踪请求成功总数",
			}, []string{"event_type"}),
			Failure: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "dsp_tracking_failure_total",
				Help: "跟踪请求失败总数",
			}, []string{"event_type"}),
		},
	}

	// 注册全局采集器
	registry.MustRegister(
		metrics.HTTP.RequestTotal,
		metrics.HTTP.RequestDuration,
		metrics.GRPC.RequestTotal,
		metrics.GRPC.RequestDuration,
		metrics.Bid.Requests,
		metrics.Bid.Responses,
		metrics.Bid.Errors,
		metrics.Bid.Latency,
		metrics.Bid.Price,
		metrics.Bid.WinPrice,
		metrics.Bid.Duration,
		metrics.Frequency.CheckTotal,
		metrics.Frequency.LimitExceeded,
		metrics.Frequency.CheckDuration,
		metrics.Frequency.RecordTotal,
		metrics.Frequency.RecordDuration,
		metrics.Creative.Uploaded,
		metrics.Creative.Deleted,
		metrics.Creative.Size,
		metrics.Creative.GroupCreated,
		metrics.Creative.GroupDeleted,
		metrics.Creative.UploadDuration,
		metrics.Creative.AuditTotal,
		metrics.Creative.AuditApproved,
		metrics.Creative.AuditRejected,
		metrics.Cache.Hits,
		metrics.Cache.Misses,
		metrics.Cache.Errors,
		metrics.Cache.Latency,
		metrics.Storage.UploadTotal,
		metrics.Storage.UploadErrors,
		metrics.Storage.UploadLatency,
		metrics.Storage.DeleteTotal,
		metrics.Storage.DeleteErrors,
		metrics.Storage.DeleteLatency,
		metrics.Events.Clicks,
		metrics.Events.Impressions,
		metrics.Events.Conversions,
		metrics.Budget.DailyBudget,
		metrics.Budget.Cost,
		metrics.RTA.CheckDuration,
		metrics.RTA.BatchCheckDuration,
		metrics.RTA.Requests,
		metrics.RTA.Errors,
		metrics.Tracking.Duration,
		metrics.Tracking.Success,
		metrics.Tracking.Failure,
	)

	if cfg.HTTPEnabled {
		mux := http.NewServeMux()
		mux.Handle(cfg.Path, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

		metrics.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: mux,
		}

		go func() {
			if err := metrics.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
				fmt.Printf("Metrics server error: %v\n", err)
			}
		}()
	}

	return metrics, nil
}

// Close 关闭服务
func (m *Metrics) Close() error {
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}

// StartPushGateway 推送指标到Gateway
func (m *Metrics) StartPushGateway(url string) {
	pusher := push.New(url, "dsp_metrics")

	collectors := []prometheus.Collector{
		m.HTTP.RequestTotal,
		m.HTTP.RequestDuration,
		m.GRPC.RequestTotal,
		m.GRPC.RequestDuration,
		m.Bid.Requests,
		m.Bid.Responses,
		m.Bid.Errors,
		m.Bid.Latency,
		m.Bid.Price,
		m.Bid.WinPrice,
		m.Bid.Duration,
		m.Frequency.CheckTotal,
		m.Frequency.LimitExceeded,
		m.Frequency.CheckDuration,
		m.Frequency.RecordTotal,
		m.Frequency.RecordDuration,
		m.Creative.Uploaded,
		m.Creative.Deleted,
		m.Creative.Size,
		m.Creative.GroupCreated,
		m.Creative.GroupDeleted,
		m.Creative.UploadDuration,
		m.Creative.AuditTotal,
		m.Creative.AuditApproved,
		m.Creative.AuditRejected,
		m.Cache.Hits,
		m.Cache.Misses,
		m.Cache.Errors,
		m.Cache.Latency,
		m.Storage.UploadTotal,
		m.Storage.UploadErrors,
		m.Storage.UploadLatency,
		m.Storage.DeleteTotal,
		m.Storage.DeleteErrors,
		m.Storage.DeleteLatency,
		m.Events.Clicks,
		m.Events.Impressions,
		m.Events.Conversions,
		m.Budget.DailyBudget,
		m.Budget.Cost,
		m.RTA.CheckDuration,
		m.RTA.BatchCheckDuration,
		m.RTA.Requests,
		m.RTA.Errors,
		m.Tracking.Duration,
		m.Tracking.Success,
		m.Tracking.Failure,
	}

	for _, c := range collectors {
		pusher.Collector(c)
	}

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			if err := pusher.Push(); err != nil {
				fmt.Printf("Push failed: %v\n", err)
			}
		}
	}()
}

// RecordHTTPRequest 操作方法示例
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration float64) {
	m.HTTP.RequestTotal.WithLabelValues(method, path, status).Inc()
	m.HTTP.RequestDuration.WithLabelValues(method, path).Observe(duration)
}

func (m *Metrics) RecordBidPrice(adType, campaign string, price float64) {
	m.Bid.Price.WithLabelValues(adType, campaign).Observe(price)
}
