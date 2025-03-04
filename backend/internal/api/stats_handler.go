package api

import (
	m "AI_Proxy_Go/backend/internal/model"
	"math"
	"net/http"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StatsHandler 统计数据处理器
type StatsHandler struct {
	DB *gorm.DB
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{DB: db}
}

// 计算系统负载
func calculateSystemLoad(cpuUsage, memoryUsage float64) float64 {
	// CPU使用率权重为0.6，内存使用率权重为0.4
	// 这个权重可以根据实际情况调整
	const (
		cpuWeight    = 0.6
		memoryWeight = 0.4
	)

	// 确保使用率在0-100范围内
	cpuUsage = math.Min(100, math.Max(0, cpuUsage))
	memoryUsage = math.Min(100, math.Max(0, memoryUsage))

	// 计算加权平均值
	systemLoad := (cpuUsage * cpuWeight) + (memoryUsage * memoryWeight)

	// 如果任一指标超过90%，增加系统负载权重
	if cpuUsage > 90 || memoryUsage > 90 {
		systemLoad *= 1.2 // 增加20%的负载权重
	}

	// 确保最终结果在0-100范围内
	return math.Min(100, math.Max(0, systemLoad))
}

// GetDashboardStats 获取仪表盘统计数据
func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var stats struct {
		TodayRequests      int64   `json:"todayRequests"`
		RequestsGrowth     float64 `json:"requestsGrowth"`
		TodayDialogs       int64   `json:"todayDialogs"`
		DialogsGrowth      float64 `json:"dialogsGrowth"`
		AvgResponseTime    float64 `json:"avgResponseTime"`
		ResponseTimeChange float64 `json:"responseTimeChange"`
		SystemLoad         float64 `json:"systemLoad"`
	}

	// 今日和昨日请求数
	var todayReqs, yesterdayReqs int64
	h.DB.Model(&m.MessageStats{}).Where("DATE(created_at) = ?", today).Count(&todayReqs)
	h.DB.Model(&m.MessageStats{}).Where("DATE(created_at) = ?", yesterday).Count(&yesterdayReqs)

	stats.TodayRequests = todayReqs
	if yesterdayReqs > 0 {
		stats.RequestsGrowth = float64(todayReqs-yesterdayReqs) / float64(yesterdayReqs) * 100
	}

	// 今日和昨日对话数
	var todayDialogs, yesterdayDialogs int64
	h.DB.Model(&m.MessageStats{}).Where("DATE(created_at) = ?", today).Distinct("session_id").Count(&todayDialogs)
	h.DB.Model(&m.MessageStats{}).Where("DATE(created_at) = ?", yesterday).Distinct("session_id").Count(&yesterdayDialogs)

	stats.TodayDialogs = todayDialogs
	if yesterdayDialogs > 0 {
		stats.DialogsGrowth = float64(todayDialogs-yesterdayDialogs) / float64(yesterdayDialogs) * 100
	}

	// 今日和昨日平均响应时间
	var todayAvgTime, yesterdayAvgTime float64
	h.DB.Model(&m.MessageStats{}).Where("DATE(created_at) = ?", today).Select("AVG(response_time)").Row().Scan(&todayAvgTime)
	h.DB.Model(&m.MessageStats{}).Where("DATE(created_at) = ?", yesterday).Select("AVG(response_time)").Row().Scan(&yesterdayAvgTime)

	stats.AvgResponseTime = todayAvgTime
	if yesterdayAvgTime > 0 {
		stats.ResponseTimeChange = (todayAvgTime - yesterdayAvgTime) / yesterdayAvgTime * 100
	}

	// 获取最新的系统指标
	var systemMetrics m.SystemMetrics
	if err := h.DB.Order("created_at DESC").First(&systemMetrics).Error; err == nil {
		// 计算系统整体负载
		stats.SystemLoad = calculateSystemLoad(
			systemMetrics.CPUUsage,
			systemMetrics.MemoryUsage,
		)
	}

	c.JSON(http.StatusOK, stats)
}

