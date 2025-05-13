package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simple-dsp/internal/bidding"
	pb "simple-dsp/api/proto/dsp/v1"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// 初始化日志
	log := logger.NewLogger()

	// 初始化指标收集
	metrics := metrics.NewMetrics()

	// 创建竞价引擎
	engine := bidding.NewEngine(
		repository,
		budgetMgr,
		freqCtrl,
		log,
		metrics,
	)

	// 创建 HTTP 服务器
	router := gin.Default()
	// ... 配置 HTTP 路由

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()
	pb.RegisterBidServiceServer(grpcServer, bidding.NewGRPCServer(engine, log))

	// 启动 HTTP 服务器
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", "error", err)
		}
	}()

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal("Failed to listen", "error", err)
	}
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC", "error", err)
		}
	}()

	log.Info("Server started",
		"http_port", 8080,
		"grpc_port", 9090)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	// 关闭 gRPC 服务器
	grpcServer.GracefulStop()

	log.Info("Server exited")
} 