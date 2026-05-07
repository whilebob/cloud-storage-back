package service

import (
	"CloudStorage/config"
	"CloudStorage/internal/common"
	"CloudStorage/internal/dto"
	"CloudStorage/internal/model"
	"CloudStorage/internal/repository"
	"CloudStorage/internal/utils"
	"errors"

	"go.uber.org/zap"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) Register(user *dto.UserRegisterDTO) (*dto.RegisterResponse, error) {
	// 加密密码
	hashedPassword, salt := common.HashPassword(user.Password)
	// 创建用户
	resp, err := repository.CreateUser(&model.User{
		Username: user.Username,
		Password: hashedPassword,
		Salt:     salt,
	})
	if err != nil {
		utils.Logger.Error("创建用户失败", zap.Error(err))
		return nil, err
	}
	return &dto.RegisterResponse{
		UserInfo: dto.UserInfo{
			ID:       int(resp.ID),
			Username: resp.Username,
		},
	}, nil
}

func (u *UserService) Login(userdto *dto.UserLoginDTO) (*dto.LoginResponse, error) {
	user, err := repository.QueryUserByUsername(userdto.Username)
	if err != nil {
		utils.Logger.Error("查询用户失败", zap.Error(err))
		return nil, err
	}

	// 验证密码（假设密码已加密存储）
	if !common.VerifyPassword(userdto.Password, user.Password, user.Salt) {
		return nil, errors.New("密码错误")
	}

	accessToken := common.GenAccessToken(int(user.ID), user.Username)
	refreshToken := common.GenRefreshToken(int(user.ID), user.Username)

	return &dto.LoginResponse{
		AccessToken:          accessToken,
		RefreshToken:         refreshToken,
		AccessTokenExpiresIn: int64(config.AppConfig.JWT.AccessTokenExpires) * 3600,
		TokenType:            "Bearer",
		User: dto.UserInfo{
			ID:       int(user.ID),
			Username: user.Username,
		},
	}, nil
}
