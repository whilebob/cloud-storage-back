package handler

import (
	"CloudStorage/internal/common"
	"CloudStorage/internal/dto"
	"CloudStorage/internal/service"
	"CloudStorage/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct{}

var userService *service.UserService

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) Register(c *gin.Context) {
	var user dto.UserRegisterDTO
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Logger.Error("注册请求参数错误", zap.Error(err))
		common.BadRequest(c)
		return
	}
	register, err := userService.Register(&user)
	if err != nil {
		utils.Logger.Error("注册失败", zap.Error(err))
		common.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	common.Success(c, register)
}

func (h *UserHandler) Login(c *gin.Context) {
	var user dto.UserLoginDTO
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Logger.Error("登录请求参数错误", zap.Error(err))
		common.BadRequest(c)
		return
	}
	if user.Username == "" || user.Password == "" {
		common.BadRequest(c)
		return
	}
	token, err := userService.Login(&user)
	if err != nil {
		utils.Logger.Error("登录失败", zap.Error(err))
		common.Error(c)
		return
	}
	common.Success(c, token)
}

func (h *UserHandler) Logout(c *gin.Context) {

}
