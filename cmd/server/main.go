package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simple-dsp/internal/bidding"
	"simple-dsp/internal/budget"
	"simple-dsp/internal/frequency"
	"simple-dsp/internal/repository"
	pb "simple-dsp/api/proto/dsp/v1"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
	"simple-dsp/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	if err := config.LoadConfig(*configPath); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	cfg := config.GetConfig()

	// 初始化日志
	log, err := logger.NewLogger(cfg.Log)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// 初始化指标收集
	metrics, err := metrics.NewMetrics(cfg.Metrics)
	if err != nil {
		log.Fatal("初始化指标收集失败", zap.Error(err))
	}

	// 初始化数据库连接
	db, err := initDB(cfg)
	if err != nil {
		log.Fatal("初始化数据库失败", zap.Error(err))
	}

	// 初始化Redis客户端
	rdb, err := initRedis(cfg)
	if err != nil {
		log.Fatal("初始化Redis失败", zap.Error(err))
	}

	// 初始化仓储层
	repo := repository.NewRepository(db)

	// 初始化预算管理器
	budgetMgr := budget.NewManager(rdb, log)

	// 初始化频次控制器
	freqCtrl := frequency.NewController(rdb, log)

	// 创建竞价引擎
	engine := bidding.NewEngine(
		repo,
		budgetMgr,
		freqCtrl,
		log,
		metrics,
	)

	// 创建 HTTP 服务器
	router := gin.New()
	
	// 添加中间件
	router.Use(
		gin.Recovery(),
		middleware.Logger(log),
		middleware.Metrics(metrics),
		middleware.RateLimit(cfg.Traffic.QPS, cfg.Traffic.Burst),
	)

	// 注册路由
	registerRoutes(router, engine, log)

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCMetrics(metrics)),
	)
	pb.RegisterBidServiceServer(grpcServer, bidding.NewGRPCServer(engine, log))

	// 启动 HTTP 服务器
	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP服务器错误", zap.Error(err))
		}
	}()

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal("gRPC监听失败", zap.Error(err))
	}
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("gRPC服务失败", zap.Error(err))
		}
	}()

	log.Info("服务启动成功",
		zap.Int("http_port", cfg.Server.Port),
		zap.Int("grpc_port", 9090))

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务器...")

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	
	// 关闭主HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP服务器强制关闭", zap.Error(err))
	}

	// 关闭 gRPC 服务器
	grpcServer.GracefulStop()

	// 关闭 metrics 服务器
	if err := metrics.Close(); err != nil {
		log.Error("Metrics服务器关闭失败", zap.Error(err))
	}

	log.Info("服务器已退出")
	
	// 确保所有日志都已写入
	_ = log.Sync()
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	// TODO: 实现数据库初始化
	return nil, nil
}

func initRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addresses[0],
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

func registerRoutes(r *gin.Engine, engine *bidding.Engine, log *zap.Logger) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API 版本组
	v1 := r.Group("/api/v1")
	{
		// 竞价接口
		v1.POST("/bid", handleBid(engine, log))
		
		// 竞价获胜通知
		v1.POST("/win", handleWin(engine, log))
		
		// 曝光回调
		v1.POST("/impression", handleImpression(engine, log))
		
		// 点击回调
		v1.POST("/click", handleClick(engine, log))
	}
}

// 处理函数定义
func handleBid(engine *bidding.Engine, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现竞价处理逻辑
	}
}

func handleWin(engine *bidding.Engine, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现竞价获胜处理逻辑
	}
}

func handleImpression(engine *bidding.Engine, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现曝光处理逻辑
	}
}

func handleClick(engine *bidding.Engine, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现点击处理逻辑
	}
} 