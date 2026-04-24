package service

import (
	"cloud-storage/common"
	"cloud-storage/global"
	"cloud-storage/model"
	"cloud-storage/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserService struct {
}

var _ User = (*UserService)(nil)

func (u *UserService) Register(c *gin.Context) {
	var req model.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "请求体错误"))
		return
	}

	// 1. 基础参数校验
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "用户名/密码不能为空"))
		return
	}

	key := common.UserRegisterKey

	// 2. Redis 检查用户名是否存在
	exists, err := common.RS.Contains(context.Background(), key, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "服务异常"))
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "用户名已存在"))
		return
	}

	// 3. 原子加入Set
	add, err := common.RS.Add(context.Background(), key, req.Username)
	if err != nil || add == 0 {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "用户名已存在"))
		return
	}

	// 4. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		_, _ = common.RS.Remove(context.Background(), key, req.Username)
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "密码加密失败"))
		return
	}

	user := model.User{
		Username: req.Username,
		Password: hashedPassword,
	}

	tx := global.DB.Begin() // 这里必须接收事务对象
	if tx.Error != nil {
		_, _ = common.RS.Remove(context.Background(), key, req.Username)
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "事务开启失败"))
		return
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback() // 事务回滚
		_, _ = common.RS.Remove(context.Background(), key, req.Username)
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "注册失败"))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		_, _ = common.RS.Remove(context.Background(), key, req.Username)
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "事务提交失败"))
		return
	}

	// 6. 注册成功
	c.JSON(http.StatusOK, common.Success(http.StatusOK, "注册成功", nil))
}

func (u *UserService) Login(c *gin.Context) {
	// 1. 参数接收
	var req model.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "参数格式错误"))
		return
	}

	username := req.Username
	password := req.Password

	// 2. 非空校验
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "用户名或密码不能为空"))
		return
	}

	// 3. 根据用户名查询用户
	var user model.User
	err := global.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		// 不提示“用户不存在”，提高安全性
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "用户名或密码错误"))
		return
	}

	// 4. 校验密码（关键！和注册时的加密对应）
	if !utils.CheckPassword(password, user.Password) {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "用户名或密码错误"))
		return
	}

	// 5. 生成 JWT Token（和注册返回的一样）
	token, err := utils.GenToken(username)
	if err != nil {
		err := common.RS.Set(context.Background(), common.UserLoginTokenKey+username, token, common.UserLoginTokenExpireTime)
		if err != nil {
			return
		}
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "登录凭证生成失败"))
		return
	}

	// 6. 登录成功，返回token
	c.JSON(http.StatusOK, common.SuccessWithData(gin.H{"token": token}))
}

func (u *UserService) LogOut(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)

	deleteKey := fmt.Sprintf("%s:%s", common.UserLoginTokenKey, username)
	err = global.Redis.Del(context.Background(), deleteKey).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "服务异常"))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithData("退出成功"))
	c.Abort()
	return
}
