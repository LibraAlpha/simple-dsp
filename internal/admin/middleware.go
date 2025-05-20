package admin

import (
	"fmt"
	"net/http"
	"time"

	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Middleware 中间件接口
type Middleware interface {
	Auth() gin.HandlerFunc
	RateLimit() gin.HandlerFunc
	Logger() gin.HandlerFunc
	Recovery() gin.HandlerFunc
}

// middleware 中间件实现
type middleware struct {
	logger  *logger.Logger
	limiter *rate.Limiter
	metrics *metrics.Metrics
}

// NewMiddleware 创建中间件
func NewMiddleware(logger *logger.Logger, qps float64, burst int, metrics *metrics.Metrics) Middleware {
	return &middleware{
		logger:  logger,
		limiter: rate.NewLimiter(rate.Limit(qps), burst),
		metrics: metrics,
	}
}

// Auth 认证中间件
func (m *middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现认证逻辑
		// 1. 检查请求头中的认证信息
		// 2. 验证 token
		// 3. 检查权限
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		// 这里简单实现，实际应该验证 token
		if token != "Bearer admin-token" {
			c.JSON(http.StatusForbidden, gin.H{"error": "禁止访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit 限流中间件
func (m *middleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Logger 日志中间件
func (m *middleware) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取请求信息
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		// 记录访问日志
		m.logger.Info("访问日志",
			"status", status,
			"latency", latency,
			"client_ip", clientIP,
			"method", method,
			"path", path,
		)

		// 记录指标
		m.metrics.HTTP.RequestTotal.WithLabelValues(method, path, fmt.Sprint(status)).Inc()
		m.metrics.HTTP.RequestDuration.WithLabelValues(method, path).Observe(latency.Seconds())
	}
}

// Recovery 恢复中间件
func (m *middleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				m.logger.Error("系统错误",
					"error", err,
					"client_ip", c.ClientIP(),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
				)

				// 记录指标
				m.metrics.Bid.Errors.Inc()

				// 返回错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "服务器内部错误",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
