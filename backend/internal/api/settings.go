package api

import (
	"AI_Proxy_Go/backend/internal/install"
	"AI_Proxy_Go/backend/internal/model"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	DB *gorm.DB
}

// GetRedisSettings 获取Redis设置
func (h *SettingsHandler) GetRedisSettings(c *gin.Context) {
	var settings []model.SystemSettings
	result := h.DB.Where("config_key LIKE ?", "redis.%").Find(&settings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取Redis设置失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为map格式返回
	settingsMap := make(map[string]string)
	for _, setting := range settings {
		settingsMap[setting.ConfigKey] = setting.Value
	}

	c.JSON(http.StatusOK, settingsMap)
}

// SaveRedisSettings 保存Redis设置
func (h *SettingsHandler) SaveRedisSettings(c *gin.Context) {
	var settings map[string]string
	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	for key, value := range settings {
		// 使用 Upsert 操作（更新或插入）
		result := tx.Exec(`
			INSERT INTO system_settings (config_key, value, description)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE value = ?
		`, key, value, getRedisSettingDescription(key), value)

		if result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存设置失败: " + result.Error.Error(),
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设置保存成功",
	})
}

// TestRedisConnection 测试Redis连接
func (h *SettingsHandler) TestRedisConnection(c *gin.Context) {
	var config struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	}

	if err := c.BindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		return
	}

	// 使用安装器中的测试函数
	installer := install.NewInstaller(nil) // 这里不需要配置
	err := installer.TestRedisConnection(config)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Redis连接测试成功",
	})
}

// GetSearchSettings 获取搜索设置
func (h *SettingsHandler) GetSearchSettings(c *gin.Context) {
	var settings []model.SystemSettings
	result := h.DB.Where("config_key LIKE ?", "search.%").Find(&settings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取搜索设置失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为map格式返回
	settingsMap := make(map[string]string)
	for _, setting := range settings {
		settingsMap[setting.ConfigKey] = setting.Value
	}

	// 使用默认值填充未设置的配置项
	for key, defaultValue := range model.DefaultSettingsValues {
		if strings.HasPrefix(key, "search.") {
			if _, exists := settingsMap[key]; !exists {
				settingsMap[key] = defaultValue
			}
		}
	}

	c.JSON(http.StatusOK, settingsMap)
}

// SaveSearchSettings 保存搜索设置
func (h *SettingsHandler) SaveSearchSettings(c *gin.Context) {
	var settings map[string]string
	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 验证设置值
	for key, value := range settings {
		// 验证数值类型的设置
		if isNumericSetting(key) {
			if _, err := strconv.Atoi(value); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("设置项 %s 必须是数字", key),
				})
				return
			}
		}

		// 使用 Upsert 操作（更新或插入）
		result := tx.Exec(`
			INSERT INTO system_settings (config_key, value, description)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE value = ?
		`, key, value, getSearchSettingDescription(key), value)

		if result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存设置失败: " + result.Error.Error(),
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设置保存成功",
	})
}

// isNumericSetting 判断设置项是否为数值类型
func isNumericSetting(key string) bool {
	numericSettings := map[string]bool{
		model.KeySearchResultCount:          true,
		model.KeySearchMaxKeywords:          true,
		model.KeySearchConcurrency:          true,
		model.KeySearchTimeout:              true,
		model.KeySearchRetryCount:           true,
		model.KeySearchCacheDuration:        true,
		model.KeySearchWeightOfficial:       true,
		model.KeySearchWeightEdu:            true,
		model.KeySearchWeightNews:           true,
		model.KeySearchWeightPortal:         true,
		model.KeySearchTimeWeightDay:        true,
		model.KeySearchTimeWeightWeek:       true,
		model.KeySearchTimeWeightMonth:      true,
		model.KeySearchTimeWeightYear:       true,
		model.KeySearchMinContentLength:     true,
		model.KeySearchMaxSummaryLength:     true,
		model.KeySearchMinTitleLength:       true,
		model.KeySearchMaxResults:           true,
		model.KeySearchCacheCleanupInterval: true,
		model.KeySearchRateLimit:            true,
		model.KeySearchConnectionTimeout:    true,
		model.KeySearchMaxRetryInterval:     true,
		model.KeySearchMaxCacheSize:         true,
		model.KeySessionContextLength:       true,
		model.KeyAIRequestTimeout:           true,
	}
	return numericSettings[key]
}

// 获取Redis设置的描述
func getRedisSettingDescription(key string) string {
	descriptions := map[string]string{
		"redis.host":     "Redis服务器地址",
		"redis.port":     "Redis端口",
		"redis.password": "Redis密码",
		"redis.db":       "Redis数据库索引",
	}
	return descriptions[key]
}