// TokenStats Token使用统计数据
type TokenStats struct {
	Date             string `json:"date"`
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
	TotalTokens      int64  `json:"totalTokens"`
}

// ModelUsageStats 模型使用统计数据
type ModelUsageStats struct {
	ModelName string  `json:"modelName"`
	Usage     int64   `json:"usage"`
	Percent   float64 `json:"percent"`
}

// GetTokenStats 获取Token使用趋势
func (h *StatsHandler) GetTokenStats(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	period := c.DefaultQuery("period", "day") // day, week, month

	var stats []TokenStats
	query := h.DB.Model(&m.MessageStats{})

	switch period {
	case "week":
		// 最近7天的数据，按天统计
		query = query.Select("DATE(created_at) as date, " +
			"SUM(prompt_tokens) as prompt_tokens, " +
			"SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens")
		query = query.Where("created_at >= ?", time.Now().AddDate(0, 0, -7))
		query = query.Group("DATE(created_at)")
	case "month":
		// 最近30天的数据，按天统计
		query = query.Select("DATE(created_at) as date, " +
			"SUM(prompt_tokens) as prompt_tokens, " +
			"SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens")
		query = query.Where("created_at >= ?", time.Now().AddDate(0, 0, -30))
		query = query.Group("DATE(created_at)")
	default:
		// 今天的数据，按小时统计
		query = query.Select("DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') as date, " +
			"SUM(prompt_tokens) as prompt_tokens, " +
			"SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens")
		query = query.Where("DATE(created_at) = ?", time.Now().Format("2006-01-02"))
		query = query.Group("DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')")
	}

	query.Order("date").Find(&stats)
	c.JSON(http.StatusOK, stats)
}

// GetModelUsage 获取模型使用分布
func (h *StatsHandler) GetModelUsage(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	var stats []ModelUsageStats
	var total int64

	// 获取总使用量
	h.DB.Model(&m.MessageStats{}).Count(&total)

	// 按模型统计使用量，注意 usage 是保留字，需要用反引号
	h.DB.Model(&m.MessageStats{}).
		Select("m.name as model_name, COUNT(*) as `usage`").
		Joins("JOIN ai_models m ON message_stats.model_id = m.id").
		Group("m.name").
		Scan(&stats)

	// 计算百分比
	if total > 0 {
		for i := range stats {
			stats[i].Percent = float64(stats[i].Usage) / float64(total) * 100
		}
	}

	// 如果没有数据，返回空数组而不是 null
	if stats == nil {
		stats = make([]ModelUsageStats, 0)
	}

	c.JSON(http.StatusOK, stats)
}

// RequestStats 请求统计数据
type RequestStats struct {
	Time    string `json:"time"`
	Count   int64  `json:"count"`
	Success int64  `json:"success"`
	Failed  int64  `json:"failed"`
	AvgTime int64  `json:"avgTime"`
}

// ProviderStats 服务商统计数据
type ProviderStats struct {
	ProviderName string  `json:"providerName"`
	ModelName    string  `json:"modelName"`
	Usage        int64   `json:"usage"`
	SuccessRate  float64 `json:"successRate"`
}

// GetRequestMonitor 获取实时请求监控数据
func (h *StatsHandler) GetRequestMonitor(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	var stats []RequestStats

	// 获取最近30分钟的数据，按分钟统计
	h.DB.Model(&m.MessageStats{}).
		Select("DATE_FORMAT(created_at, '%H:%i') as time, "+
			"COUNT(*) as count, "+
			"SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success, "+
			"SUM(CASE WHEN status != 'success' THEN 1 ELSE 0 END) as failed, "+
			"CAST(AVG(response_time) AS SIGNED) as avg_time"). // 使用 CAST 转换为整数
		Where("created_at >= ?", time.Now().Add(-30*time.Minute)).
		Group("time").
		Order("time").
		Scan(&stats)

	// 如果没有数据，返回空数组而不是 null
	if stats == nil {
		stats = make([]RequestStats, 0)
	}

	c.JSON(http.StatusOK, stats)
}

