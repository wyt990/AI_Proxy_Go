package model

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:100;not null" json:"-"`
	Name         string    `gorm:"size:50;not null" json:"name"`
	Email        string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Role         string    `gorm:"size:20;not null;default:'user'" json:"role"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	LastLogin    time.Time `gorm:"type:datetime;null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Token 使用统计
	PromptTokens     int `gorm:"default:0;comment:提示词使用的token数量" json:"promptTokens"`
	CompletionTokens int `gorm:"default:0;comment:回复使用的token数量" json:"completionTokens"`
	TotalTokens      int `gorm:"default:0;comment:总token数量" json:"totalTokens"`
}
