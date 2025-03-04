package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	m "AI_Proxy_Go/backend/internal/model"
	"log"
)

// ChatSessionHandler 会话管理处理器
type ChatSessionHandler struct {
	DB *gorm.DB
}

// ListSessions 获取用户的会话列表
func (h *ChatSessionHandler) ListSessions(c *gin.Context) {
	// 从查询参数获取用户ID
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	// 使用子查询统计消息数量
	var sessions []struct {
		m.ChatSession
		MessageCount int64 `json:"message_count"`
	}

	query := h.DB.Model(&m.ChatSession{}).
		Select("chat_sessions.*, COUNT(chat_messages.id) as message_count").
		Joins("LEFT JOIN chat_messages ON chat_messages.session_id = chat_sessions.id").
		Where("chat_sessions.user_id = ? AND chat_sessions.status = ?", userID, "active").
		Group("chat_sessions.id").
		Order("chat_sessions.updated_at DESC")

	if err := query.Scan(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话列表失败"})
		return
	}

	var result []gin.H
	for _, session := range sessions {
		result = append(result, gin.H{
			"id":           session.ID,
			"title":        session.Title,
			"messageCount": session.MessageCount,
			"createdAt":    session.CreatedAt,
			"updatedAt":    session.UpdatedAt,
			"status":       session.Status,
		})
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{
		"items": result,
		"total": len(sessions),
	})
}

// CreateSession 创建新会话
func (h *ChatSessionHandler) CreateSession(c *gin.Context) {
	var req struct {
		UserID     int64  `json:"userId"`
		Title      string `json:"title"`
		ProviderID int64  `json:"providerId"` // 添加默认值
		ModelID    int64  `json:"modelId"`    // 添加默认值
		KeyID      int64  `json:"keyId"`      // 添加默认值
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 如果没有指定，设置默认值
	if req.ProviderID == 0 {
		req.ProviderID = 13 // 设置默认服务商ID
	}
	if req.ModelID == 0 {
		req.ModelID = 10 // 设置默认模型ID
	}
	if req.KeyID == 0 {
		req.KeyID = 15 // 设置默认密钥ID
	}

	session := m.ChatSession{
		UserID:     req.UserID,
		Title:      req.Title,
		ProviderID: req.ProviderID,
		ModelID:    req.ModelID,
		KeyID:      req.KeyID,
		Status:     "active",
	}

	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, session)
}

// ArchiveSession 归档会话
func (h *ChatSessionHandler) ArchiveSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID := c.GetInt64("userId")

	if err := h.DB.Model(&m.ChatSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("status", "archived").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归档会话失败"})
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{"message": "会话已归档"})
}

// GetSessionMessages 获取会话的消息历史
func (h *ChatSessionHandler) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("id")
	// 从查询参数获取用户ID
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	//log.Printf("获取会话消息 - 会话ID: %s, 用户ID: %d", sessionID, userID)

	var messages []m.ChatMessage
	if err := h.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		log.Printf("查询消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息历史失败"})
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	//log.Printf("找到 %d 条消息", len(messages))
	c.JSON(http.StatusOK, messages)
}

// GetSession 获取会话详情
func (h *ChatSessionHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	//log.Printf("获取会话详情 - 会话ID: %s, 用户ID: %d", sessionID, userID)

	var session m.ChatSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		//log.Printf("查询会话失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话不存在"})
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, session)
}

// UpdateSession 更新会话信息
func (h *ChatSessionHandler) UpdateSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if err := h.DB.Model(&m.ChatSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("title", req.Title).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新会话失败"})
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteSession 删除会话及其消息
func (h *ChatSessionHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	// 开启事务
	tx := h.DB.Begin()

	// 删除会话消息
	if err := tx.Where("session_id = ? AND user_id = ?", sessionID, userID).
		Delete(&m.ChatMessage{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除会话消息失败"})
		return
	}

	// 删除会话
	if err := tx.Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&m.ChatSession{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除会话失败"})
		return
	}

	// 提交事务
	tx.Commit()

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
