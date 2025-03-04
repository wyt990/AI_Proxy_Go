package middleware

import (
	"AI_Proxy_Go/backend/internal/install"
	"strings"

	"github.com/gin-gonic/gin"
)

func InstallCheck(installer *install.Installer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是安装页面或安装相关的API，直接放行
		if c.Request.URL.Path == "/install" ||
			strings.HasPrefix(c.Request.URL.Path, "/api/install/") {
			c.Next()
			return
		}

		// 检查是否已安装
		if !installer.IsInstalled() {
			// 如果是API请求，返回JSON
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(403, gin.H{
					"error": "System not installed",
					"code":  "NOT_INSTALLED",
				})
				c.Abort()
				return
			}
			// 非API请求重定向到安装页面
			c.Redirect(302, "/install")
			c.Abort()
			return
		}

		c.Next()
	}
}
