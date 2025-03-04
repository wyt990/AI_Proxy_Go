package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// CaptchaHandler 验证码处理器
type CaptchaHandler struct {
	store base64Captcha.Store
}

// NewCaptchaHandler 创建验证码处理器
func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{
		store: base64Captcha.DefaultMemStore,
	}
}

// GenerateCaptcha 生成验证码
func (h *CaptchaHandler) GenerateCaptcha(c *gin.Context) {
	// 配置验证码参数
	driver := base64Captcha.NewDriverDigit(
		40,  // 高度
		120, // 宽度
		4,   // 验证码长度改为4位，更容易输入
		0.7, // 干扰强度
		50,  // 前景色数量
	)

	// 生成验证码
	captcha := base64Captcha.NewCaptcha(driver, h.store)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		c.JSON(500, gin.H{"error": "生成验证码失败"})
		return
	}

	// 添加调试日志
	//log.Printf("生成新验证码: id=%s", id)

	c.JSON(200, gin.H{
		"captchaId":   id,
		"imageBase64": b64s,
	})
}

// VerifyCaptcha 验证验证码
func (h *CaptchaHandler) VerifyCaptcha(captchaId, code string) bool {
	//log.Printf("验证码验证: id=%s, code=%s", captchaId, code)
	if captchaId == "" || code == "" {
		return false
	}
	result := h.store.Verify(captchaId, code, true)
	//log.Printf("验证码验证结果: %v", result)
	return result
}
