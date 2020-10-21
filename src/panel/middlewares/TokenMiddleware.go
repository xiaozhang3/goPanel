package middlewares

import (
	"github.com/gin-gonic/gin"
	"goPanel/src/panel/common"
	core "goPanel/src/panel/core/database"
	"goPanel/src/panel/services"
)

type TokenMiddleware struct {
	userService *services.UserService
}

func (m *TokenMiddleware) Middleware() gin.HandlerFunc {
	return func(g *gin.Context) {
		m.userService = new(services.UserService)
		token := g.Request.Header.Get("Account-Token")

		state, msg, code := m.userService.IsUserLogin(token)
		if !state {
			common.RetJson(g, code, msg, "")
			return
		}

		userData := m.userService.TokenByData(core.Db, token)
		g.Set("userinfo", &userData)

		// 处理请求
		g.Next()
	}
}
