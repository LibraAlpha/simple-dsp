package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"simple-dsp/pkg/logger"
)

// ChunkInfo 分片信息
type ChunkInfo struct {
	UploadID   string    `json:"upload_id"`
	ChunkIndex int       `json:"chunk_index"`
	ChunkSize  int64     `json:"chunk_size"`
	ChunkPath  string    `json:"chunk_path"`
	CreateTime time.Time `json:"create_time"`
}

// ChunkUpload 分片上传信息
type ChunkUpload struct {
	UploadID    string    `json:"upload_id"`
	FileName    string    `json:"file_name"`
	TotalSize   int64     `json:"total_size"`
	ChunkSize   int64     `json:"chunk_size"`
	ChunkCount  int       `json:"chunk_count"`
	Status      string    `json:"status"`
	StoragePath string    `json:"storage_path"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// ChunkUploader 分片上传管理器
type ChunkUploader struct {
	redis   *redis.Client
	logger  *logger.Logger
	storage Storage
}

// NewChunkUploader 创建分片上传管理器
func NewChunkUploader(redis *redis.Client, logger *logger.Logger, storage Storage) *ChunkUploader {
	return &ChunkUploader{
		redis:   redis,
		logger:  logger,
		storage: storage,
	}
}

// InitUpload 初始化分片上传
func (cu *ChunkUploader) InitUpload(ctx context.Context, fileName string, totalSize int64, chunkSize int64) (*ChunkUpload, error) {
	uploadID := generateUploadID()
	chunkCount := (totalSize + chunkSize - 1) / chunkSize

	upload := &ChunkUpload{
		UploadID:    uploadID,
		FileName:    fileName,
		TotalSize:   totalSize,
		ChunkSize:   chunkSize,
		ChunkCount:  int(chunkCount),
		Status:      "uploading",
		StoragePath: fmt.Sprintf("uploads/%s/%s", time.Now().Format("20060102"), uploadID),
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	if err := cu.saveUpload(ctx, upload); err != nil {
		return nil, err
	}

	return upload, nil
}

// UploadChunk 上传分片
func (cu *ChunkUploader) UploadChunk(ctx context.Context, uploadID string, chunkIndex int, reader io.Reader) error {
	// 获取上传信息
	upload, err := cu.GetUpload(ctx, uploadID)
	if err != nil {
		return err
	}

	// 验证分片索引
	if chunkIndex < 0 || chunkIndex >= upload.ChunkCount {
		return ErrInvalidChunkIndex
	}

	// 保存分片
	chunkPath := fmt.Sprintf("%s/chunk_%d", upload.StoragePath, chunkIndex)
	if err := cu.storage.SaveStream(ctx, chunkPath, reader); err != nil {
		return err
	}

	// 记录分片信息
	chunk := &ChunkInfo{
		UploadID:   uploadID,
		ChunkIndex: chunkIndex,
		ChunkSize:  upload.ChunkSize,
		ChunkPath:  chunkPath,
		CreateTime: time.Now(),
	}

	if err := cu.saveChunk(ctx, chunk); err != nil {
		return err
	}

	return nil
}

// CompleteUpload 完成上传
func (cu *ChunkUploader) CompleteUpload(ctx context.Context, uploadID string) (string, error) {
	// 获取上传信息
	upload, err := cu.GetUpload(ctx, uploadID)
	if err != nil {
		return "", err
	}

	// 获取所有分片
	chunks, err := cu.listChunks(ctx, uploadID)
	if err != nil {
		return "", err
	}

	// 验证分片完整性
	if len(chunks) != upload.ChunkCount {
		return "", ErrIncompleteUpload
	}

	// 按索引排序分片
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkIndex < chunks[j].ChunkIndex
	})

	// 合并分片
	finalPath := filepath.Join("creatives", time.Now().Format("20060102"), filepath.Base(upload.FileName))
	if err := cu.mergeChunks(ctx, chunks, finalPath); err != nil {
		return "", err
	}

	// 清理分片
	if err := cu.cleanupChunks(ctx, upload); err != nil {
		cu.logger.Error("清理分片失败", "error", err)
	}

	return finalPath, nil
}

// GetUpload 获取上传信息
func (cu *ChunkUploader) GetUpload(ctx context.Context, uploadID string) (*ChunkUpload, error) {
	key := cu.getUploadKey(uploadID)
	data, err := cu.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrUploadNotFound
		}
		return nil, err
	}

	var upload ChunkUpload
	if err := json.Unmarshal(data, &upload); err != nil {
		return nil, err
	}

	return &upload, nil
}

// 内部方法

func (cu *ChunkUploader) saveUpload(ctx context.Context, upload *ChunkUpload) error {
	data, err := json.Marshal(upload)
	if err != nil {
		return err
	}

	key := cu.getUploadKey(upload.UploadID)
	return cu.redis.Set(ctx, key, data, 24*time.Hour).Err()
}

func (cu *ChunkUploader) saveChunk(ctx context.Context, chunk *ChunkInfo) error {
	data, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	key := cu.getChunkKey(chunk.UploadID, chunk.ChunkIndex)
	return cu.redis.Set(ctx, key, data, 24*time.Hour).Err()
}

func (cu *ChunkUploader) listChunks(ctx context.Context, uploadID string) ([]*ChunkInfo, error) {
	pattern := cu.getChunkPattern(uploadID)
	keys, err := cu.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var chunks []*ChunkInfo
	for _, key := range keys {
		data, err := cu.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var chunk ChunkInfo
		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}

		chunks = append(chunks, &chunk)
	}

	return chunks, nil
}

func (cu *ChunkUploader) mergeChunks(ctx context.Context, chunks []*ChunkInfo, finalPath string) error {
	return cu.storage.MergeFiles(ctx, finalPath, chunks)
}

func (cu *ChunkUploader) cleanupChunks(ctx context.Context, upload *ChunkUpload) error {
	// 删除分片文件
	if err := cu.storage.DeleteDir(ctx, upload.StoragePath); err != nil {
		return err
	}

	// 删除Redis记录
	pattern := cu.getChunkPattern(upload.UploadID)
	keys, err := cu.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return cu.redis.Del(ctx, keys...).Err()
	}

	return nil
}

func (cu *ChunkUploader) getUploadKey(uploadID string) string {
	return fmt.Sprintf("upload:%s", uploadID)
}

func (cu *ChunkUploader) getChunkKey(uploadID string, chunkIndex int) string {
	return fmt.Sprintf("upload:%s:chunk:%d", uploadID, chunkIndex)
}

func (cu *ChunkUploader) getChunkPattern(uploadID string) string {
	return fmt.Sprintf("upload:%s:chunk:*", uploadID)
}

func generateUploadID() string {
	return fmt.Sprintf("%d%06d", time.Now().Unix(), time.Now().Nanosecond()/1000)
} 