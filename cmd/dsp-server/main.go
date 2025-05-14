package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"simple-dsp/internal/bidding"
	"simple-dsp/internal/budget"
	"simple-dsp/internal/event"
	"simple-dsp/internal/frequency"
	"simple-dsp/internal/rta"
	"simple-dsp/internal/stats"
	"simple-dsp/internal/traffic"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// RedisClient Redis客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Close() error
}

// KafkaClient Kafka客户端接口
type KafkaClient interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
	Close() error
}

// RedisClientWrapper Redis客户端包装器
type RedisClientWrapper struct {
	*redis.Client
}

// KafkaWriter Kafka写入器
type KafkaWriter struct {
	*kafka.Writer
}

// SendMessage 发送消息到Kafka
func (w *KafkaWriter) SendMessage(ctx context.Context, topic string, key string, value []byte) error {
	return w.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	})
}

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logger.NewLogger(cfg.Log)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := log.Sync(); err != nil {
			fmt.Printf("同步日志失败: %v\n", err)
		}
	}()

	// 初始化监控指标
	metricsCollector := metrics.NewMetrics(cfg.Metrics.Port, cfg.Metrics.Path)
	if cfg.Metrics.PushGatewayURL != "" {
		metricsCollector.StartPushGateway(cfg.Metrics.PushGatewayURL)
	}

	// 初始化Redis客户端
	redisClient := initRedis(cfg.Redis)
	defer redisClient.Close()

	// 初始化Kafka客户端
	kafkaClient := initKafka(cfg.Kafka)
	defer kafkaClient.Close()

	// 初始化RTA客户端
	rtaClient := rta.NewClient(
		cfg.RTA.BaseURL,
		cfg.RTA.AppKey,
		cfg.RTA.AppSecret,
		log,
		metricsCollector,
	)

	// 初始化预算管理器
	budgetMgr := budget.NewManager(redisClient, log, metricsCollector)

	// 初始化频次控制器
	freqCtrl := frequency.NewController(redisClient, log, metricsCollector)

	// 初始化数据统计收集器
	statsCollector := stats.NewCollector(kafkaClient, redisClient, log, metricsCollector)

	// 初始化竞价引擎
	biddingEngine := bidding.NewEngine(
		nil, // TODO: 实现广告服务
		budgetMgr,
		freqCtrl,
		log,
		metricsCollector,
	)

	// 初始化事件处理器
	eventHandler := event.NewHandler(statsCollector, log, metricsCollector)

	// 初始化流量处理器
	trafficHandler := traffic.NewHandler(
		rtaClient,
		biddingEngine,
		eventHandler,
		log,
		metricsCollector,
	)

	// 初始化路由
	router := initRouter(trafficHandler, eventHandler)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("启动服务器失败", "error", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("关闭服务器失败", "error", err)
		os.Exit(1)
	}
}

// initRedis 初始化Redis客户端
func initRedis(cfg config.RedisConfig) RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("连接Redis失败: %v\n", err)
		os.Exit(1)
	}

	return &RedisClientWrapper{client}
}

// initKafka 初始化Kafka客户端
func initKafka(cfg config.KafkaConfig) KafkaClient {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		Async:        cfg.Async,
	}

	return &KafkaWriter{writer}
}

// initRouter 初始化路由
func initRouter(trafficHandler *traffic.Handler, eventHandler *event.Handler) *gin.Engine {
	router := gin.Default()

	// 流量接入接口
	router.POST("/api/v1/traffic", gin.HandlerFunc(trafficHandler.HandleRequest))

	// 事件处理接口
	router.POST("/api/v1/events/impression", gin.HandlerFunc(eventHandler.HandleImpression))
	router.POST("/api/v1/events/click", gin.HandlerFunc(eventHandler.HandleClick))
	router.POST("/api/v1/events/conversion", gin.HandlerFunc(eventHandler.HandleConversion))
	router.GET("/api/v1/events/stats", gin.HandlerFunc(eventHandler.GetEventStats))

	// 健康检查接口
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
