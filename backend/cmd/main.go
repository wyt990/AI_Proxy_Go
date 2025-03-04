package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"AI_Proxy_Go/backend/internal/api"
	"AI_Proxy_Go/backend/internal/config"
	"AI_Proxy_Go/backend/internal/install"
	"AI_Proxy_Go/backend/internal/middleware"
	"AI_Proxy_Go/backend/internal/model"
	"AI_Proxy_Go/backend/internal/service"
	"AI_Proxy_Go/backend/internal/service/search"
	"AI_Proxy_Go/backend/internal/version"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	configFile string
)

func init() {
	// 命令行参数，默认使用项目目录下的配置文件
	defaultConfig := filepath.Join("config", "config.yaml")
	flag.StringVar(&configFile, "config", defaultConfig, "配置文件路径")
	flag.Parse()
}

func main() {
	// 获取项目根目录
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前目录失败: %v", err)
	}

	// 加载配置文件
	configPath := filepath.Join(rootDir, configFile)
	log.Printf("使用配置文件: %s", configPath)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 添加调试日志
	log.Printf("当前配置: %+v", cfg)

	// 初始化安装器
	installer := install.NewInstaller(cfg)

	// 添加调试日志
	log.Printf("BasePath: %s", cfg.BasePath)

	// 检查是否已安装
	isInstalled := installer.IsInstalled()
	log.Printf("安装状态检查结果: %v", isInstalled)

	// 创建Gin引擎
	r := gin.Default()

	// 加载静态文件
	r.Static("/static", "./frontend/static")

	// 添加 favicon.ico 路由
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/static/images/favicon.ico")
	})

	// 加载HTML模板
	r.LoadHTMLGlob("frontend/templates/**/*.html")

	// 注册安装检查中间件
	r.Use(middleware.InstallCheck(installer))

	// 创建验证码处理器
	captchaHandler := api.NewCaptchaHandler()

	// 创建版本号处理器
	versionHandler := api.NewVersionHandler(version.Version, version.BuildTime)

	// 不需要认证的路由组
	public := r.Group("")
	{
		// 安装页面路由
		public.GET("/install", func(c *gin.Context) {
			c.HTML(http.StatusOK, "install/index", nil)
		})

		// 安装相关API路由
		installHandler := &api.InstallHandler{
			Installer: installer,
		}
		public.GET("/api/install/status", installHandler.CheckInstallStatus)
		public.POST("/api/install", installHandler.Install)
		public.GET("/api/install/check-environment", installHandler.CheckEnvironment)
		public.POST("/api/install/test-database", installHandler.TestDatabase)
		public.POST("/api/install/test-redis", installHandler.TestRedis)
		public.POST("/api/install/complete", installHandler.CompleteInstall)

		// 验证码路由
		public.GET("/api/captcha/generate", captchaHandler.GenerateCaptcha)

		// 版本号路由
		public.GET("/api/version", versionHandler.GetVersion)

		// 系统配置相关路由
		public.GET("/api/system/config", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"siteName": config.SiteName,
			})
		})

		// 登录相关路由
		public.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "login/index", getLoginTemplateData())
		})
	}

	// 检查是否已安装
	if isInstalled {
		log.Printf("系统已安装，初始化数据库连接")
		log.Printf("数据库配置: %+v", cfg.Database)

		// 构建DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Database,
		)
		log.Printf("数据库DSN: %s", dsn)

		// 连接数据库
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("连接数据库失败: %v", err)
		}

		// 获取底层的数据库连接以配置连接池
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("获取数据库实例失败: %v", err)
		}

		// 设置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		// 自动迁移数据库表结构
		log.Printf("开始自动迁移数据库表结构...")
		if err := db.AutoMigrate(
			&model.AI_Provider{},
			&model.AI_Model{},
			&model.AI_APIKey{},
			&model.ChatMessage{},
			&model.ChatSession{},
			&model.SystemSettings{},
			&model.User{},
			&model.MessageStats{},  // 添加消息统计表
			&model.SystemMetrics{}, // 添加系统指标表
		); err != nil {
			log.Fatalf("自动迁移失败: %v", err)
		}
		log.Printf("数据库表结构迁移完成")

		// 创建认证处理器
		authHandler := &api.AuthHandler{
			DB:      db,
			Captcha: captchaHandler,
		}

		// 注册认证中间件
		r.Use(middleware.AuthMiddleware())

		// 注册认证相关路由
		r.POST("/api/login", authHandler.Login)
		r.POST("/api/logout", authHandler.Logout)

		// 需要认证的路由组
		authorized := r.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			// 用户信息API
			authorized.GET("/api/user/info", authHandler.GetUserInfo)

			// 用户管理API
			userHandler := &api.UserHandler{DB: db}
			authorized.GET("/api/users", userHandler.GetUsers)
			authorized.GET("/api/users/:id", userHandler.GetUser)
			authorized.PUT("/api/users/:id", userHandler.UpdateUser)
			authorized.DELETE("/api/users/:id", userHandler.DeleteUser)
			authorized.POST("/api/users", userHandler.CreateUser)

			// 页面路由
			authorized.GET("/home", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("home/index", "控制台"))
			})

			authorized.GET("/users", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("users/index", "用户管理"))
			})

			// 服务商管理API
			providerHandler := &api.ProviderHandler{DB: db}
			authorized.GET("/api/providers", providerHandler.List)
			authorized.GET("/api/providers/:id", providerHandler.Get)
			authorized.POST("/api/providers", providerHandler.Create)
			authorized.PUT("/api/providers/:id", providerHandler.Update)
			authorized.DELETE("/api/providers/:id", providerHandler.Delete)
			authorized.POST("/api/providers/:id/check", providerHandler.Check)

			// 服务商管理页面
			authorized.GET("/providers", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("providers/index", "服务商管理"))
			})

			// 模型管理API
			modelHandler := &api.ModelHandler{DB: db}
			authorized.GET("/api/models", modelHandler.List)
			authorized.GET("/api/models/:id", modelHandler.Get)
			authorized.POST("/api/models", modelHandler.Create)
			authorized.PUT("/api/models/:id", modelHandler.Update)
			authorized.DELETE("/api/models/:id", modelHandler.Delete)

			// 模型管理页面
			authorized.GET("/AI_models", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("models/index", "模型管理"))
			})

			// 密钥管理API
			keyHandler := &api.KeyHandler{DB: db}
			authorized.GET("/api/keys/check", keyHandler.CheckKeyExists)
			authorized.GET("/api/keys", keyHandler.List)
			authorized.GET("/api/keys/:id", keyHandler.Get)
			authorized.POST("/api/keys", keyHandler.Create)
			authorized.PUT("/api/keys/:id", keyHandler.Update)
			authorized.DELETE("/api/keys/:id", keyHandler.Delete)

			// 密钥管理页面
			authorized.GET("/AI_keys", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("keys/index", "密钥管理"))
			})

			// 对话管理API
			chatHandler := api.NewChatHandler(db)
			if chatHandler == nil {
				log.Fatal("ChatHandler初始化失败")
			}
			//log.Printf("ChatHandler 初始化成功")

			authorized.GET("/api/chat/providers", chatHandler.GetProviders)
			authorized.GET("/api/chat/provider/:id/models", chatHandler.GetProviderModels)
			authorized.GET("/api/chat/provider/:id/keys", chatHandler.GetProviderKeys)
			authorized.POST("/api/chat/messages", chatHandler.SendMessage)
			authorized.GET("/api/chat/history", chatHandler.GetHistory)
			authorized.DELETE("/api/chat/messages/:id", chatHandler.DeleteMessage)

			// 会话管理API
			chatSessionHandler := &api.ChatSessionHandler{DB: db}
			authorized.GET("/api/chat/sessions", chatSessionHandler.ListSessions)
			authorized.POST("/api/chat/sessions", chatSessionHandler.CreateSession)
			authorized.GET("/api/chat/sessions/:id", chatSessionHandler.GetSession)
			authorized.PUT("/api/chat/sessions/:id", chatSessionHandler.UpdateSession)
			authorized.DELETE("/api/chat/sessions/:id", chatSessionHandler.DeleteSession)
			authorized.GET("/api/chat/sessions/:id/messages", chatSessionHandler.GetSessionMessages)
			authorized.POST("/api/chat/sessions/:id/archive", chatSessionHandler.ArchiveSession)

			// 对话页面
			authorized.GET("/chat", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("chat/index", "AI对话"))
			})

			// 日志查看页面
			authorized.GET("/logs", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("logs/index", "日志查看"))
			})

			// 系统设置相关API
			settingsHandler := &api.SettingsHandler{DB: db}
			authorized.GET("/api/settings/redis", settingsHandler.GetRedisSettings)
			authorized.POST("/api/settings/redis", settingsHandler.SaveRedisSettings)
			authorized.POST("/api/settings/redis/test", settingsHandler.TestRedisConnection)
			// 搜索设置相关API
			authorized.GET("/api/settings/search", settingsHandler.GetSearchSettings)
			authorized.POST("/api/settings/search", settingsHandler.SaveSearchSettings)

			// 搜索相关API
			searchEngine := search.NewSearchEngine(db) // 只需要传入数据库连接
			searchHandler := api.NewSearchHandler(searchEngine, &service.AIService{DB: db})
			authorized.GET("/api/search", searchHandler.Search)                // 执行搜索
			authorized.POST("/api/search/process", searchHandler.ProcessQuery) // 处理查询
			authorized.POST("/api/search/filter", searchHandler.FilterResults) // 过滤结果

			// AI请求控制相关API
			authorized.GET("/api/settings/chat", settingsHandler.GetChatSettings)
			authorized.POST("/api/settings/chat", settingsHandler.SaveChatSettings)

			// 系统设置页面
			authorized.GET("/settings", func(c *gin.Context) {
				c.HTML(http.StatusOK, "base", getTemplateData("settings/index", "系统设置"))
			})

			// 系统指标相关API
			metricsService := service.NewMetricsService(db)
			metricsHandler := api.NewMetricsHandler(metricsService)
			authorized.GET("/api/metrics/latest", metricsHandler.GetLatestMetrics)

			// 统计数据相关API
			statsHandler := api.NewStatsHandler(db)
			authorized.GET("/api/stats/dashboard", statsHandler.GetDashboardStats)
			authorized.GET("/api/stats/tokens", statsHandler.GetTokenStats)
			authorized.GET("/api/stats/model-usage", statsHandler.GetModelUsage)
			authorized.GET("/api/stats/request-monitor", statsHandler.GetRequestMonitor)
			authorized.GET("/api/stats/provider-stats", statsHandler.GetProviderStats)
			authorized.GET("/api/stats/token-ranking", statsHandler.GetTokenRanking)
		}

		// 初始化系统指标服务
		metricsService := service.NewMetricsService(db)

		// 启动系统指标收集（每分钟收集一次）
		metricsService.StartMetricsCollection(time.Minute)

		// 每天清理30天前的数据
		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			for range ticker.C {
				if err := metricsService.CleanOldMetrics(30 * 24 * time.Hour); err != nil {
					log.Printf("清理系统指标数据失败: %v", err)
				}
			}
		}()
	} else {
		log.Printf("系统未安装，跳过数据库连接")
	}

	// 添加根路由处理
	r.GET("/", func(c *gin.Context) {
		if !installer.IsInstalled() {
			c.Redirect(302, "/install")
			return
		}

		// 检查是否已登录
		token := c.GetHeader("Authorization")
		if token == "" {
			// 从 cookie 中获取 token
			token, _ = c.Cookie("token")
		}

		if token == "" {
			// 未登录时重定向到登录页面
			c.Redirect(302, "/login")
			return
		}

		// 已登录则重定向到主页
		c.Redirect(302, "/home")
	})

	// 启动服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("服务启动于: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// 生成通用的模板数据
func getTemplateData(template, pageTitle string) gin.H {
	return gin.H{
		"template":  template,
		"title":     config.SiteName, // 网站标题
		"pageTitle": pageTitle,       // 页面标题
		"config": gin.H{
			"siteName": config.SiteName,
		},
	}
}

// 专门用于登录页面的模板数据
func getLoginTemplateData() gin.H {
	return gin.H{
		"title":     config.SiteName, // 网站标题
		"pageTitle": "登录",            // 页面标题
	}
}
