package model

import (
	"time"
)

// SystemSettings 系统配置表
type SystemSettings struct {
	ID          uint       `gorm:"primarykey"`
	ConfigKey   string     `gorm:"column:config_key;type:varchar(50);uniqueIndex;not null"`
	Value       string     `gorm:"type:text"`
	Description string     `gorm:"type:text;comment:配置项说明"`
	CreatedAt   time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP()"`
	UpdatedAt   time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP()"`
	DeletedAt   *time.Time `gorm:"type:datetime;null"`
}

// TableName 指定表名
func (SystemSettings) TableName() string {
	return "system_settings"
}

// 预定义的配置键
const (
	// 系统状态相关
	KeyInstalled        = "system.installed"
	KeyInstallTime      = "system.install_time"
	KeyAdminInitialized = "system.admin_initialized"

	// Redis配置相关
	KeyRedisHost     = "redis.host"
	KeyRedisPort     = "redis.port"
	KeyRedisPassword = "redis.password"
	KeyRedisDB       = "redis.db"

	// 搜索设置相关
	KeySearchEngineURL   = "search.engine_url"
	KeySearchResultCount = "search.result_count"
	KeySearchMaxKeywords = "search.max_keywords"
	KeySearchConcurrency = "search.concurrency"

	// 新增配置
	KeySearchTimeout       = "search.timeout"
	KeySearchRetryCount    = "search.retry_count"
	KeySearchCacheDuration = "search.cache_duration"

	// 权重配置
	KeySearchWeightOfficial = "search.weight.official"
	KeySearchWeightEdu      = "search.weight.edu"
	KeySearchWeightNews     = "search.weight.news"
	KeySearchWeightPortal   = "search.weight.portal"

	// 时效性权重
	KeySearchTimeWeightDay   = "search.time_weight.day"
	KeySearchTimeWeightWeek  = "search.time_weight.week"
	KeySearchTimeWeightMonth = "search.time_weight.month"
	KeySearchTimeWeightYear  = "search.time_weight.year"

	// 过滤配置
	KeySearchMinContentLength = "search.min_content_length"
	KeySearchMaxSummaryLength = "search.max_summary_length"
	KeySearchFilterDomains    = "search.filter_domains"

	// 搜索质量控制
	KeySearchMinTitleLength    = "search.min_title_length"
	KeySearchMaxResults        = "search.max_results"
	KeySearchMinRelevanceScore = "search.min_relevance_score"

	// 缓存控制
	KeySearchCacheCleanupInterval = "search.cache_cleanup_interval"
	KeySearchMaxCacheSize         = "search.max_cache_size"

	// 请求控制
	KeySearchRateLimit         = "search.rate_limit"
	KeySearchMaxRetryInterval  = "search.max_retry_interval"
	KeySearchConnectionTimeout = "search.connection_timeout"

	// 内容过滤
	KeySearchBlockedKeywords  = "search.blocked_keywords"
	KeySearchRequiredKeywords = "search.required_keywords"
	KeySearchLanguageFilter   = "search.language_filter"

	// 会话控制

	KeySessionContextLength = "session.context_length" // 上下文最大条数

	// AI请求控制
	KeyAIRequestTimeout = "ai.request_timeout" // AI请求超时时间

	// tokens统计
	PromptTokens     = "prompt_tokens"
	CompletionTokens = "completion_tokens"
	TotalTokens      = "total_tokens"
)

