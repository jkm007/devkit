package model

import "time"

// MediaAsset 媒体资产（视频/音频元数据）
type MediaAsset struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FileAssetID     uint      `gorm:"uniqueIndex;comment:关联文件资产" json:"fileAssetId"`
	FileType        string    `gorm:"size:20;comment:媒体类型(video/audio/image)" json:"fileType"`
	Duration        float64   `gorm:"comment:时长(秒)" json:"duration"`
	Width           int       `gorm:"comment:视频宽度" json:"width"`
	Height          int       `gorm:"comment:视频高度" json:"height"`
	Bitrate         int       `gorm:"comment:比特率" json:"bitrate"`
	Title           string    `gorm:"size:255;comment:标题(音频)" json:"title"`
	Artist          string    `gorm:"size:255;comment:艺术家(音频)" json:"artist"`
	Album           string    `gorm:"size:255;comment:专辑(音频)" json:"album"`
	Genre           string    `gorm:"size:64;comment:流派(音频)" json:"genre"`
	Year            int       `gorm:"comment:年份(音频)" json:"year"`
	TranscodeStatus string    `gorm:"size:20;default:none;comment:转码状态" json:"transcodeStatus"` // none, pending, processing, completed, failed
	HLSPath         string    `gorm:"size:500;comment:HLS路径" json:"hlsPath"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (MediaAsset) TableName() string { return "sys_media_assets" }
