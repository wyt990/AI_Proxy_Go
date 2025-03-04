package model

import "time"

// AI_Provider AI服务商模型
type AI_Provider struct {
	ID      uint   `gorm:"primarykey"`
	Name    string `gorm:"size:100;not null;comment:服务商名称"`
	Type    string `gorm:"size:50;not null;comment:服务商类型(OPENAI/ANTHROPIC/GOOGLE/BAIDU/CUSTOM)"`
	BaseURL string `gorm:"size:255;not null;column:base_url;comment:API基础URL"`

	// 请求格式配置
	RequestFormat  string `gorm:"type:text;not null;column:request_format;comment:请求数据格式(JSON Schema)"`
	ResponseFormat string `gorm:"type:text;not null;column:response_format;comment:响应数据格式(JSON Schema)"`
	AuthFormat     string `gorm:"type:text;not null;column:auth_format;comment:认证格式(JSON Schema)"`

	// 请求头和代理配置
	Headers   string `gorm:"type:text;column:headers;comment:请求头模板(JSON)"`
	AuthType  string `gorm:"size:50;not null;column:auth_type;comment:认证类型"`
	ProxyMode string `gorm:"size:50;not null;column:proxy_mode;default:'AGENT';comment:代理模式(AGENT/FORWARD)"`

	// 请求控制
	RateLimit     int `gorm:"not null;default:60;column:rate_limit;comment:每分钟请求限制"`
	Timeout       int `gorm:"not null;default:30;column:timeout;comment:请求超时时间(秒)"`
	RetryCount    int `gorm:"not null;default:3;column:retry_count;comment:重试次数"`
	RetryInterval int `gorm:"not null;default:1;column:retry_interval;comment:重试间隔(秒)"`

	// 状态信息
	Status        string    `gorm:"size:50;not null;default:'NORMAL';comment:状态(NORMAL/ERROR)"`
	LastError     string    `gorm:"type:text;column:last_error;comment:最后一次错误信息"`
	LastCheckTime time.Time `gorm:"column:last_check_time;comment:最后一次健康检查时间"`

	// 时间戳
	CreatedAt time.Time  `gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time  `gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
	DeletedAt *time.Time `gorm:"type:datetime(3);index;comment:删除时间"`

	// Token 使用统计
	PromptTokens     int `gorm:"default:0;comment:提示词使用的token数量" json:"promptTokens"`
	CompletionTokens int `gorm:"default:0;comment:回复使用的token数量" json:"completionTokens"`
	TotalTokens      int `gorm:"default:0;comment:总token数量" json:"totalTokens"`
}

// TableName 指定表名
func (AI_Provider) TableName() string {
	return "AI_providers"
}

// 示例 JSON Schema 格式
const (
	// OpenAI 请求格式示例
	OpenAIRequestSchema = `{
        "type": "object",
        "required": ["model", "messages"],
        "properties": {
            "model": {
                "type": "string",
                "title": "模型"
            },
            "messages": {
                "type": "array",
                "items": {
                    "type": "object",
                    "required": ["role", "content"],
                    "properties": {
                        "role": {
                            "type": "string",
                            "enum": ["system", "user", "assistant"]
                        },
                        "content": {
                            "type": "string"
                        }
                    }
                }
            },
            "temperature": {
                "type": "number",
                "minimum": 0,
                "maximum": 2,
                "default": 1
            }
        }
    }`

	// OpenAI 认证格式示例
	OpenAIAuthSchema = `{
        "type": "object",
        "required": ["api_key"],
        "properties": {
            "api_key": {
                "type": "string",
                "title": "API Key"
            },
            "organization": {
                "type": "string",
                "title": "Organization ID"
            }
        }
    }`
)
