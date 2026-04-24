package model

import "gorm.io/gorm"

// Chunk 文件分片（大文件）
type Chunk struct {
	gorm.Model
	MD5         string `gorm:"type:varchar(255);not null" json:"md5"`
	FileName    string `gorm:"type:varchar(255);not null" json:"file_name"`
	ChunkIndex  int    `gorm:"not null" json:"chunk_index"`
	TotalChunks int    `gorm:"not null" json:"total_chunks"`
	IsUploaded  bool   `gorm:"not null;default:false" json:"is_uploaded"`
	MinioURL    string `gorm:"type:varchar(500);not null" json:"minio_url"`

	Username string `gorm:"ype:varchar(255);not null;index" json:"username"`
	User     User   `gorm:"foreignKey:Username;references:Username" json:"user"`
}

type File struct {
	gorm.Model
	FileName   string `gorm:"type:varchar(255);not null" json:"file_name"`
	IsUploaded bool   `gorm:"not null;default:false" json:"is_uploaded"`
	Md5        string `gorm:"type:varchar(255);not null" json:"md5"`
	MinioURL   string `gorm:"type:varchar(500);not null" json:"minio_url"`
	Size       int64  `gorm:"not null;default:0" json:"size"`

	Username string `gorm:"type:varchar(255);not null;index" json:"username"`
	User     User   `gorm:"foreignKey:Username;references:Username" json:"user"`
}

func (File) TableName() string {
	return "files"
}
func (Chunk) TableName() string {
	return "chunks"
}
