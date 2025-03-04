package api

import (
	"AI_Proxy_Go/backend/internal/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProviderHandler 服务商处理器
type ProviderHandler struct {
	DB *gorm.DB
}

// List 获取服务商列表
func (h *ProviderHandler) List(c *gin.Context) {
	var providers []model.AI_Provider
	var total int64

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	// 查询总数
	h.DB.Model(&model.AI_Provider{}).Count(&total)

	// 查询列表
	if err := h.DB.Offset(offset).Limit(pageSize).Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取服务商列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": providers,
		"total": total,
	})
}

// Get 获取单个服务商
func (h *ProviderHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var provider model.AI_Provider

	if err := h.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务商不存在"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// Create 创建服务商
func (h *ProviderHandler) Create(c *gin.Context) {
	var provider model.AI_Provider

	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证必填字段
	if provider.Name == "" || provider.Type == "" || provider.BaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称、类型和基础URL为必填项"})
		return
	}

	// 验证JSON Schema格式
	if err := validateJSONSchema(provider.RequestFormat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式Schema无效"})
		return
	}
	if err := validateJSONSchema(provider.ResponseFormat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "响应格式Schema无效"})
		return
	}
	if err := validateJSONSchema(provider.AuthFormat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "认证格式Schema无效"})
		return
	}

	if err := h.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建服务商失败"})
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// Update 更新服务商
func (h *ProviderHandler) Update(c *gin.Context) {
	id := c.Param("id")
	// 将字符串ID转换为uint
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var existingProvider model.AI_Provider
	var updateData model.AI_Provider

	// 先获取现有记录
	if err := h.DB.First(&existingProvider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务商不存在"})
		return
	}

	// 绑定更新数据
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Printf("绑定JSON数据失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 确保ID不变
	updateData.ID = uint(idUint)

	// 验证必填字段
	if updateData.Name == "" || updateData.Type == "" || updateData.BaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称、类型和基础URL为必填项"})
		return
	}

	// 验证JSON Schema格式
	if err := validateJSONSchema(updateData.RequestFormat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式Schema无效"})
		return
	}
	if err := validateJSONSchema(updateData.ResponseFormat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "响应格式Schema无效"})
		return
	}
	if err := validateJSONSchema(updateData.AuthFormat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "认证格式Schema无效"})
		return
	}

	// 更新记录，排除零值字段
	if err := h.DB.Model(&existingProvider).Updates(map[string]interface{}{
		"name":            updateData.Name,
		"type":            updateData.Type,
		"base_url":        updateData.BaseURL,
		"request_format":  updateData.RequestFormat,
		"response_format": updateData.ResponseFormat,
		"auth_format":     updateData.AuthFormat,
		"headers":         updateData.Headers,
		"auth_type":       updateData.AuthType,
		"proxy_mode":      updateData.ProxyMode,
		"rate_limit":      updateData.RateLimit,
		"timeout":         updateData.Timeout,
		"retry_count":     updateData.RetryCount,
		"retry_interval":  updateData.RetryInterval,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新服务商失败"})
		return
	}

	// 重新获取更新后的记录
	h.DB.First(&existingProvider, id)
	c.JSON(http.StatusOK, existingProvider)
}

// Delete 删除服务商
func (h *ProviderHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var provider model.AI_Provider

	// 检查是否存在
	if err := h.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务商不存在"})
		return
	}

	// 检查是否有关联的模型
	var modelCount int64
	if err := h.DB.Model(&model.AI_Model{}).Where("provider_id = ?", id).Count(&modelCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查关联模型失败"})
		return
	}

	if modelCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该服务商下还有关联的模型，无法删除"})
		return
	}

	if err := h.DB.Delete(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除服务商失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// Check 执行健康检查
func (h *ProviderHandler) Check(c *gin.Context) {
	id := c.Param("id")
	var provider model.AI_Provider

	if err := h.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务商不存在"})
		return
	}

	// 执行健康检查
	err := h.checkProviderHealth(&provider)
	if err != nil {
		provider.Status = "ERROR"
		provider.LastError = err.Error()
	} else {
		provider.Status = "NORMAL"
		provider.LastError = ""
	}

	provider.LastCheckTime = time.Now()

	// 更新状态
	if err := h.DB.Save(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// 健康检查实现
func (h *ProviderHandler) checkProviderHealth(provider *model.AI_Provider) error {
	client := &http.Client{
		Timeout: time.Duration(provider.Timeout) * time.Second,
	}

	// 构建测试请求
	req, err := http.NewRequest("GET", provider.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("构建请求失败: %v", err)
	}

	// 添加自定义请求头
	if provider.Headers != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(provider.Headers), &headers); err != nil {
			return fmt.Errorf("解析请求头失败: %v", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode >= 400 {
		return fmt.Errorf("服务返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// validateJSONSchema 验证JSON Schema格式
func validateJSONSchema(schema string) error {
	if schema == "" {
		return fmt.Errorf("schema cannot be empty")
	}

	// 解析 JSON Schema
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaMap); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	// 验证基本结构
	if _, ok := schemaMap["type"]; !ok {
		return fmt.Errorf("schema must have 'type' field")
	}

	// 验证必需字段
	requiredFields := []string{"type", "properties"}
	for _, field := range requiredFields {
		if _, ok := schemaMap[field]; !ok {
			return fmt.Errorf("schema missing required field: %s", field)
		}
	}

	// 验证属性
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid properties format")
	}

	// 验证每个属性的类型
	for name, prop := range properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid property format for: %s", name)
		}

		// 检查属性类型
		if _, ok := propMap["type"]; !ok {
			return fmt.Errorf("property %s missing type", name)
		}
	}

	return nil
}
