package creative

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"simple-dsp/internal/creative/storage"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"

	"github.com/go-redis/redis/v8"
)

// Service 素材管理服务
type Service struct {
	redis   *redis.Client
	logger  *logger.Logger
	metrics *metrics.Metrics
	storage storage.Storage
}

// Creative 素材信息
type Creative struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`         // image, video, html
	Format      string    `json:"format"`       // jpg, png, mp4, etc.
	Size        int64     `json:"size"`         // 文件大小
	Width       int       `json:"width"`        // 宽度
	Height      int       `json:"height"`       // 高度
	Duration    float64   `json:"duration"`     // 视频时长
	URL         string    `json:"url"`          // 访问URL
	StoragePath string    `json:"storage_path"` // 存储路径
	Tags        []string  `json:"tags"`         // 标签
	Status      string    `json:"status"`       // active, inactive, deleted
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// CreativeGroup 素材组
type CreativeGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Creatives   []string  `json:"creatives"` // 素材ID列表
	Status      string    `json:"status"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// NewService 创建素材管理服务
func NewService(redis *redis.Client, logger *logger.Logger, metrics *metrics.Metrics, storage storage.Storage) *Service {
	return &Service{
		redis:   redis,
		logger:  logger,
		metrics: metrics,
		storage: storage,
	}
}

// UploadCreative 上传素材
func (s *Service) UploadCreative(ctx context.Context, file *multipart.FileHeader, tags []string) (*Creative, error) {
	// 生成素材ID
	id := generateID()

	// 获取文件信息
	filename := file.Filename
	size := file.Size
	format := filepath.Ext(filename)

	// 构建存储路径
	storagePath := fmt.Sprintf("creatives/%s/%s", time.Now().Format("20060102"), id+format)

	// 保存文件
	if err := s.storage.Save(ctx, storagePath, file); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 获取文件URL
	url, err := s.storage.GetURL(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件URL失败: %v", err)
	}

	// 创建素材信息
	creative := &Creative{
		ID:          id,
		Name:        filename,
		Type:        getCreativeType(format),
		Format:      format,
		Size:        size,
		URL:         url,
		StoragePath: storagePath,
		Tags:        tags,
		Status:      "active",
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	// 保存素材信息
	if err := s.saveCreative(ctx, creative); err != nil {
		return nil, fmt.Errorf("保存素材信息失败: %v", err)
	}

	// 更新指标
	s.metrics.Creative.Uploaded.Inc()
	s.metrics.Creative.Size.Observe(float64(size))

	return creative, nil
}

// DeleteCreative 删除素材
func (s *Service) DeleteCreative(ctx context.Context, id string) error {
	// 获取素材信息
	creative, err := s.GetCreative(ctx, id)
	if err != nil {
		return err
	}

	// 标记为删除状态
	creative.Status = "deleted"
	creative.UpdateTime = time.Now()

	// 保存更新
	if err := s.saveCreative(ctx, creative); err != nil {
		return err
	}

	// 删除存储文件
	if err := s.storage.Delete(ctx, creative.StoragePath); err != nil {
		s.logger.Error("删除存储文件失败", "error", err)
	}

	// 更新指标
	s.metrics.Creative.Deleted.Inc()

	return nil
}

// GetCreative 获取素材信息
func (s *Service) GetCreative(ctx context.Context, id string) (*Creative, error) {
	key := fmt.Sprintf("creative:%s", id)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("素材不存在")
		}
		return nil, err
	}

	var creative Creative
	if err := json.Unmarshal(data, &creative); err != nil {
		return nil, err
	}

	return &creative, nil
}

