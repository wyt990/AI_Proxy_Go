package model

import "time"

// AI_model AI模型配置
type AI_Model struct {
	ID          uint        `gorm:"primarykey"`
	Name        string      `gorm:"size:100;not null;comment:模型名称" binding:"required"`
	ProviderID  uint        `gorm:"not null;comment:所属服务商ID" binding:"required"`
	Provider    AI_Provider `gorm:"foreignKey:ProviderID"`
	ModelID     string      `gorm:"size:100;not null;comment:模型ID/标识符" binding:"required"`
	Type        string      `gorm:"size:20;not null;default:'chat';comment:模型类型(chat/text2image/image2image/audio/video/embedding/tool)"`
	Parameters  string      `gorm:"type:text;comment:模型参数(JSON)"`
	Status      string      `gorm:"size:50;not null;default:'NORMAL';comment:状态(NORMAL/DISABLED)"`
	Description string      `gorm:"type:text;comment:模型描述"`
	Sort        int         `gorm:"default:0;comment:排序(数字越大越靠前)"`
	MaxTokens   int         `gorm:"default:0;comment:最大Token数(0表示不限制)"`
	InputPrice  string      `gorm:"type:decimal(10,6);default:0;comment:输入价格(每1M token)"`
	OutputPrice string      `gorm:"type:decimal(10,6);default:0;comment:输出价格(每1M token)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`

	// Token 使用统计
	PromptTokens     int `gorm:"default:0;comment:提示词使用的token数量" json:"promptTokens"`
	CompletionTokens int `gorm:"default:0;comment:回复使用的token数量" json:"completionTokens"`
	TotalTokens      int `gorm:"default:0;comment:总token数量" json:"totalTokens"`
}

// TableName 指定表名
func (AI_Model) TableName() string {
	return "AI_models"
}
