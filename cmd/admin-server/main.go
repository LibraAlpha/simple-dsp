/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: main.go
 * Project: simple-dsp
 * Description: 管理后台服务器主程序，提供系统管理和监控功能
 * 
 * 主要功能:
 * - 提供系统配置管理接口
 * - 提供数据统计和监控接口
 * - 提供预算和频次控制管理
 * - 提供系统状态查询接口
 * 
 * 实现细节:
 * - 使用gin框架提供HTTP服务
 * - 实现配置的动态管理
 * - 提供实时监控指标
 * - 支持系统状态查询
 * 
 * 依赖关系:
 * - github.com/gin-gonic/gin
 * - simple-dsp/internal/admin
 * - simple-dsp/internal/budget
 * - simple-dsp/internal/stats
 * - simple-dsp/pkg/* (所有基础包)
 * 
 * 注意事项:
 * - 需要正确配置管理权限
 * - 注意保护敏感配置信息
 * - 合理设置接口访问限制
 * - 注意处理并发访问
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"simple-dsp/internal/admin"
	"simple-dsp/internal/budget"
	"simple-dsp/internal/frequency"
	"simple-dsp/internal/stats"
	"simple-dsp/pkg/clients"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

func main() {
	// 1. 加载配置
	cfg := config.GetConfig()

	// 2. 初始化日志
	log, err := logger.NewLoggerFromConfig(cfg.Log)
	if err != nil {
		log.Fatal("初始化日志失败", "error", err)
	}
	defer log.Sync()

	// 3. 初始化监控指标
	metricsCollector := metrics.NewMetrics(cfg.Metrics.Port, cfg.Metrics.Path)
	if cfg.Metrics.PushGateway != "" {
		metricsCollector.StartPushGateway(cfg.Metrics.PushGateway)
	}

	// 4. 初始化Redis客户端
	redisClient, err := clients.NewRedisClient(cfg.Redis, log)
	if err != nil {
		log.Fatal("初始化Redis客户端失败", "error", err)
	}
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
		metricsCollector,
	)

	// 7.2 初始化数据统计服务
	statsService := stats.NewService(
		redisClient,
		log,
		metricsCollector,
	)

	// 7.3 初始化频次控制器
	freqCtrl := frequency.NewController(
		redisClient,
		log,
		metricsCollector,
	)

	// 7.4 初始化管理后台服务
	adminService := admin.NewService(
		budgetMgr,
		statsService,
		log,
		metricsCollector,
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