package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func Ok(c *gin.Context) {
	Success(c, nil)
}
func Unauthorized(c *gin.Context) {
	Fail(c, 401, "未登录或 token 已过期")
}
func Forbidden(c *gin.Context) {
	Fail(c, 403, "无权限访问")
}
func BadRequest(c *gin.Context) {
	Fail(c, 400, "请求参数错误")
}
func TooManyRequests(c *gin.Context) {
	Fail(c, 429, "请求过于频繁，请稍后再试")
}
func Error(c *gin.Context) {
	Fail(c, 500, "服务器异常")
}
