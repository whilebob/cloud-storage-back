package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username  string `gorm:"unique;not null;type:varchar(255);index" json:"username"`
	Password  string `gorm:"not null;type:varchar(255)" json:"password"`
	TotalSize uint64 `gorm:"default:5368709120" json:"total_size"`   //默认5GB
	UsedSize  uint64 `gorm:"default:0" json:"used_size"`             //已用空间
	Salt      string `gorm:"not null;type:varchar(255)" json:"salt"` //盐值
}

func (User) TableName() string {
	return "users"
}
