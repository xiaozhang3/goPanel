package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"goPanel/src/common"
	"goPanel/src/constants"
	core "goPanel/src/core/database"
	"goPanel/src/gps/coer/socket"
	"goPanel/src/gps/models"
	"goPanel/src/gps/services"
	"goPanel/src/gps/validations"
	"io/ioutil"
	"time"
)

type CommandController struct {
	BaseController
	commandService *services.CommandService
}

func NewCommandController() *CommandController {
	return &CommandController{
		commandService: new(services.CommandService),
	}
}

func (c *CommandController) Add(g *gin.Context) {
	inputData, _ := ioutil.ReadAll(g.Request.Body)
	var addVail validations.CommandAdd
	c.JsonPost(&addVail, inputData)

	if err := c.Validations(addVail); err != nil {
		common.RetJson(g, constants.MISSING_PARAMETER_FAIL, err.Error(), "")
		return
	}

	flag, _ := common.StringUtils(addVail.Flag).Int()
	userinfo := c.GetUserInfo(g)

	for _, item := range addVail.Ids {
		// 构建数据
		tmpAddCommandData := models.CommandModel{}
		tmpAddCommandData.Command = addVail.Command
		tmpAddCommandData.Flag = flag
		tmpAddCommandData.CreateUid = userinfo.Id
		tmpAddCommandData.PlanExecTime = time.Now()
		tmpAddCommandData.MachineId = item
		tmpAddCommandData.CreateTime = time.Now()

		if flag == 2 {
			tmpAddCommandData.PlanExecTime, _ = time.ParseInLocation(constants.TIME_TEMPLATE, addVail.PlanExecTime, time.Local)
		}

		_, err := c.commandService.Add(core.Db, &tmpAddCommandData)
		if err != nil {
			common.RetJson(g, constants.ERROR_FAIL, constants.ERROR_FAIL_MSG, "")
			return
		}

		// 发送要跑的命令
		cliConn := socket.ControlManager.FindClientIdByClientConn(tmpAddCommandData.MachineId)
		if cliConn == nil {
			common.RetJson(g, constants.CLIENT_NOT_FOND_FAIL, constants.CLIENT_NOT_FOND_MSG, "")
			return
		}

		msg, _ := json.Marshal(socket.Message{
			Type:  0,
			Event: "handleCommand",
			Data:  tmpAddCommandData,
			Code:  constants.SUCCESS,
		})

		cliConn.Write <- msg
	}

	common.RetJson(g, constants.SUCCESS, constants.SUCCESS_MSG, "")
	return
}
