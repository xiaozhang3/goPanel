package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"goPanel/src/panel/common"
	"goPanel/src/panel/library/ssh"
	"log"
	"net/http"
	"time"
)

func Ssh(c *gin.Context) {
	ws, err := (&websocket.Upgrader{
		HandshakeTimeout: time.Duration(time.Second * 30),
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer func() {
		ws.Close()
	}()

	cols, _ := common.StringUtils(c.Param("cols")).Uint32()
	rows, _ := common.StringUtils(c.Param("rows")).Uint32()
	host := c.Param("host")

	// 通过ip获取相关ssh客户端数据
	sh := ssh.NewSsh(host, "yeyu", "ZpB123", 22)
	sshChannel, err := sh.RunShell(ssh.TermConfig{
		Cols: cols,
		Rows: rows,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := sshChannel.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	wsRead := make(chan []byte, 5120)
	wsWrite := make(chan []byte, 10240)
	var ch chan bool

	go sh.Read(sshChannel, wsWrite)
	go sh.Write(sshChannel, wsRead)

	// 读ws客户端数据
	go func() {
		defer func() {
			ch <- true
		}()

		for {
			mt, message, err := ws.ReadMessage()
			// 其他错误，如果是 1001 和 1000 就不打印日志
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("ReadMessage other remote:%v error: %v \n", ws.RemoteAddr(), err)
				return
			}

			if mt == websocket.TextMessage {
				wsRead <- message
			}
		}
	}()

	// 写ws客户端数据
	go func() {
		defer func() {
			ch <- true
		}()

		for {
			select {
			case message := <-wsWrite:
				err = ws.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Fatal(err)
					return
				}
			}
		}
	}()

	<-ch
}
