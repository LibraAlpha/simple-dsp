package creative

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"simple-dsp/pkg/logger"
)

// Version 素材版本
type Version struct {
	ID          string    `json:"id"`
	CreativeID  string    `json:"creative_id"`
	Version     int       `json:"version"`
	Changes     string    `json:"changes"`
	StoragePath string    `json:"storage_path"`
	Status      string    `json:"status"`
	Creator     string    `json:"creator"`
	CreateTime  time.Time `json:"create_time"`
}

// VersionService 版本控制服务
type VersionService struct {
	redis   *redis.Client
	logger  *logger.Logger
	storage Storage
}

// NewVersionService 创建版本控制服务
func NewVersionService(redis *redis.Client, logger *logger.Logger, storage Storage) *VersionService {
	return &VersionService{
		redis:   redis,
		logger:  logger,
		storage: storage,
	}
}

// CreateVersion 创建新版本
func (vs *VersionService) CreateVersion(ctx context.Context, creative *Creative, changes string, creator string) (*Version, error) {
	// 获取当前版本号
	currentVersion, err := vs.getCurrentVersion(ctx, creative.ID)
	if err != nil {
		currentVersion = 0
	}

	// 创建新版本
	version := &Version{
		ID:          generateID(),
		CreativeID:  creative.ID,
		Version:     currentVersion + 1,
		Changes:     changes,
		StoragePath: creative.StoragePath,
		Status:      "active",
		Creator:     creator,
		CreateTime:  time.Now(),
	}

	// 保存版本信息
	if err := vs.saveVersion(ctx, version); err != nil {
		return nil, err
	}

	return version, nil
}

// GetVersion 获取指定版本
func (vs *VersionService) GetVersion(ctx context.Context, creativeID string, version int) (*Version, error) {
	key := vs.getVersionKey(creativeID, version)
	data, err := vs.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}

	var v Version
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

// ListVersions 获取版本列表
func (vs *VersionService) ListVersions(ctx context.Context, creativeID string) ([]*Version, error) {
	currentVersion, err := vs.getCurrentVersion(ctx, creativeID)
	if err != nil {
		return nil, err
	}

	var versions []*Version
	for i := 1; i <= currentVersion; i++ {
		version, err := vs.GetVersion(ctx, creativeID, i)
		if err != nil {
			vs.logger.Error("获取版本失败", "error", err)
			continue
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// RollbackVersion 回滚到指定版本
func (vs *VersionService) RollbackVersion(ctx context.Context, creativeID string, version int) error {
	// 获取目标版本
	targetVersion, err := vs.GetVersion(ctx, creativeID, version)
	if err != nil {
		return err
	}

	// 获取素材信息
	creative, err := vs.storage.GetCreative(ctx, creativeID)
	if err != nil {
		return err
	}

	// 更新素材信息
	creative.StoragePath = targetVersion.StoragePath
	creative.UpdateTime = time.Now()

	// 保存更新
	if err := vs.storage.SaveCreative(ctx, creative); err != nil {
		return err
	}

	return nil
}

// 内部方法

func (vs *VersionService) getCurrentVersion(ctx context.Context, creativeID string) (int, error) {
	key := vs.getCurrentVersionKey(creativeID)
	version, err := vs.redis.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return version, nil
}

func (vs *VersionService) saveVersion(ctx context.Context, version *Version) error {
	data, err := json.Marshal(version)
	if err != nil {
		return err
	}

	pipe := vs.redis.Pipeline()

	// 保存版本信息
	pipe.Set(ctx, vs.getVersionKey(version.CreativeID, version.Version), data, 0)

	// 更新当前版本号
	pipe.Set(ctx, vs.getCurrentVersionKey(version.CreativeID), version.Version, 0)

	_, err = pipe.Exec(ctx)
	return err
}

func (vs *VersionService) getVersionKey(creativeID string, version int) string {
	return fmt.Sprintf("creative:version:%s:%d", creativeID, version)
}

func (vs *VersionService) getCurrentVersionKey(creativeID string) string {
	return fmt.Sprintf("creative:current_version:%s", creativeID)
} 