package clients

import (
	"context"
	"github.com/klauspost/compress/gzhttp/writer"
	"time"

	"github.com/segmentio/kafka-go"
	"simple-dsp/pkg/config"
	"simple-dsp/pkg/logger"
)

type KafkaClientAdapter struct {
	writer *kafka.Writer
	log    *logger.Logger
}

// WriteMessages 实现接口方法
func (a *KafkaClientAdapter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return a.writer.WriteMessages(ctx, msgs...)
}

func (a *KafkaClientAdapter) SendMessage(ctx context.Context, topic string, key string, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}
	return a.writer.WriteMessages(ctx, msg)
}

func (a *KafkaClientAdapter) Close() error {
	return a.writer.Close()
}

// NewKafkaClient 创建Kafka客户端
func NewKafkaClient(cfg config.KafkaConfig, log *logger.Logger) (*KafkaClientAdapter, error) {
	baseWriter := &kafka.Writer{
		Addr:        kafka.TCP(cfg.Brokers...),
		Topic:       cfg.Topic,
		Balancer:    &kafka.LeastBytes{},
		MaxAttempts: cfg.MaxRetries,
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := writer.GzipWriter(ctx, kafka.Message{}); err != nil && err != kafka.ErrInvalidMessage {
		log.Error("Kafka连接失败", "error", err)
		return nil, err
	}

	log.Info("Kafka连接成功", "brokers", cfg.Brokers)
	return &KafkaClientAdapter{writer: baseWriter, log: log}, nil
}