// ListCreatives 获取素材列表
func (s *Service) ListCreatives(ctx context.Context, tags []string) ([]*Creative, error) {
	var creatives []*Creative

	// 如果指定了标签，通过标签索引获取
	if len(tags) > 0 {
		for _, tag := range tags {
			key := fmt.Sprintf("creative:tag:%s", tag)
			ids, err := s.redis.SMembers(ctx, key).Result()
			if err != nil {
				continue
			}
			for _, id := range ids {
				if creative, err := s.GetCreative(ctx, id); err == nil && creative.Status != "deleted" {
					creatives = append(creatives, creative)
				}
			}
		}
		return creatives, nil
	}

	// 获取所有素材
	keys, err := s.redis.Keys(ctx, "creative:*").Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var creative Creative
		if err := json.Unmarshal(data, &creative); err != nil {
			continue
		}

		if creative.Status != "deleted" {
			creatives = append(creatives, &creative)
		}
	}

	return creatives, nil
}

// CreateGroup 创建素材组
func (s *Service) CreateGroup(ctx context.Context, group *CreativeGroup) error {
	// 生成组ID
	group.ID = generateID()
	group.CreateTime = time.Now()
	group.UpdateTime = time.Now()
	group.Status = "active"

	// 保存组信息
	if err := s.saveGroup(ctx, group); err != nil {
		return err
	}

	// 更新指标
	s.metrics.Creative.GroupCreated.Inc()

	return nil
}

// UpdateGroup 更新素材组
func (s *Service) UpdateGroup(ctx context.Context, group *CreativeGroup) error {
	// 获取现有组信息
	existingGroup, err := s.GetGroup(ctx, group.ID)
	if err != nil {
		return err
	}

	// 更新信息
	group.CreateTime = existingGroup.CreateTime
	group.UpdateTime = time.Now()

	// 保存更新
	if err := s.saveGroup(ctx, group); err != nil {
		return err
	}

	return nil
}

// DeleteGroup 删除素材组
func (s *Service) DeleteGroup(ctx context.Context, id string) error {
	// 获取组信息
	group, err := s.GetGroup(ctx, id)
	if err != nil {
		return err
	}

	// 标记为删除状态
	group.Status = "deleted"
	group.UpdateTime = time.Now()

	// 保存更新
	if err := s.saveGroup(ctx, group); err != nil {
		return err
	}

	// 更新指标
	s.metrics.Creative.GroupDeleted.Inc()

	return nil
}

// GetGroup 获取素材组信息
func (s *Service) GetGroup(ctx context.Context, id string) (*CreativeGroup, error) {
	key := fmt.Sprintf("creative:group:%s", id)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("素材组不存在")
		}
		return nil, err
	}

	var group CreativeGroup
	if err := json.Unmarshal(data, &group); err != nil {
		return nil, err
	}

	return &group, nil
}

// ListGroups 获取素材组列表
func (s *Service) ListGroups(ctx context.Context) ([]*CreativeGroup, error) {
	var groups []*CreativeGroup

	keys, err := s.redis.Keys(ctx, "creative:group:*").Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var group CreativeGroup
		if err := json.Unmarshal(data, &group); err != nil {
			continue
		}

		if group.Status != "deleted" {
			groups = append(groups, &group)
		}
	}

	return groups, nil
}

// 内部方法

func (s *Service) saveCreative(ctx context.Context, creative *Creative) error {
	data, err := json.Marshal(creative)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("creative:%s", creative.ID)
	if err := s.redis.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}

	// 更新标签索引
	for _, tag := range creative.Tags {
		tagKey := fmt.Sprintf("creative:tag:%s", tag)
		s.redis.SAdd(ctx, tagKey, creative.ID)
	}

	return nil
}

func (s *Service) saveGroup(ctx context.Context, group *CreativeGroup) error {
	data, err := json.Marshal(group)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("creative:group:%s", group.ID)
	return s.redis.Set(ctx, key, data, 0).Err()
}

func getCreativeType(format string) string {
	switch format {
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	case ".mp4", ".avi", ".mov":
		return "video"
	case ".html", ".htm":
		return "html"
	default:
		return "other"
	}
}

func generateID() string {
	return fmt.Sprintf("%d%06d", time.Now().Unix(), time.Now().Nanosecond()/1000)
}
