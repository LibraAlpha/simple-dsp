/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: postgres.go
 * Project: simple-dsp
 * Description: PostgreSQL客户端封装，提供统一的数据库操作接口
 * 
 * 主要功能:
 * - 提供统一的PostgreSQL客户端接口定义
 * - 封装数据库查询和事务操作
 * - 支持连接池管理
 * - 提供错误处理和日志记录
 * 
 * 实现细节:
 * - 使用database/sql标准库实现底层连接
 * - 支持事务和预处理语句
 * - 实现连接池参数配置
 * - 提供标准的数据库操作接口
 * 
 * 依赖关系:
 * - database/sql
 * - github.com/lib/pq
 * - simple-dsp/pkg/config
 * - simple-dsp/pkg/logger
 * 
 * 注意事项:
 * - 需要正确配置数据库连接参数
 * - 注意处理连接池资源管理
 * - 合理设置超时参数
 * - 注意处理事务提交和回滚
 */

package clients

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
)

// PostgresClient PostgreSQL客户端接口
type PostgresClient interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

// NewPostgresClient 创建PostgreSQL客户端
func NewPostgresClient(cfg config.PostgresConfig, log *logger.Logger) (PostgresClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error("PostgreSQL连接失败", "error", err)
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Error("PostgreSQL连接测试失败", "error", err)
		return nil, err
	}

	log.Info("PostgreSQL连接成功", "host", cfg.Host, "port", cfg.Port)
	return db, nil
} 