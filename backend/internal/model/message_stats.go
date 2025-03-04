package model

import (
	"time"
)

// MessageStats 消息统计模型
type MessageStats struct {
	ID               int64     `gorm:"primarykey"`
	UserID           int64     `gorm:"index;comment:用户ID"`
	ProviderID       int64     `gorm:"index;comment:服务商ID"`
	ModelID          int64     `gorm:"index;comment:模型ID"`
	SessionID        uint      `gorm:"index;comment:会话ID"`
	PromptTokens     int64     `gorm:"type:bigint;comment:提示词token数"`
	CompletionTokens int64     `gorm:"type:bigint;comment:补全token数"`
	TotalTokens      int64     `gorm:"type:bigint;comment:总token数"`
	ResponseTime     int64     `gorm:"type:bigint;comment:响应时间(毫秒)"`
	Status           string    `gorm:"type:varchar(20);default:'success';comment:请求状态"`
	CreatedAt        time.Time `gorm:"type:datetime(3);index;comment:创建时间"`
}

// TableName 指定表名
func (MessageStats) TableName() string {
	return "message_stats"
}
