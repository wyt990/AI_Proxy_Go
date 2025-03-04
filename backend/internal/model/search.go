package model

import "time"

// SearchSource 搜索结果源
type SearchSource struct {
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	PublishTime    time.Time `json:"publish_time"`
	Domain         string    `json:"domain"`
	Type           string    `json:"type"` // 官方、教育、新闻等
	Weight         int       `json:"weight"`
	RelevanceScore float64   `json:"relevance_score"` // 相关度分数 0-1
	Language       string    `json:"language"`        // 语言代码，如 zh、en
}

// SearchReference 引用信息
type SearchReference struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Content     string    `json:"content"`
	PublishTime time.Time `json:"publish_time"`
}

// SearchResult 定义搜索结果的结构
type SearchResult struct {
	NumberOfResults     int      `json:"number_of_results"`
	Query               string   `json:"query"`
	Results             []Result `json:"results"`
	Suggestions         []string `json:"suggestions"`
	UnresponsiveEngines []string `json:"unresponsive_engines"`
}

// Result 定义单个搜索结果的结构
type Result struct {
	ID        int      `json:"id"`
	Category  string   `json:"category"`
	Content   string   `json:"content"`
	Engine    string   `json:"engine"`
	Engines   []string `json:"engines"`
	ParsedURL []string `json:"parsed_url"`
	Positions []int    `json:"positions"`
	Score     float64  `json:"score"`
	Template  string   `json:"template"`
	Thumbnail *string  `json:"thumbnail"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`

	// 添加缺失的字段
	Type           string    `json:"type"`            // 来源类型
	Domain         string    `json:"domain"`          // 域名
	Language       string    `json:"language"`        // 语言
	PublishTime    time.Time `json:"publish_time"`    // 发布时间
	RelevanceScore float64   `json:"relevance_score"` // 相关度分数
}
