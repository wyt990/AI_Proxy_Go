package api

import (
	"log"

	"AI_Proxy_Go/backend/internal/config"
	"AI_Proxy_Go/backend/internal/install"

	"github.com/gin-gonic/gin"
)

type InstallHandler struct {
	Installer *install.Installer
}

// 检查安装状态
func (h *InstallHandler) CheckInstallStatus(c *gin.Context) {
	if h.Installer.IsInstalled() {
		c.JSON(200, gin.H{"installed": true})
		return
	}
	c.JSON(200, gin.H{"installed": false})
}

// 执行安装
func (h *InstallHandler) Install(c *gin.Context) {
	var req struct {
		Database struct {
			Type     string `json:"type"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			DBName   string `json:"dbname"`
		} `json:"database"`
		Admin struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Name     string `json:"name"`
			Email    string `json:"email"`
		} `json:"admin"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// 检查环境
	if err := h.Installer.CheckEnvironment(); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 初始化数据库
	if err := h.Installer.InitDatabase(); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 创建管理员
	if err := h.Installer.CreateAdminUser(
		req.Admin.Username,
		req.Admin.Password,
		req.Admin.Name,
		req.Admin.Email,
	); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 完成安装
	if err := h.Installer.CompleteInstallation(); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Installation completed successfully"})
}

// 检查环境
func (h *InstallHandler) CheckEnvironment(c *gin.Context) {
	checks := h.Installer.CheckSystemEnvironment()

	// 添加日志，查看完整的返回数据结构
	log.Printf("Environment check result: %+v", checks)

	// 修改返回的数据结构，使其更扁平化
	c.JSON(200, gin.H{
		"success": len(checks.Errors) == 0,
		"items":   checks.Checks,
		"errors":  checks.Errors,
	})
}

// 测试数据库连接
func (h *InstallHandler) TestDatabase(c *gin.Context) {
	var req config.DatabaseConfig

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "invalid request"})
		return
	}

	// 测试数据库连接
	if err := h.Installer.TestDatabaseConnection(req); err != nil {
		c.JSON(200, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 保存数据库配置
	if err := h.Installer.SaveDatabaseConfig(req); err != nil {
		c.JSON(200, gin.H{"success": false, "error": "数据库配置保存失败: " + err.Error()})
		return
	}

	// 初始化数据库
	if err := h.Installer.InitDatabase(); err != nil {
		c.JSON(200, gin.H{"success": false, "error": "数据库初始化失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

// 测试Redis连接
func (h *InstallHandler) TestRedis(c *gin.Context) {
	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "invalid request"})
		return
	}

	if err := h.Installer.TestRedisConnection(req); err != nil {
		c.JSON(200, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 保存Redis配置到数据库
	if err := h.Installer.SaveRedisConfig(req); err != nil {
		c.JSON(200, gin.H{"success": false, "error": "Redis配置保存失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

// CompleteInstall 完成安装
func (h *InstallHandler) CompleteInstall(c *gin.Context) {
	var req struct {
		Admin struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Name     string `json:"name"`
			Email    string `json:"email"`
		} `json:"admin"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的请求参数"})
		return
	}

	// 创建管理员账户
	if err := h.Installer.CreateAdminUser(
		req.Admin.Username,
		req.Admin.Password,
		req.Admin.Name,
		req.Admin.Email,
	); err != nil {
		c.JSON(200, gin.H{"success": false, "error": "创建管理员账户失败: " + err.Error()})
		return
	}

	// 完成安装
	if err := h.Installer.CompleteInstallation(); err != nil {
		c.JSON(200, gin.H{"success": false, "error": "完成安装失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}
