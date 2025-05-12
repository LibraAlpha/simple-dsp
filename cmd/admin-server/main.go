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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"simple-dsp/internal/admin"
	"simple-dsp/internal/budget"
	"simple-dsp/internal/frequency"
	"simple-dsp/internal/stats"
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

	// 5. 初始化动态配置管理器
	dynamicConfig := config.NewDynamicConfig(redisClient, log)

	// 6. 初始化配置管理服务
	configService := config.NewService(redisClient, log)
	configHandler := admin.NewConfigHandler(configService)

	// 7. 初始化各个模块
	// 7.1 初始化预算管理器
	budgetMgr := budget.NewManager(
		redisClient,
		log,
		metrics,
	)

	// 7.2 初始化数据统计服务
	statsService := stats.NewService(
		redisClient,
		log,
		metrics,
	)

	// 7.3 初始化频次控制器
	freqCtrl := frequency.NewController(
		redisClient,
		log,
		metrics,
	)

	// 7.4 初始化管理后台服务
	adminService := admin.NewService(
		budgetMgr,
		statsService,
		log,
		metrics,
		freqCtrl,
	)

	// 8. 初始化HTTP服务器
	router := initRouter(adminService, configHandler)
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Admin.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// 9. 启动服务器
	go func() {
		log.Info("启动管理后台服务器", "port", cfg.Admin.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("管理后台服务器启动失败", "error", err)
		}
	}()

	// 10. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭管理后台服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("管理后台服务器关闭失败", "error", err)
	}
	log.Info("管理后台服务器已关闭")
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
				Filename:   "logs/admin.log",
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

	m := metrics.NewMetrics(cfg.AdminPort, "/admin/metrics")
	if cfg.PushGateway != "" {
		go m.StartPushGateway(cfg.PushGateway)
	}
	return m
}

// initRedis 初始化Redis客户端
func initRedis(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addresses[0],
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

// initRouter 初始化路由
func initRouter(adminService *admin.Service, configHandler *admin.ConfigHandler) *gin.Engine {
	router := gin.Default()

	// 注册配置管理路由
	configHandler.RegisterRoutes(router)

	// 注册管理后台路由
	adminGroup := router.Group("/api/v1/admin")
	{
		adminGroup.GET("/stats/daily", adminService.GetDailyStats)
		adminGroup.GET("/stats/hourly", adminService.GetHourlyStats)
		adminGroup.GET("/system/status", adminService.GetSystemStatus)
		adminGroup.GET("/system/metrics", adminService.GetSystemMetrics)
	}

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