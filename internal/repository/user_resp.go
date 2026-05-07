package repository

import (
	"CloudStorage/internal/global"
	"CloudStorage/internal/model"
	"CloudStorage/internal/utils/redis"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type UserResp struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func CreateUser(user *model.User) (*UserResp, error) {
	ctx := context.Background()

	usernameKey := fmt.Sprintf("%s:%s", global.RedisRegisterKeyUser, user.Username)
	exists, _ := redis.RDU.Exist(ctx, usernameKey)
	if exists {
		return nil, errors.New("用户名已存在")
	}

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := checkUsernameExist(tx, user.Username); err != nil {
			return err
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return redis.RDU.Set(ctx, usernameKey, user.ID, global.RedisRegisterKeyUserExpireTime)
	})
	if err != nil {
		return nil, err
	}
	return &UserResp{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

func checkUsernameExist(tx *gorm.DB, username string) error {
	var count int64
	tx.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	return nil
}

func QueryUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := global.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
