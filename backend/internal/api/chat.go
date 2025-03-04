package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	m "AI_Proxy_Go/backend/internal/model"
	"AI_Proxy_Go/backend/internal/service"
	"encoding/json"
	"log"
)

// ChatHandler 对话处理器
type ChatHandler struct {
	DB        *gorm.DB
	aiService *service.AIService
}

// NewChatHandler 创建新的聊天处理器
func NewChatHandler(db *gorm.DB) *ChatHandler {
	if db == nil {
		log.Printf("数据库连接为空")
		return nil
	}

	log.Printf("开始初始化 ChatHandler")
	aiService := service.NewAIService(db)
	if aiService == nil {
		log.Printf("AIService 初始化失败")
		return nil
	}
	log.Printf("AIService 初始化成功")

	handler := &ChatHandler{
		DB:        db,
		aiService: aiService,
	}

	log.Printf("ChatHandler 初始化完成")
	return handler
}

// GetProviders 获取可用的服务商列表
func (h *ChatHandler) GetProviders(c *gin.Context) {
	var providers []m.AI_Provider
	if err := h.DB.Where("status = ?", "NORMAL").Find(&providers).Error; err != nil {
		//log.Printf("查询服务商失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取服务商列表失败"})
		return
	}

	//log.Printf("找到 %d 个服务商", len(providers))

	// 构建响应数据
	var items []gin.H
	for _, p := range providers {
		items = append(items, gin.H{
			"id":   p.ID,   // 使用小写的id
			"name": p.Name, // 使用小写的name
			"type": p.Type,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

// GetProviderModels 获取服务商的模型列表
func (h *ChatHandler) GetProviderModels(c *gin.Context) {
	providerId := c.Param("id")
	var models []m.AI_Model

	providerIDUint, err := strconv.ParseUint(providerId, 10, 32)
	if err != nil {
		//log.Printf("服务商ID转换失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务商ID"})
		return
	}

	//log.Printf("查询模型列表 - 服务商ID: %d", providerIDUint)

	if err := h.DB.Where("provider_id = ? AND status = ?", uint(providerIDUint), "NORMAL").Find(&models).Error; err != nil {
		//log.Printf("查询失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型列表失败"})
		return
	}

	//log.Printf("找到 %d 个模型", len(models))

	var items []gin.H
	for _, m := range models {
		//log.Printf("模型数据: ID=%d, Name=%s, Parameters=%s", m.ID, m.Name, m.Parameters)

		items = append(items, gin.H{
			"id":          m.ID,
			"name":        m.Name,
			"modelId":     m.ModelID,
			"description": m.Description,
			"maxTokens":   m.MaxTokens,
			"parameters":  m.Parameters, // 确保返回参数字段
		})
	}

	//log.Printf("返回数据: %+v", items)

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

// GetProviderKeys 获取服务商的可用密钥
func (h *ChatHandler) GetProviderKeys(c *gin.Context) {
	providerId := c.Param("id")
	userId := c.GetInt64("userId") // 从JWT中获取用户ID

	var publicKeys, privateKeys []m.AI_APIKey

	// 查询公钥
	h.DB.Where("provider_id = ? AND type = ? AND is_active = ?",
		providerId, "PUBLIC", true).Find(&publicKeys)

	// 查询用户的私钥
	h.DB.Where("provider_id = ? AND type = ? AND user_id = ? AND is_active = ?",
		providerId, "PRIVATE", userId, true).Find(&privateKeys)

	c.JSON(http.StatusOK, gin.H{
		"publicKeys":  publicKeys,
		"privateKeys": privateKeys,
	})
}

// SendMessage 发送对话消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var message struct {
		SessionID  uint                   `json:"sessionId"`
		ProviderID int64                  `json:"providerId"`
		ModelID    int64                  `json:"modelId"`
		KeyID      int64                  `json:"keyId"`
		Content    string                 `json:"content"`
		UserID     int64                  `json:"userId"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证会话存在且属于当前用户
	var session m.ChatSession
	if err := h.DB.First(&session, message.SessionID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话不存在"})
		return
	}

	if session.UserID != message.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此会话"})
		return
	}

	// 保存用户消息
	userMessage := m.ChatMessage{
		SessionID:  message.SessionID,
		UserID:     message.UserID,
		ProviderID: message.ProviderID,
		ModelID:    message.ModelID,
		KeyID:      message.KeyID,
		Role:       "user",
		Content:    message.Content,
	}

	if err := h.DB.Create(&userMessage).Error; err != nil {
		//log.Printf("保存用户消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存消息失败"})
		return
	}
	//log.Printf("用户消息已保存 - ID: %d, Content: %s", userMessage.ID, userMessage.Content)

	// 获取模型信息，检查是否使用流式传输
	var model m.AI_Model
	if err := h.DB.First(&model, message.ModelID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模型不存在"})
		return
	}

	// 解析模型参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(model.Parameters), &params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析模型参数失败"})
		return
	}

	// 检查是否使用流式传输
	useStream := false
	if stream, ok := params["stream"].(bool); ok {
		useStream = stream
	}

	// 记录开始时间
	startTime := time.Now()

	if useStream {
		// 设置SSE响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		// 发送开始事件
		c.SSEvent("message", gin.H{
			"type":          "start",
			"userMessageId": userMessage.ID,
		})
		c.Writer.Flush()

		// 发送消息
		response, err := h.aiService.SendChatMessage(&service.ChatRequest{
			SessionID:  message.SessionID,
			ProviderID: message.ProviderID,
			ModelID:    message.ModelID,
			KeyID:      message.KeyID,
			Content:    message.Content,
			UserID:     message.UserID,
			Parameters: message.Parameters,
		})

		if err != nil {
			c.SSEvent("message", gin.H{
				"type":  "error",
				"error": err.Error(),
			})
			return
		}

		// 检查是否是流式响应
		if streamResp, ok := response.(*service.StreamResponse); ok {
			// 处理流式数据
			for {
				select {
				case content, ok := <-streamResp.Stream:
					if !ok {
						continue
					}
					// 发送内容事件
					c.SSEvent("message", gin.H{
						"type":    "content",
						"content": content,
					})
					c.Writer.Flush()
				case <-streamResp.Done:
					// 记录会话统计
					stats := &m.MessageStats{
						UserID:           message.UserID,
						ProviderID:       message.ProviderID,
						ModelID:          message.ModelID,
						SessionID:        message.SessionID,
						PromptTokens:     streamResp.PromptTokens,
						CompletionTokens: streamResp.CompletionTokens,
						TotalTokens:      streamResp.TotalTokens,
						ResponseTime:     time.Since(startTime).Milliseconds(),
						CreatedAt:        startTime,
					}

					// 异步保存统计数据
					go func(stats *m.MessageStats) {
						if err := h.DB.Create(stats).Error; err != nil {
							log.Printf("记录消息统计失败: %v", err)
						}
					}(stats)

					c.SSEvent("message", gin.H{
						"type":               "end",
						"assistantMessageId": streamResp.MessageID,
						"promptTokens":       streamResp.PromptTokens,
						"completionTokens":   streamResp.CompletionTokens,
						"totalTokens":        streamResp.TotalTokens,
					})
					c.Writer.Flush()
					return
				case <-c.Request.Context().Done():
					return
				}
			}
		}
	} else {
		// 调用AI服务
		response, err := h.aiService.SendChatMessage(&service.ChatRequest{
			SessionID:  message.SessionID,
			ProviderID: message.ProviderID,
			ModelID:    message.ModelID,
			KeyID:      message.KeyID,
			Content:    message.Content,
			UserID:     message.UserID,
			Parameters: message.Parameters,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 类型断言，获取完整响应
		chatResp, ok := response.(*service.ChatResponse)
		if !ok {
			log.Printf("响应类型错误: 期望ChatResponse，获得%T", response)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "响应类型错误"})
			return
		}

		// 保存AI响应
		assistantMessage := m.ChatMessage{
			SessionID:        message.SessionID,
			UserID:           message.UserID,
			ProviderID:       message.ProviderID,
			ModelID:          message.ModelID,
			KeyID:            message.KeyID,
			Role:             "assistant",
			Content:          chatResp.Content,
			PromptTokens:     chatResp.PromptTokens,
			CompletionTokens: chatResp.CompletionTokens,
			TotalTokens:      chatResp.TotalTokens,
		}

		if err := h.DB.Create(&assistantMessage).Error; err != nil {
			log.Printf("保存AI响应失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存响应失败"})
			return
		}
		log.Printf("AI普通响应已保存，消息ID: %d", assistantMessage.ID)

		// 记录会话统计
		stats := &m.MessageStats{
			UserID:           message.UserID,
			ProviderID:       message.ProviderID,
			ModelID:          message.ModelID,
			SessionID:        message.SessionID,
			PromptTokens:     chatResp.PromptTokens,
			CompletionTokens: chatResp.CompletionTokens,
			TotalTokens:      chatResp.TotalTokens,
			ResponseTime:     time.Since(startTime).Milliseconds(),
			CreatedAt:        startTime,
		}

		// 异步保存统计数据
		go func(stats *m.MessageStats) {
			if err := h.DB.Create(stats).Error; err != nil {
				log.Printf("记录消息统计失败: %v", err)
			}
		}(stats)

		// 普通JSON响应
		c.JSON(http.StatusOK, gin.H{
			"content":            chatResp.Content,
			"userMessageId":      userMessage.ID,
			"assistantMessageId": assistantMessage.ID,
		})
	}

	// 更新会话的 updated_at 时间
	if err := h.DB.Model(&m.ChatSession{}).
		Where("id = ?", message.SessionID).
		Update("updated_at", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新会话时间失败"})
		return
	}
}

// GetHistory 获取对话历史
func (h *ChatHandler) GetHistory(c *gin.Context) {
	userId := c.GetInt64("userId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	var messages []m.ChatMessage
	var total int64

	h.DB.Model(&m.ChatMessage{}).Where("user_id = ?", userId).Count(&total)

	if err := h.DB.Where("user_id = ?", userId).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": messages,
		"total": total,
	})
}

// DeleteMessage 删除消息
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("id")
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	//log.Printf("开始删除消息 - MessageID: %s, UserID: %d", messageID, userID)

	// 验证消息存在且属于当前用户
	var message m.ChatMessage
	if err := h.DB.Where("id = ? AND user_id = ?", messageID, userID).First(&message).Error; err != nil {
		//log.Printf("消息查询失败 - MessageID: %s, UserID: %d, Error: %v", messageID, userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "消息不存在或无权删除"})
		return
	}
	//log.Printf("找到待删除消息 - MessageID: %s, SessionID: %d, Content: %s", messageID, message.SessionID, message.Content)

	// 删除消息
	if err := h.DB.Delete(&message).Error; err != nil {
		//log.Printf("删除消息失败 - MessageID: %s, Error: %v", messageID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除消息失败"})
		return
	}
	//log.Printf("消息删除成功 - MessageID: %s", messageID)

	// 更新会话的消息计数
	if err := h.DB.Model(&m.ChatSession{}).Where("id = ?", message.SessionID).
		UpdateColumn("message_count", gorm.Expr("message_count - 1")).Error; err != nil {
		//log.Printf("更新会话消息计数失败 - SessionID: %d, Error: %v", message.SessionID, err)
	} else {
		//log.Printf("会话消息计数更新成功 - SessionID: %d", message.SessionID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetSessionMessages 获取会话消息
func (h *ChatHandler) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("id")
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	//log.Printf("获取会话消息 - SessionID: %s, UserID: %d", sessionID, userID)

	var messages []m.ChatMessage
	if err := h.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		//log.Printf("查询消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}

	//log.Printf("找到 %d 条消息", len(messages))

	// 直接返回消息数组
	c.JSON(http.StatusOK, messages)
}
