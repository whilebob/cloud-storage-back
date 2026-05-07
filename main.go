package main

import (
	"CloudStorage/config"
	"CloudStorage/internal/router"
	"CloudStorage/internal/utils"

	"go.uber.org/zap"
)

func init() {
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	utils.InitAll()
}

func main() {
	r := router.InitRouter()
	utils.Logger.Info("CloudStorage 启动",
		zap.String("port", ":"+config.AppConfig.Server.Port))
	err := r.Run(":" + config.AppConfig.Server.Port)
	if err != nil {
		utils.Logger.Fatal("服务启动失败", zap.Error(err))
	}
}
