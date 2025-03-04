package api

import (
	m "AI_Proxy_Go/backend/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ModelHandler 模型处理器
type ModelHandler struct {
	DB *gorm.DB
}

// List 获取模型列表
func (h *ModelHandler) List(c *gin.Context) {
	var models []m.AI_Model
	var total int64

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	// 查询总数
	h.DB.Model(&m.AI_Model{}).Count(&total)

	// 查询列表，包含服务商信息
	if err := h.DB.Preload("Provider").Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": models,
		"total": total,
	})
}

// Get 获取单个模型
func (h *ModelHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var model m.AI_Model

	if err := h.DB.Preload("Provider").First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}

	c.JSON(http.StatusOK, model)
}

// Create 创建模型
func (h *ModelHandler) Create(c *gin.Context) {
	var model m.AI_Model

	if err := c.ShouldBindJSON(&model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证必填字段
	if model.Name == "" || model.ModelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和模型ID为必填项"})
		return
	}

	// 验证服务商是否存在
	var provider m.AI_Provider
	if err := h.DB.First(&provider, model.ProviderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所选服务商不存在"})
		return
	}

	if err := h.DB.Create(&model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建模型失败"})
		return
	}

	c.JSON(http.StatusCreated, model)
}

// Update 更新模型
func (h *ModelHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var existingModel m.AI_Model
	var updateData m.AI_Model

	// 先获取现有记录
	if err := h.DB.First(&existingModel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}

	// 绑定更新数据
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证必填字段
	if updateData.Name == "" || updateData.ModelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和模型ID为必填项"})
		return
	}

	// 验证服务商是否存在
	var provider m.AI_Provider
	if err := h.DB.First(&provider, updateData.ProviderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所选服务商不存在"})
		return
	}

	// 更新记录
	if err := h.DB.Model(&existingModel).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新模型失败"})
		return
	}

	// 重新获取更新后的记录
	h.DB.Preload("Provider").First(&existingModel, id)
	c.JSON(http.StatusOK, existingModel)
}

// Delete 删除模型
func (h *ModelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var model m.AI_Model

	if err := h.DB.First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}

	if err := h.DB.Delete(&model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除模型失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
