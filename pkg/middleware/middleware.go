package middleware

import (
	"context"
	"strconv"
	"time"

	"simple-dsp/pkg/metrics"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

// Logger 日志中间件
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		if len(c.Errors) > 0 {
			// 记录错误
			for _, e := range c.Errors.Errors() {
				log.Error("请求处理错误",
					zap.String("path", path),
					zap.String("query", query),
					zap.String("method", c.Request.Method),
					zap.Int("status", c.Writer.Status()),
					zap.String("error", e),
					zap.Duration("latency", latency),
				)
			}
		} else {
			log.Info("请求处理完成",
				zap.String("path", path),
				zap.String("query", query),
				zap.String("method", c.Request.Method),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
			)
		}
	}
}

// Metrics 指标收集中间件
func Metrics(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		// 记录请求延迟
		m.HTTP.RequestDuration.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(c.Writer.Status()),
		).Observe(time.Since(start).Seconds())

		// 记录请求总数
		m.HTTP.RequestTotal.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()
	}
}

// RateLimit 限流中间件
func RateLimit(qps float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatus(429) // Too Many Requests
			return
		}
		c.Next()
	}
}

// GRPCMetrics gRPC指标收集拦截器
func GRPCMetrics(m *metrics.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// 记录gRPC请求延迟
		m.GRPC.RequestDuration.WithLabelValues(
			info.FullMethod,
			errorToCode(err),
		).Observe(duration.Seconds())

		// 记录gRPC请求总数
		m.GRPC.RequestTotal.WithLabelValues(
			info.FullMethod,
			errorToCode(err),
		).Inc()

		return resp, err
	}
}

// errorToCode 将错误转换为状态码字符串
func errorToCode(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}
