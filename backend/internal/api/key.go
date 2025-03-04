package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	m "AI_Proxy_Go/backend/internal/model"
)

// KeyHandler 密钥处理器
type KeyHandler struct {
	DB *gorm.DB
}

// List 获取密钥列表
func (h *KeyHandler) List(c *gin.Context) {
	var total int64

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	// 查询总数
	h.DB.Model(&m.AI_APIKey{}).Count(&total)

	// 创建一个临时结构体来接收查询结果
	type KeyWithProvider struct {
		m.AI_APIKey
		ProviderName string
	}
	var results []KeyWithProvider

	// 查询列表，并关联查询服务商信息
	query := h.DB.Table("AI_api_keys").
		Select("AI_api_keys.*, AI_providers.name as provider_name").
		Joins("LEFT JOIN AI_providers ON AI_api_keys.provider_id = AI_providers.id")

	if err := query.Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取密钥列表失败"})
		return
	}

	// 构建响应数据
	var responseItems []gin.H
	for _, result := range results {
		item := gin.H{
			"ID":           result.ID,
			"Name":         result.Name,
			"Type":         result.Type,
			"UserID":       result.UserID,
			"ProviderID":   result.ProviderID,
			"ProviderName": result.ProviderName, // 从JOIN查询结果中获取
			"IsActive":     result.IsActive,
			"RateLimit":    result.RateLimit,
			"QuotaLimit":   result.QuotaLimit,
			"QuotaUsed":    result.QuotaUsed,
			"ExpireTime":   result.ExpireTime,
			"CreatorID":    result.CreatorID,
			"CreatorName":  result.CreatorName,
		}
		responseItems = append(responseItems, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": responseItems,
		"total": total,
	})
}

// Get 获取单个密钥详情
func (h *KeyHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var key m.AI_APIKey

	// 添加日志
	log.Printf("正在获取密钥详情 - ID: %s", id)

	if err := h.DB.First(&key, id).Error; err != nil {
		log.Printf("获取密钥失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "密钥不存在"})
		return
	}

	// 添加日志
	log.Printf("成功获取密钥 - ID: %d, Name: %s", key.ID, key.Name)

	c.JSON(http.StatusOK, key)
}

// Create 创建密钥
func (h *KeyHandler) Create(c *gin.Context) {
	var key m.AI_APIKey

	if err := c.ShouldBindJSON(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证必填字段
	if key.Name == "" || key.KeyValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和密钥为必填项"})
		return
	}

	if err := h.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建密钥失败"})
		return
	}

	c.JSON(http.StatusCreated, key)
}

// Update 更新密钥
func (h *KeyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var existingKey m.AI_APIKey
	var updateData m.AI_APIKey

	// 先获取现有记录
	if err := h.DB.First(&existingKey, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "密钥不存在"})
		return
	}

	// 绑定更新数据
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证必填字段
	if updateData.Name == "" || updateData.KeyValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和密钥为必填项"})
		return
	}

	// 更新记录
	if err := h.DB.Model(&existingKey).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密钥失败"})
		return
	}

	// 重新获取更新后的记录
	h.DB.First(&existingKey, id)
	c.JSON(http.StatusOK, existingKey)
}

// Delete 删除密钥
func (h *KeyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var key m.AI_APIKey

	if err := h.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "密钥不存在"})
		return
	}

	if err := h.DB.Delete(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// CheckKeyExists 检查密钥是否已存在
func (h *KeyHandler) CheckKeyExists(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密钥不能为空"})
		return
	}

	log.Printf("正在检查密钥是否存在: %s", key)

	var existingKey m.AI_APIKey
	err := h.DB.Where("`key_value` = ?", key).First(&existingKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("密钥不存在: %s", key)
			c.JSON(http.StatusOK, gin.H{
				"exists": false,
			})
			return
		}
		log.Printf("检查密钥时发生错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("检查密钥失败: %v", err)})
		return
	}

	log.Printf("找到已存在的密钥: ID=%d", existingKey.ID)
	c.JSON(http.StatusOK, gin.H{
		"exists": true,
		"id":     existingKey.ID,
	})
}