// 获取搜索设置的描述
func getSearchSettingDescription(key string) string {
	descriptions := map[string]string{
		// 基础搜索设置
		model.KeySearchEngineURL:   model.DefaultSettingsDescription[model.KeySearchEngineURL],
		model.KeySearchResultCount: model.DefaultSettingsDescription[model.KeySearchResultCount],
		model.KeySearchMaxKeywords: model.DefaultSettingsDescription[model.KeySearchMaxKeywords],
		model.KeySearchConcurrency: model.DefaultSettingsDescription[model.KeySearchConcurrency],

		// 性能设置
		model.KeySearchTimeout:       model.DefaultSettingsDescription[model.KeySearchTimeout],
		model.KeySearchRetryCount:    model.DefaultSettingsDescription[model.KeySearchRetryCount],
		model.KeySearchCacheDuration: model.DefaultSettingsDescription[model.KeySearchCacheDuration],

		// 权重设置
		model.KeySearchWeightOfficial: model.DefaultSettingsDescription[model.KeySearchWeightOfficial],
		model.KeySearchWeightEdu:      model.DefaultSettingsDescription[model.KeySearchWeightEdu],
		model.KeySearchWeightNews:     model.DefaultSettingsDescription[model.KeySearchWeightNews],
		model.KeySearchWeightPortal:   model.DefaultSettingsDescription[model.KeySearchWeightPortal],

		// 时效性权重设置
		model.KeySearchTimeWeightDay:   model.DefaultSettingsDescription[model.KeySearchTimeWeightDay],
		model.KeySearchTimeWeightWeek:  model.DefaultSettingsDescription[model.KeySearchTimeWeightWeek],
		model.KeySearchTimeWeightMonth: model.DefaultSettingsDescription[model.KeySearchTimeWeightMonth],
		model.KeySearchTimeWeightYear:  model.DefaultSettingsDescription[model.KeySearchTimeWeightYear],

		// 过滤设置
		model.KeySearchMinContentLength: model.DefaultSettingsDescription[model.KeySearchMinContentLength],
		model.KeySearchMaxSummaryLength: model.DefaultSettingsDescription[model.KeySearchMaxSummaryLength],
		model.KeySearchFilterDomains:    model.DefaultSettingsDescription[model.KeySearchFilterDomains],

		// 搜索质量控制设置
		model.KeySearchMinTitleLength:       model.DefaultSettingsDescription[model.KeySearchMinTitleLength],
		model.KeySearchMaxResults:           model.DefaultSettingsDescription[model.KeySearchMaxResults],
		model.KeySearchMinRelevanceScore:    model.DefaultSettingsDescription[model.KeySearchMinRelevanceScore],
		model.KeySearchCacheCleanupInterval: model.DefaultSettingsDescription[model.KeySearchCacheCleanupInterval],
		model.KeySearchRateLimit:            model.DefaultSettingsDescription[model.KeySearchRateLimit],
		model.KeySearchConnectionTimeout:    model.DefaultSettingsDescription[model.KeySearchConnectionTimeout],
		model.KeySearchMaxRetryInterval:     model.DefaultSettingsDescription[model.KeySearchMaxRetryInterval],
		model.KeySearchMaxCacheSize:         model.DefaultSettingsDescription[model.KeySearchMaxCacheSize],
	}
	return descriptions[key]
}

// GetChatSettings 获取对话设置
func (h *SettingsHandler) GetChatSettings(c *gin.Context) {
	var settings = make(map[string]string)

	// 需要获取的设置键列表
	keys := []string{
		model.KeyAIRequestTimeout,
		model.KeySessionContextLength,
	}

	// 查询数据库中的设置
	var dbSettings []model.SystemSettings
	if err := h.DB.Where("config_key IN ?", keys).Find(&dbSettings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设置失败"})
		return
	}

	// 先填充默认值
	for _, key := range keys {
		settings[key] = model.DefaultSettingsValues[key]
	}

	// 用数据库中的值覆盖默认值
	for _, setting := range dbSettings {
		settings[setting.ConfigKey] = setting.Value
	}

	c.JSON(http.StatusOK, settings)
}

// SaveChatSettings 保存对话设置
func (h *SettingsHandler) SaveChatSettings(c *gin.Context) {
	var settings map[string]string
	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 验证设置值
	for key, value := range settings {
		// 验证数值类型的设置
		if isNumericSetting(key) {
			if _, err := strconv.Atoi(value); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("设置项 %s 必须是数字", key),
				})
				return
			}
		}

		// 使用 Upsert 操作（更新或插入）
		result := tx.Exec(`
			INSERT INTO system_settings (config_key, value, description)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE value = ?
		`, key, value, getChatSettingDescription(key), value)

		if result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存设置失败: " + result.Error.Error(),
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设置保存成功",
	})
}

// 获取对话设置的描述
func getChatSettingDescription(key string) string {
	descriptions := map[string]string{
		model.KeyAIRequestTimeout:     model.DefaultSettingsDescription[model.KeyAIRequestTimeout],
		model.KeySessionContextLength: model.DefaultSettingsDescription[model.KeySessionContextLength],
	}
	return descriptions[key]
}
