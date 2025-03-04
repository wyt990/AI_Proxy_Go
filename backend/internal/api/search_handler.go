package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"AI_Proxy_Go/backend/internal/model"
	"AI_Proxy_Go/backend/internal/service"
	"AI_Proxy_Go/backend/internal/service/search"
)

// SearchHandler 处理搜索相关的API请求
type SearchHandler struct {
	searchEngine search.SearchService // 使用接口类型
	aiService    *service.AIService
}

// NewSearchHandler 创建新的搜索处理器
func NewSearchHandler(searchEngine search.SearchService, aiService *service.AIService) *SearchHandler {
	return &SearchHandler{
		searchEngine: searchEngine,
		aiService:    aiService,
	}
}

// Search 处理搜索请求
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "搜索查询不能为空",
		})
		return
	}

	// 执行搜索
	result, err := h.searchEngine.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ProcessQuery 处理搜索查询
func (h *SearchHandler) ProcessQuery(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		return
	}

	// 处理查询
	processedQuery, err := h.searchEngine.ProcessQuery(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"processed_query": processedQuery,
	})
}

// FilterResults 过滤搜索结果
func (h *SearchHandler) FilterResults(c *gin.Context) {
	var req struct {
		Results *model.SearchResult `json:"results" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		return
	}

	// 过滤结果
	filteredResults, err := h.searchEngine.FilterResults(req.Results)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "过滤结果失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, filteredResults)
}
