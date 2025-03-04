package model

import "time"

// Captcha 验证码模型
type Captcha struct {
	ID        int64     `gorm:"primaryKey"`
	Code      string    `gorm:"size:10;not null;comment:验证码内容"`
	ImageData string    `gorm:"type:text;not null;comment:验证码图片数据(Base64)"`
	IP        string    `gorm:"size:45;not null;comment:请求IP"`
	UserAgent string    `gorm:"size:255;comment:用户代理"`
	SessionID string    `gorm:"size:64;comment:会话ID"`
	Type      string    `gorm:"size:20;default:LOGIN;comment:验证码类型(LOGIN/REGISTER/RESET)"`
	ExpiredAt time.Time `gorm:"not null;comment:过期时间"`
	Used      bool      `gorm:"default:false;comment:是否已使用"`
	UsedAt    time.Time `gorm:"comment:使用时间"`
	FailCount int       `gorm:"default:0;comment:验证失败次数"`
	CreatedAt time.Time `gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time `gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

// TableName 指定表名
func (Captcha) TableName() string {
	return "captcha"
}
