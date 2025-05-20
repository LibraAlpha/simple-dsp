/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: main.go
 * Project: simple-dsp
 * Description: DSP服务器主程序，负责初始化和启动DSP服务
 *
 * 主要功能:
 * - 初始化系统配置和日志
 * - 初始化各类客户端连接（Redis、Kafka等）
 * - 初始化业务模块（竞价引擎、预算管理等）
 * - 启动HTTP服务器和优雅关闭
 *
 * 实现细节:
 * - 使用gin框架提供HTTP服务
 * - 实现优雅启动和关闭
 * - 统一的错误处理和日志记录
 * - 模块化的服务初始化
 *
 * 依赖关系:
 * - github.com/gin-gonic/gin
 * - simple-dsp/internal/* (所有内部业务模块)
 * - simple-dsp/pkg/* (所有基础包)
 *
 * 注意事项:
 * - 需要正确配置所有依赖服务
 * - 注意处理服务优雅关闭
 * - 合理设置超时参数
 * - 注意资源释放和错误处理
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"simple-dsp/pkg/clients"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"

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

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

func main() {
	// 初始化配置
	cfg := config.GetConfig()

	// 初始化日志
	log, err := logger.NewLoggerFromConfig(cfg.Log)

	if err != nil {
		log.Fatal("加载配置失败", "error", err)
	}

	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := log.Sync(); err != nil {
			log.Error("同步日志失败", "error", err)
		}
	}()

	// 初始化监控指标
	metricsCollector, err := metrics.NewMetrics(cfg.Metrics)
	if cfg.Metrics.PushGateway != "" {
		metricsCollector.StartPushGateway(cfg.Metrics.PushGateway)
	}

	// 初始化Redis客户端
	redisClient, err := clients.InitRedis(cfg, log)
	defer func(redisClient *redis.Client) {
		err := redisClient
		if err != nil {

		}
	}(redisClient)

	// 初始化Kafka客户端
	kafkaClient := clients.InitKafka(cfg.Kafka, log)
	defer func(kafkaClient *kafka.Writer) {
		err := kafkaClient.Close()
		if err != nil {

		}
	}(kafkaClient)

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
		log.Info("启动DSP服务器", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("DSP服务器启动失败", "error", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭DSP服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("DSP服务器关闭失败", "error", err)
	}
	log.Info("DSP服务器已关闭")
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
