/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: logger.go
 * Project: simple-dsp
 * Description: 日志管理模块，提供统一的日志记录功能
 *
 * 主要功能:
 * - 提供日志记录接口
 * - 支持多级别日志
 * - 实现日志轮转
 * - 提供结构化日志
 *
 * 实现细节:
 * - 使用zap实现日志记录
 * - 支持JSON和文本格式
 * - 实现日志分级输出
 * - 提供日志采样功能
 *
 * 依赖关系:
 * - go.uber.org/zap
 * - simple-dsp/pkg/config
 *
 * 注意事项:
 * - 注意日志性能影响
 * - 合理设置日志级别
 * - 注意日志文件管理
 * - 确保日志安全性
 */

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"simple-dsp/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 是日志记录器的包装结构体
type Logger struct {
	*zap.Logger
}

// NewLogger 创建一个新的日志记录器
func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{zapLogger}
}

// NewLoggerFromConfig 从配置创建新的日志记录器
func NewLoggerFromConfig(cfg config.LogConfig) (*Logger, error) {
	// 创建基础配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置日志级别
	var level zapcore.Level
	err := level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, fmt.Errorf("无效的日志级别: %v", err)
	}

	// 确保日志目录存在
	if cfg.Filename != "" {
		logDir := filepath.Dir(cfg.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 创建Core
	var core zapcore.Core
	if cfg.Filename == "" {
		// 如果没有指定文件名，仅输出到控制台
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
	} else {
		// 同时输出到文件和控制台
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,    // 每个日志文件最大尺寸（MB）
			MaxBackups: cfg.MaxBackups, // 保留的旧日志文件最大数量
			MaxAge:     cfg.MaxAge,     // 保留的旧日志文件最大天数
			Compress:   cfg.Compress,   // 是否压缩旧日志文件
		})

		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

		core = zapcore.NewTee(
			zapcore.NewCore(jsonEncoder, fileWriter, level),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
		)
	}

	// 创建Logger
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &Logger{zapLogger}, nil
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Debugw(msg, keysAndValues...)
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Infow(msg, keysAndValues...)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Warnw(msg, keysAndValues...)
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Errorw(msg, keysAndValues...)
}

// Fatal 记录致命错误级别日志并退出程序
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Fatalw(msg, keysAndValues...)
}

// Sync 同步日志缓冲区
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// With 返回带有额外字段的日志记录器
func (l *Logger) With(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Debugw(msg, keysAndValues...)
}
