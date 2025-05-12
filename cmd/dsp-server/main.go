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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"simple-dsp/internal/bidding"
	"simple-dsp/internal/budget"
	"simple-dsp/internal/event"
	"simple-dsp/internal/rta"
	"simple-dsp/internal/stats"
	"simple-dsp/internal/traffic"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

func main() {
	// 1. 加载配置
	if err := config.LoadConfig("configs/config.yaml"); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	cfg := config.GetConfig()

	// 2. 初始化日志
	log := initLogger(cfg.Log)
	defer log.Sync()

	// 3. 初始化监控指标
	metrics := initMetrics(cfg.Metrics)

	// 4. 初始化Redis客户端
	redisClient := initRedis(cfg.Redis)
	defer redisClient.Close()

	// 5. 初始化Kafka客户端
	kafkaClient := initKafka(cfg.Kafka)
	defer kafkaClient.Close()

	// 6. 初始化各个模块
	// 6.1 初始化RTA客户端
	rtaClient := rta.NewClient(
		cfg.RTA.BaseURL,
		log,
		metrics,
	)

	// 6.2 初始化预算管理器
	budgetMgr := budget.NewManager(
		redisClient,
		log,
		metrics,
	)

	// 6.3 初始化数据统计收集器
	statsCollector := stats.NewCollector(
		kafkaClient,
		redisClient,
		log,
		metrics,
	)

	// 6.4 初始化竞价引擎
	biddingEngine := bidding.NewEngine(
		nil, // TODO: 实现广告服务
		budgetMgr,
		log,
		metrics,
	)

	// 6.5 初始化事件处理器
	eventHandler := event.NewHandler(
		statsCollector,
		log,
		metrics,
	)

	// 6.6 初始化流量处理器
	trafficHandler := traffic.NewHandler(
		biddingEngine,
		rtaClient,
		budgetMgr,
		log,
		metrics,
		traffic.HandlerConfig{
			QPS:           cfg.Traffic.QPS,
			Burst:         cfg.Traffic.Burst,
			RTATimeout:    cfg.Traffic.RTATimeout,
			BidTimeout:    cfg.Traffic.BidTimeout,
			MaxAdSlots:    cfg.Traffic.MaxAdSlots,
			MinAdSlotSize: cfg.Traffic.MinAdSlotSize,
			MaxAdSlotSize: cfg.Traffic.MaxAdSlotSize,
		},
	)

	// 7. 初始化HTTP服务器
	router := initRouter(trafficHandler, eventHandler)
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// 8. 启动服务器
	go func() {
		log.Info("启动HTTP服务器", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP服务器启动失败", "error", err)
		}
	}()

	// 9. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器关闭失败", "error", err)
	}
	log.Info("服务器已关闭")
}

// initLogger 初始化日志
func initLogger(cfg config.LogConfig) *logger.Logger {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
		os.Exit(1)
	}

	// 配置日志编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建日志核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(&logger.RotateWriter{
				Filename:   cfg.Filename,
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}),
		),
		zap.NewAtomicLevelAt(getLogLevel(cfg.Level)),
	)

	// 创建日志记录器
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger.NewLogger(zapLogger)
}

// initMetrics 初始化监控指标
func initMetrics(cfg config.MetricsConfig) *metrics.Metrics {
	if !cfg.Enabled {
		return metrics.NewNoopMetrics()
	}

	m := metrics.NewMetrics(cfg.Port, cfg.Path)
	if cfg.PushGateway != "" {
		go m.StartPushGateway(cfg.PushGateway)
	}
	return m
}

// initRedis 初始化Redis客户端
func initRedis(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addresses[0], // 使用第一个地址
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
		fmt.Printf("Redis连接失败: %v\n", err)
		os.Exit(1)
	}

	return client
}

// initKafka 初始化Kafka客户端
func initKafka(cfg config.KafkaConfig) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        "", // 动态设置
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  cfg.MaxRetries,
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	return writer
}

// initRouter 初始化路由
func initRouter(trafficHandler *traffic.Handler, eventHandler *event.Handler) *gin.Engine {
	router := gin.Default()

	// 流量接入接口
	router.POST("/api/v1/traffic", trafficHandler.HandleRequest)

	// 事件处理接口
	router.POST("/api/v1/events/impression", eventHandler.HandleImpression)
	router.POST("/api/v1/events/click", eventHandler.HandleClick)
	router.POST("/api/v1/events/conversion", eventHandler.HandleConversion)
	router.GET("/api/v1/events/stats", eventHandler.GetEventStats)

	// 健康检查接口
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
