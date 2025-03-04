package model

import "time"

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	SessionID  uint      `gorm:"not null;index;comment:会话ID" json:"sessionId"`
	UserID     int64     `gorm:"not null;index;comment:用户ID" json:"userId"`
	ProviderID int64     `gorm:"not null;comment:服务商ID"`
	ModelID    int64     `gorm:"not null;comment:模型ID"`
	KeyID      int64     `gorm:"not null;comment:密钥ID"`
	Role       string    `gorm:"size:20;not null;comment:角色(user/assistant)" json:"role"`
	Content    string    `gorm:"type:text;not null;comment:消息内容" json:"content"`
	CreatedAt  time.Time `gorm:"type:datetime(3)" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"type:datetime(3)" json:"updatedAt"`

	// Token 使用统计
	PromptTokens     int64    `gorm:"type:bigint;default:0;comment:提示词使用的token数量" json:"promptTokens"`
	CompletionTokens int64    `gorm:"type:bigint;default:0;comment:回复使用的token数量" json:"completionTokens"`
	TotalTokens      int64    `gorm:"type:bigint;default:0;comment:总token数量" json:"totalTokens"`

	// 关联查询字段（不存储在数据库中）
	ProviderName string `gorm:"-"`
	ModelName    string `gorm:"-"`
	UserName     string `gorm:"-"`
}

// TableName 指定表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}
