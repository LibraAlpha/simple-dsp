package types

import "time"

// Creative 素材信息
type Creative struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	StoragePath string    `json:"storage_path"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}
