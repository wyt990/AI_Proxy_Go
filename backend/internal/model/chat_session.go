package model

import "time"

// ChatSession 聊天会话模型
type ChatSession struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UserID       int64     `gorm:"not null;index;comment:所属用户ID" json:"userId"`
	Title        string    `gorm:"size:200;comment:会话标题" json:"title"`
	ProviderID   int64     `gorm:"not null;index;comment:服务商ID" json:"providerId"`
	ModelID      int64     `gorm:"not null;index;comment:模型ID" json:"modelId"`
	KeyID        int64     `gorm:"not null;comment:使用的密钥ID" json:"keyId"`
	Status       string    `gorm:"size:20;not null;default:'active';comment:状态(active/archived)" json:"status"`
	LastMessage  string    `gorm:"type:text;comment:最后一条消息" json:"lastMessage"`
	MessageCount int       `gorm:"not null;default:0;comment:消息数量" json:"messageCount"`
	CreatedAt    time.Time `gorm:"type:datetime(3)" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"type:datetime(3)" json:"updatedAt"`
}

// TableName 指定表名
func (ChatSession) TableName() string {
	return "chat_sessions"
} 