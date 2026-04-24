package main

import (
	"cloud-storage/api"
	"cloud-storage/common"
	"cloud-storage/config"
	"cloud-storage/utils"
	"fmt"
)

func main() {
	utils.InitZapLogger()
	defer utils.Logger.Sync()
	utils.Logger.Info("日志初始化成功")

	utils.Logger.Info("✅ 数据库日志初始化完成！")
	r := api.RouterAll()
	r.Run(":3000")
	fmt.Println("Server is running on port 3000")
}

// 初始化配置
func init() {
	common.InitDBLogger()
	err := config.LoadConfig()
	common.InitAll()
	if err != nil {
		panic(err)
	}
}
