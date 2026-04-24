package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null;type:varchar(255);index" json:"username"`
	Password string `gorm:"not null;type:varchar(255)" json:"password"`
}

func (User) TableName() string {
	return "users"
}
