package api

import (
	"AI_Proxy_Go/backend/internal/middleware"
	"AI_Proxy_Go/backend/internal/model"

	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB      *gorm.DB
	Captcha *CaptchaHandler // 改为大写，使其公开
}

// GetUserInfo 获取当前登录用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	// 从上下文中获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(400, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(500, gin.H{"error": "获取用户信息失败"})
		return
	}

	c.JSON(200, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		CaptchaId   string `json:"captchaId" binding:"required"`
		CaptchaCode string `json:"captchaCode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		//log.Printf("登录请求参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写所有必填字段"})
		return
	}

	//log.Printf("登录请求参数: username=%s, captchaId=%s, captchaCode=%s",
	//	req.Username, req.CaptchaId, req.CaptchaCode)

	// 验证验证码
	if !h.Captcha.VerifyCaptcha(req.CaptchaId, req.CaptchaCode) {
		//log.Printf("验证码验证失败: id=%s, code=%s", req.CaptchaId, req.CaptchaCode)
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误"})
		return
	}

	// 查找用户
	var user model.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 检查用户状态
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "账户已被禁用"})
		return
	}

	// 生成Token
	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	// 更新最后登录时间
	now := time.Now()
	if err := h.DB.Model(&user).UpdateColumn("last_login", now).Error; err != nil {
		// 记录错误日志，但不影响登录流程
		log.Printf("更新用户登录时间失败: %v", err)
	}

	// 设置Cookie
	c.SetCookie(
		"token",
		token,
		86400, // 24小时
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Logout 处理登出请求
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(
		"token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(200, gin.H{"success": true})
}
