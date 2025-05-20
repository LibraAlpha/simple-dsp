package creative

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"simple-dsp/internal/creative/storage"
	"simple-dsp/pkg/logger"

	"github.com/go-redis/redis/v8"
)

// AuditStatus 审核状态
type AuditStatus string

const (
	AuditStatusPending  AuditStatus = "pending"
	AuditStatusApproved AuditStatus = "approved"
	AuditStatusRejected AuditStatus = "rejected"
	AuditStatusRevision AuditStatus = "revision"
)

// AuditRecord 审核记录
type AuditRecord struct {
	ID         string      `json:"id"`
	CreativeID string      `json:"creative_id"`
	Status     AuditStatus `json:"status"`
	Reviewer   string      `json:"reviewer"`
	Comments   string      `json:"comments"`
	CreateTime time.Time   `json:"create_time"`
	UpdateTime time.Time   `json:"update_time"`
}

// AuditService 审核服务
type AuditService struct {
	redis   *redis.Client
	logger  *logger.Logger
	storage storage.Storage
}

// NewAuditService 创建审核服务
func NewAuditService(redis *redis.Client, logger *logger.Logger, storage storage.Storage) *AuditService {
	return &AuditService{
		redis:   redis,
		logger:  logger,
		storage: storage,
	}
}

// SubmitForAudit 提交审核
func (as *AuditService) SubmitForAudit(ctx context.Context, creativeID string) error {
	record := &AuditRecord{
		ID:         generateID(),
		CreativeID: creativeID,
		Status:     AuditStatusPending,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	return as.saveAuditRecord(ctx, record)
}

// ReviewCreative 审核素材
func (as *AuditService) ReviewCreative(ctx context.Context, creativeID string, status AuditStatus, reviewer string, comments string) error {
	record, err := as.GetLatestAuditRecord(ctx, creativeID)
	if err != nil {
		return err
	}

	record.Status = status
	record.Reviewer = reviewer
	record.Comments = comments
	record.UpdateTime = time.Now()

	if err := as.saveAuditRecord(ctx, record); err != nil {
		return err
	}

	// 更新素材状态
	return as.updateCreativeStatus(ctx, creativeID, status)
}

// GetLatestAuditRecord 获取最新审核记录
func (as *AuditService) GetLatestAuditRecord(ctx context.Context, creativeID string) (*AuditRecord, error) {
	key := as.getAuditKey(creativeID)
	data, err := as.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("audit record not found")
		}
		return nil, err
	}

	var record AuditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// GetAuditHistory 获取审核历史
func (as *AuditService) GetAuditHistory(ctx context.Context, creativeID string) ([]*AuditRecord, error) {
	key := as.getAuditHistoryKey(creativeID)
	data, err := as.redis.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var records []*AuditRecord
	for _, item := range data {
		var record AuditRecord
		if err := json.Unmarshal([]byte(item), &record); err != nil {
			as.logger.Error("解析审核记录失败", "error", err)
			continue
		}
		records = append(records, &record)
	}

	return records, nil
}

// 内部方法

func (as *AuditService) saveAuditRecord(ctx context.Context, record *AuditRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	pipe := as.redis.Pipeline()

	// 保存最新记录
	pipe.Set(ctx, as.getAuditKey(record.CreativeID), data, 0)

	// 添加到历史记录
	pipe.LPush(ctx, as.getAuditHistoryKey(record.CreativeID), data)
	pipe.LTrim(ctx, as.getAuditHistoryKey(record.CreativeID), 0, 99) // 保留最近100条记录

	_, err = pipe.Exec(ctx)
	return err
}

func (as *AuditService) updateCreativeStatus(ctx context.Context, creativeID string, status AuditStatus) error {
	creative, err := as.storage.GetCreative(ctx, creativeID)
	if err != nil {
		return err
	}

	switch status {
	case AuditStatusApproved:
		creative.Status = "active"
	case AuditStatusRejected:
		creative.Status = "rejected"
	case AuditStatusRevision:
		creative.Status = "revision"
	}

	return as.storage.SaveCreative(ctx, creative)
}

func (as *AuditService) getAuditKey(creativeID string) string {
	return "creative:audit:" + creativeID
}

func (as *AuditService) getAuditHistoryKey(creativeID string) string {
	return "creative:audit:history:" + creativeID
}
