package main

import (
	"github.com/gin-gonic/gin"
	core "goPanel/src/core/database"
	core_log "goPanel/src/core/log"
	"goPanel/src/gps/coer/socket"
	"goPanel/src/gps/config"
	"goPanel/src/gps/models"
	"goPanel/src/gps/router"
	"strconv"
	"time"
)

func main() {
	// 服务启动前，初始化操作
	_, _ = time.LoadLocation("Asia/Shanghai")
	core_log.Initialization(
		config.Conf.App.LogOutputType,
		config.Conf.App.Debug,
		config.Conf.App.LogLevel,
	)
	core_log.LogSetOutput(config.Conf.App.LogPath, config.Conf.App.LogOutputFlag)
	createTable()
	go socket.ServerWsManager.Start()
	go socket.ControlManager.Start()
	//common.GenRsaKey(common.GetRsaFilePath(), 2048)

	g := gin.Default()
	g = (new(router.Route)).Init(g)
	_ = g.Run(":" + strconv.Itoa(config.Conf.App.HttpPort))
}

// 创建表
func createTable() {
	core.CreateTables(
		new(models.UserModel),
		new(models.MachineModel),
		new(models.MachineGroupModel),
		new(models.CommandModel),
	)
}