// 初始化配置说明
var DefaultSettingsDescription = map[string]string{
	KeyInstalled:        "系统是否已完成安装",
	KeyInstallTime:      "系统安装完成时间",
	KeyAdminInitialized: "是否已初始化管理员账户",

	KeyRedisHost:     "Redis服务器地址",
	KeyRedisPort:     "Redis端口",
	KeyRedisPassword: "Redis密码",
	KeyRedisDB:       "Redis数据库索引",

	// 搜索设置
	KeySearchEngineURL:   "搜索引擎地址",
	KeySearchResultCount: "参考记录数",
	KeySearchMaxKeywords: "最大关键词字数",
	KeySearchConcurrency: "并发搜索进程数",

	// 搜索性能设置
	KeySearchTimeout:       "搜索超时时间(秒)",
	KeySearchRetryCount:    "失败重试次数",
	KeySearchCacheDuration: "缓存时间(秒)",

	// 搜索权重设置
	KeySearchWeightOfficial: "官方网站权重",
	KeySearchWeightEdu:      "教育机构权重",
	KeySearchWeightNews:     "新闻网站权重",
	KeySearchWeightPortal:   "门户网站权重",

	// 时效性权重设置
	KeySearchTimeWeightDay:   "24小时内内容权重",
	KeySearchTimeWeightWeek:  "一周内内容权重",
	KeySearchTimeWeightMonth: "一月内内容权重",
	KeySearchTimeWeightYear:  "一年内内容权重",

	// 过滤设置
	KeySearchMinContentLength: "最小内容长度",
	KeySearchMaxSummaryLength: "最大摘要长度",
	KeySearchFilterDomains:    "需要过滤的域名列表",

	// 搜索质量控制说明
	KeySearchMinTitleLength:    "最小标题长度",
	KeySearchMaxResults:        "最大返回结果数",
	KeySearchMinRelevanceScore: "最低相关度分数(0-1)",

	// 缓存控制说明
	KeySearchCacheCleanupInterval: "缓存清理间隔(秒)",
	KeySearchMaxCacheSize:         "最大缓存条目数",

	// 请求控制说明
	KeySearchRateLimit:         "每分钟最大请求次数",
	KeySearchMaxRetryInterval:  "最大重试间隔(秒)",
	KeySearchConnectionTimeout: "连接超时时间(秒)",

	// 内容过滤说明
	KeySearchBlockedKeywords:  "屏蔽关键词列表(逗号分隔)",
	KeySearchRequiredKeywords: "必需关键词列表(逗号分隔)",
	KeySearchLanguageFilter:   "语言过滤设置(逗号分隔)",

	// 会话控制说明

	KeySessionContextLength: "对话上下文最大保留条数，影响AI回复的上下文理解",

	// AI请求控制说明
	KeyAIRequestTimeout: "AI请求超时时间(秒)，超过此时间将中断请求",

	PromptTokens:     "提示词已使用的token总数",
	CompletionTokens: "回复已使用的token总数",
	TotalTokens:      "已使用的token总数",
}

// DefaultSettingsValues 定义所有设置的默认值
var DefaultSettingsValues = map[string]string{
	// 基础搜索设置
	KeySearchEngineURL:   "http://127.0.0.1/search?format=json&q=",
	KeySearchResultCount: "5",
	KeySearchMaxKeywords: "50",
	KeySearchConcurrency: "3",

	// 性能设置默认值
	KeySearchTimeout:       "30",
	KeySearchRetryCount:    "3",
	KeySearchCacheDuration: "3600",

	// 权重设置默认值
	KeySearchWeightOfficial: "5",
	KeySearchWeightEdu:      "5",
	KeySearchWeightNews:     "4",
	KeySearchWeightPortal:   "3",

	// 时效性权重默认值
	KeySearchTimeWeightDay:   "2",
	KeySearchTimeWeightWeek:  "1",
	KeySearchTimeWeightMonth: "0",
	KeySearchTimeWeightYear:  "-1",

	// 过滤设置默认值
	KeySearchMinContentLength: "50",
	KeySearchMaxSummaryLength: "200",
	KeySearchFilterDomains:    "",

	// 搜索质量控制默认值
	KeySearchMinTitleLength:    "10",
	KeySearchMaxResults:        "50",
	KeySearchMinRelevanceScore: "0.6",

	// 缓存控制默认值
	KeySearchCacheCleanupInterval: "3600",
	KeySearchMaxCacheSize:         "10000",

	// 请求控制默认值
	KeySearchRateLimit:         "60",
	KeySearchMaxRetryInterval:  "30",
	KeySearchConnectionTimeout: "10",

	// 内容过滤默认值
	KeySearchBlockedKeywords:  "广告,推广,软文",
	KeySearchRequiredKeywords: "",
	KeySearchLanguageFilter:   "zh,en",

	// 会话控制默认值
	KeySessionContextLength: "3", // 默认保留3条上下文

	// AI请求控制默认值
	KeyAIRequestTimeout: "120", // 默认120秒
}
