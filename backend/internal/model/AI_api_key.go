package model

import "time"

// AI_APIKey API密钥配置
type AI_APIKey struct {
	ID          int64  `gorm:"primaryKey"`
	ProviderID  int64  `gorm:"index;not null;comment:所属服务商ID"`
	UserID      int64  `gorm:"index;comment:所属用户ID(公钥为0)"`
	Type        string `gorm:"size:20;not null;comment:密钥类型(PUBLIC/PRIVATE)"`
	Name        string `gorm:"size:100;not null;comment:密钥名称"`
	KeyValue    string `gorm:"column:key_value;size:255;not null;index;comment:加密存储的密钥"`
	CreatorID   int64  `gorm:"not null;comment:创建者ID"`
	CreatorName string `gorm:"size:100;not null;comment:创建者姓名"`

	// 使用限制
	RateLimit  int        `gorm:"default:0;comment:每分钟请求限制(0表示使用服务商限制)"`
	QuotaLimit int64      `gorm:"default:0;comment:配额限制(0表示无限制)"`
	QuotaUsed  int64      `gorm:"default:0;comment:已使用配额"`
	ExpireTime *time.Time `gorm:"comment:过期时间"`
	IsActive   bool       `gorm:"default:true;comment:是否启用"`

	// Token 使用统计
	PromptTokens     int64 `gorm:"default:0;comment:提示词已使用的token总数" json:"promptTokens"`
	CompletionTokens int64 `gorm:"default:0;comment:回复已使用的token总数" json:"completionTokens"`
	TotalTokens      int64 `gorm:"default:0;comment:已使用的token总数" json:"totalTokens"`

	// 时间戳
	CreatedAt time.Time `gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time `gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
	DeletedAt time.Time `gorm:"type:datetime(3)"`

	// 关联查询字段（不存储在数据库中）
	ProviderName string `gorm:"-"` // 服务商名称
}

// TableName 指定表名
func (AI_APIKey) TableName() string {
	return "AI_api_keys"
}
