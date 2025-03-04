package middleware

import (
	"AI_Proxy_Go/backend/internal/model"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	SecretKey = "your-secret-key" // 在生产环境中应该从配置文件读取
)

// Claims 定义JWT的声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过登录和安装相关的路由
		if c.Request.URL.Path == "/login" ||
			c.Request.URL.Path == "/install" ||
			c.Request.URL.Path == "/favicon.ico" ||
			strings.HasPrefix(c.Request.URL.Path, "/static/") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/login") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/install") {
			c.Next()
			return
		}

		// 检查是否已登录
		token := c.GetHeader("Authorization")
		if token == "" {
			// 从 cookie 中获取 token
			token, _ = c.Cookie("token")
		}

		if token == "" {
			// 对于 API 请求返回 JSON
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				log.Printf("未登录，获取用户信息失败")
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":  "UNAUTHORIZED",
					"error": "未登录",
				})
				c.Abort()
				return
			}

			// 对于页面请求重定向到登录页
			log.Printf("未登录，重定向到登录页")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// TODO: 验证 token 的有效性

		c.Next()
	}
}

// GenerateToken 生成JWT Token
func GenerateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token有效期24小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}
