package socket

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"goPanel/src/common"
	"goPanel/src/gpc/config"
	"goPanel/src/gpc/router"
	"goPanel/src/gpc/service"
	"io"
	"io/ioutil"
	"net"
	"time"
)

var isReconnControlTcp = true
var Ctx, Cancel = context.WithCancel(context.Background())

func StartClientTcp() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	for {
		if err := Ctx.Err(); err != nil {
			Ctx, Cancel = context.WithCancel(context.Background())
		}

		if isReconnControlTcp {
			handelConnControlTcp(Ctx)
			log.Info("重连重试中！")
		}

		time.Sleep(time.Second * time.Duration(config.Conf.App.ControlReconnTcpTime))
	}
}

func closeClientTcp(ctx context.Context, conn *net.TCPConn) {
	for true {
		select {
		case <-ctx.Done():
			conn.Close()
			return
		}
	}
}

// 心跳
func heartbeat(ctx context.Context, conn *net.TCPConn) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second * time.Duration(config.Conf.App.ControlHeartbeatTime))

			write := service.RequestWsMessage{
				Event: "heartbeat",
				Data:  nil,
			}
			writeJson, err := json.Marshal(write)
			if err != nil {
				continue
			}

			log.Info("正在执行控制端心跳包")
			if _, err = conn.Write(writeJson); err != nil {
				log.Info(err)
				return
			}
		}
	}
}

// 注册本机数据
func registerLocalData(conn *net.TCPConn) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			conn.Close()
		}
	}()

	// 获取本机id数据
	uidFilePath := config.Conf.App.UidPath + "uid"
	var uid []byte
	var err error

	if common.DirOrFileByIsExists(uidFilePath) {
		uid, err = ioutil.ReadFile(uidFilePath)
		if err != nil {
			log.Error("客户机，uid文件读取错误：", err)
		}
	}

	if len(uid) == 0 {
		if !common.DirOrFileByIsExists(config.Conf.App.UidPath) {
			if !common.CreatePath(config.Conf.App.UidPath) {
				log.Error("uid目录创建失败!")
			}
		}

		id, _ := common.GenToken()
		uid = []byte(id)

		err = ioutil.WriteFile(uidFilePath, uid, 0755)
		if err != nil {
			log.Error("uid写文件出错！", err)
		}
	}

	config.Conf.App.Uid = string(uid)
	localComputerData := map[string]string{
		"name": config.Conf.App.LocalName,
		"uid":  string(uid),
	}
	write := service.RequestWsMessage{
		Event: "local_register",
		Data:  localComputerData,
	}
	writeJson, err := json.Marshal(write)
	if err != nil {
		log.Error(err)
		return
	}
	if _, err = conn.Write(writeJson); err != nil {
		log.Error(err)
		return
	}
}

func handelConnControlTcp(ctx context.Context) {
	defer func() {
		isReconnControlTcp = true
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	isReconnControlTcp = false
	conn, err := common.ConnTcp(service.ControlAddr)
	if err != nil {
		log.Error(err)
		return
	}

	defer func() {
		conn.Close()
	}()

	go closeClientTcp(ctx, conn)
	registerLocalData(conn)
	go heartbeat(ctx, conn)
	readControlTcpMess(ctx, conn)
}

func readControlTcpMess(ctx context.Context, conn *net.TCPConn) {
	for {
		select {
		case <-ctx.Done():
			isReconnControlTcp = true

			return
		default:
			var data = make([]byte, 10240)
			size, err := conn.Read(data)
			if err != nil || err == io.EOF {
				log.Error(err)
				break
			}
			data = data[:size]

			var message service.Message
			err = json.Unmarshal(data, &message)
			if err != nil {
				log.Info(err)
				continue
			}

			if _, ok := router.Route[message.Event]; ok {
				router.Route[message.Event](ctx, conn, message)
				continue
			}

			log.Error("请求路由不存在！")
		}
	}
}