// GetProviderStats 获取服务商统计数据
func (h *StatsHandler) GetProviderStats(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	period := c.DefaultQuery("period", "day")
	var stats []ProviderStats

	query := h.DB.Model(&m.MessageStats{}).
		Joins("JOIN ai_models m ON message_stats.model_id = m.id").
		Joins("JOIN ai_providers p ON m.provider_id = p.id")

	switch period {
	case "week":
		query = query.Where("message_stats.created_at >= ?", time.Now().AddDate(0, 0, -7))
	case "month":
		query = query.Where("message_stats.created_at >= ?", time.Now().AddDate(0, 0, -30))
	default:
		query = query.Where("DATE(message_stats.created_at) = ?", time.Now().Format("2006-01-02"))
	}

	query.Select("p.name as provider_name, m.name as model_name, " +
		"COUNT(*) as `usage`, " +
		"(SUM(CASE WHEN message_stats.status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as `success_rate`").
		Group("p.name, m.name").
		Order("`usage` DESC").
		Scan(&stats)

	// 如果没有数据，返回空数组而不是 null
	if stats == nil {
		stats = make([]ProviderStats, 0)
	}

	c.JSON(http.StatusOK, stats)
}

// TokenRankStats Token消耗排行数据
type TokenRankStats struct {
	Username    string  `json:"username"`
	TotalTokens int64   `json:"totalTokens"`
	Percent     float64 `json:"percent"`
	Growth      float64 `json:"growth"` // 相比昨日的增长率
}

// GetTokenRanking 获取Token消耗排行
func (h *StatsHandler) GetTokenRanking(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	var stats []TokenRankStats
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 获取今日总消耗
	var totalTokens int64
	result := h.DB.Model(&m.MessageStats{}).
		Where("DATE(created_at) = ?", today).
		Select("COALESCE(SUM(total_tokens), 0) as total").
		Row().Scan(&totalTokens)

	if result != nil {
		//log.Printf("获取今日总消耗失败: %v", result)
	}
	//log.Printf("今日总消耗: %d", totalTokens)

	// 获取用户消耗排行
	query := h.DB.Model(&m.MessageStats{}).
		Select("u.username, "+
			"COALESCE(SUM(message_stats.total_tokens), 0) as total_tokens, "+
			"COALESCE((SELECT SUM(total_tokens) FROM message_stats "+
			"WHERE user_id = message_stats.user_id AND DATE(created_at) = ?), 0) as yesterday_tokens", yesterday).
		Joins("JOIN users u ON message_stats.user_id = u.id").
		Where("DATE(message_stats.created_at) = ?", today).
		Group("message_stats.user_id, u.username").
		Order("total_tokens DESC").
		Limit(10)

	if err := query.Scan(&stats).Error; err != nil {
		log.Printf("获取用户消耗排行失败: %v", err)
	}
	//log.Printf("查询到的用户数: %d", len(stats))

	// 计算百分比和增长率
	for i := range stats {
		if totalTokens > 0 {
			stats[i].Percent = float64(stats[i].TotalTokens) / float64(totalTokens) * 100
		}
		yesterdayTokens := stats[i].Growth // 这里的 Growth 字段暂时存储的是昨日数据
		if yesterdayTokens > 0 {
			stats[i].Growth = (float64(stats[i].TotalTokens) - yesterdayTokens) / yesterdayTokens * 100
		}
		//log.Printf("用户 %s: 今日消耗=%d, 昨日消耗=%d, 占比=%.2f%%, 增长率=%.2f%%",
		//	stats[i].Username, stats[i].TotalTokens, int64(yesterdayTokens),
		//	stats[i].Percent, stats[i].Growth)
	}

	// 如果没有数据，返回空数组而不是 null
	if stats == nil {
		stats = make([]TokenRankStats, 0)
	}

	c.JSON(http.StatusOK, stats)
}
