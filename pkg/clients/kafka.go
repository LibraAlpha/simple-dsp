/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: kafka.go
 * Project: simple-dsp
 * Description: Kafka客户端封装，提供统一的消息队列操作接口
 *
 * 主要功能:
 * - 提供统一的Kafka客户端接口定义
 * - 封装消息生产和发送操作
 * - 支持消息压缩和批量发送
 * - 提供连接管理和错误处理
 *
 * 实现细节:
 * - 使用kafka-go库实现底层连接
 * - 通过适配器模式封装原生客户端
 * - 支持消息压缩和批处理
 * - 实现标准的消息发送接口
 *
 * 依赖关系:
 * - github.com/segmentio/kafka-go
 * - github.com/klauspost/compress/gzhttp/writer
 * - simple-dsp/pkg/config
 * - simple-dsp/pkg/logger
 *
 * 注意事项:
 * - 需要正确配置Kafka连接参数
 * - 注意处理消息发送超时
 * - 合理设置批处理参数
 * - 注意处理连接资源释放
 */

package clients

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
)

// KafkaClient Kafka客户端接口
type KafkaClient interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
	Close() error
}

// InitKafka 初始化Kafka客户端
func InitKafka(cfg config.KafkaConfig, log *logger.Logger) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:        kafka.TCP(cfg.Brokers...),
		Topic:       cfg.Topic,
		Balancer:    &kafka.LeastBytes{},
		MaxAttempts: cfg.MaxRetries,
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := writer.WriteMessages(ctx, kafka.Message{}); err != nil {
		log.Fatal("Kafka连接失败", "error", err)
	}

	return writer
}
