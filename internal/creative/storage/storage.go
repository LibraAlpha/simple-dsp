package storage

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"simple-dsp/internal/creative/types"
)

// Storage 存储接口
type Storage interface {
	// SaveStream 保存流数据
	SaveStream(ctx context.Context, path string, reader io.Reader) error
	// MergeFiles 合并文件
	MergeFiles(ctx context.Context, finalPath string, chunks []*ChunkInfo) error
	// DeleteDir 删除目录
	DeleteDir(ctx context.Context, path string) error
	// GetCreative 获取素材信息
	GetCreative(ctx context.Context, creativeID string) (*types.Creative, error)
	// SaveCreative 保存素材信息
	SaveCreative(ctx context.Context, creative *types.Creative) error
	// Save 保存文件
	Save(ctx context.Context, path string, file *multipart.FileHeader) error
	// GetURL 获取文件URL
	GetURL(ctx context.Context, path string) (string, error)
	// Delete 删除文件
	Delete(ctx context.Context, path string) error
}

// 错误定义
var (
	ErrInvalidChunkIndex = errors.New("无效的分片索引")
	ErrIncompleteUpload  = errors.New("上传未完成")
	ErrUploadNotFound    = errors.New("上传记录不存在")
)
