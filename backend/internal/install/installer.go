package install

import (
	"AI_Proxy_Go/backend/internal/config"
	"AI_Proxy_Go/backend/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Installer struct {
	config *config.Config
	db     *gorm.DB
}

// 构造函数
func NewInstaller(cfg *config.Config) *Installer {
	return &Installer{
		config: cfg,
	}
}

// 检查是否已安装
func (i *Installer) IsInstalled() bool {
	// 使用项目根目录
	lockFile := filepath.Join(i.config.BasePath, "install.lock")

	// 添加调试日志
	// log.Printf("检查安装状态...")
	// log.Printf("项目根目录: %s", i.config.BasePath)
	// log.Printf("安装锁文件路径: %s", lockFile)

	_, err := os.Stat(lockFile)
	isInstalled := err == nil

	if isInstalled {
		// 读取锁文件内容
		_, err := os.ReadFile(lockFile)
		if err == nil {
			//log.Printf("安装锁文件内容: %s", string(content))
		}
		//log.Printf("系统已安装")
	} else {
		if os.IsNotExist(err) {
			//log.Printf("系统未安装: 安装锁文件不存在")
		} else {
			//log.Printf("系统未安装: 检查安装锁文件出错: %v", err)
		}
	}

	return isInstalled
}

// 检查环境依赖
func (i *Installer) CheckEnvironment() error {
	// 检查目录权限
	dirs := []string{"config", "logs", "data"}
	for _, dir := range dirs {
		path := filepath.Join(i.config.BasePath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create directory %s failed: %v", dir, err)
		}
	}

	// 检查数据库连接
	sqlDB, err := i.db.DB()
	if err != nil {
		return fmt.Errorf("get database instance failed: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}

	return nil
}

// 确保数据库连接
func (i *Installer) ensureDBConnection() error {
	if i.db == nil {
		// 从配置文件构建DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			i.config.Database.Username,
			i.config.Database.Password,
			i.config.Database.Host,
			i.config.Database.Port,
			i.config.Database.Database,
		)

		// 建立新连接
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("连接数据库失败: %v", err)
		}
		i.db = db
	} else {
		// 测试现有连接
		sqlDB, err := i.db.DB()
		if err != nil {
			return fmt.Errorf("获取数据库实例失败: %v", err)
		}
		if err := sqlDB.Ping(); err != nil {
			// 连接已断开，重新连接
			sqlDB.Close()
			i.db = nil
			return i.ensureDBConnection()
		}
	}
	return nil
}

// 初始化数据库
func (i *Installer) InitDatabase() error {
	// 确保数据库连接
	if err := i.ensureDBConnection(); err != nil {
		return err
	}

	return i.db.Transaction(func(tx *gorm.DB) error {
		// 先删除已存在的表
		if err := i.dropTables(tx); err != nil {
			return fmt.Errorf("删除旧表失败: %v", err)
		}

		// 创建表结构
		if err := i.createTables(tx); err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}

		// 初始化系统设置
		if err := i.initSystemSettings(tx); err != nil {
			return fmt.Errorf("初始化系统设置失败: %v", err)
		}

		return nil
	})
}

// dropTables 删除已存在的表
func (i *Installer) dropTables(tx *gorm.DB) error {
	// 先检查表是否存在
	tables := []interface{}{
		&model.User{},
		&model.SystemSettings{},
		&model.AI_Provider{},
		&model.AI_Model{},
		&model.AI_APIKey{},
		&model.Captcha{},
	}

	for _, table := range tables {
		// 先检查表是否存在
		if exists := tx.Migrator().HasTable(table); exists {
			log.Printf("删除已存在的表: %T", table)
			if err := tx.Migrator().DropTable(table); err != nil {
				return fmt.Errorf("删除表 %T 失败: %v", table, err)
			}
		} else {
			log.Printf("表不存在，跳过删除: %T", table)
		}
	}
	return nil
}

// CreateAdminUser 创建管理员账户
func (i *Installer) CreateAdminUser(username, password, name, email string) error {
	// 添加调试日志
	log.Printf("创建管理员账户: username=%s, name=%s, email=%s", username, name, email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	admin := &model.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Email:        email,
		Role:         "admin",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 添加调试日志
	log.Printf("准备保存管理员信息: %+v", admin)

	if err := i.db.Create(admin).Error; err != nil {
		return fmt.Errorf("创建管理员账户失败: %v", err)
	}

	// 验证是否保存成功
	var savedUser model.User
	if err := i.db.Where("username = ?", username).First(&savedUser).Error; err != nil {
		return fmt.Errorf("验证管理员账户失败: %v", err)
	}

	log.Printf("管理员账户创建成功: %+v", savedUser)
	return nil
}

// CompleteInstallation 完成安装
func (i *Installer) CompleteInstallation() error {
	// 更新系统设置
	if err := i.db.Transaction(func(tx *gorm.DB) error {
		settings := []model.SystemSettings{
			{
				ConfigKey:   model.KeyInstalled,
				Value:       "true",
				Description: model.DefaultSettingsDescription[model.KeyInstalled],
			},
			{
				ConfigKey:   model.KeyInstallTime,
				Value:       time.Now().Format(time.RFC3339),
				Description: model.DefaultSettingsDescription[model.KeyInstallTime],
			},
			{
				ConfigKey:   model.KeyAdminInitialized,
				Value:       "true",
				Description: model.DefaultSettingsDescription[model.KeyAdminInitialized],
			},
		}

		for _, setting := range settings {
			if err := tx.Where("config_key = ?", setting.ConfigKey).
				Assign(setting).
				FirstOrCreate(&setting).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("更新系统设置失败: %v", err)
	}

	// 创建安装锁文件
	lockFile := filepath.Join(i.config.BasePath, "install.lock")
	content := map[string]interface{}{
		"installed_at": time.Now(),
		"version":      "1.0.0",
	}

	lockContent, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(lockFile, lockContent, 0644)
}

// createTables 创建表结构
func (i *Installer) createTables(tx *gorm.DB) error {
	// 添加调试日志
	log.Printf("开始创建数据库表")

	tables := []interface{}{
		&model.User{},
		&model.SystemSettings{},
		&model.AI_Provider{},
		&model.AI_Model{},
		&model.AI_APIKey{},
		&model.Captcha{},
	}

	for _, table := range tables {
		log.Printf("正在创建表: %T", table)
		if err := tx.AutoMigrate(table); err != nil {
			return fmt.Errorf("创建表失败 %T: %v", table, err)
		}

		// 打印表结构信息用于调试
		if columns, err := tx.Migrator().ColumnTypes(table); err == nil {
			for _, column := range columns {
				nullable, _ := column.Nullable()
				log.Printf("列信息: %s, 类型: %s, 是否可空: %s",
					column.Name(), column.DatabaseTypeName(),
					map[bool]string{true: "yes", false: "no"}[nullable])
			}
		}
	}

	log.Printf("数据库表创建完成")
	return nil
}

// 初始化系统设置
func (i *Installer) initSystemSettings(tx *gorm.DB) error {
	// 初始化基本系统设置
	settings := []model.SystemSettings{
		{
			ConfigKey:   model.KeyInstalled,
			Value:       "false",
			Description: model.DefaultSettingsDescription[model.KeyInstalled],
		},
		{
			ConfigKey:   model.KeyInstallTime,
			Value:       time.Now().Format(time.RFC3339),
			Description: model.DefaultSettingsDescription[model.KeyInstallTime],
		},
		{
			ConfigKey:   model.KeyAdminInitialized,
			Value:       "false",
			Description: model.DefaultSettingsDescription[model.KeyAdminInitialized],
		},
	}

	// 使用事务保存所有配置
	for _, setting := range settings {
		if err := tx.Where("config_key = ?", setting.ConfigKey).
			Assign(setting).
			FirstOrCreate(&setting).Error; err != nil {
			return err
		}
	}

	return nil
}

// SystemCheck 表示系统检查结果
type SystemCheck struct {
	Name    string `json:"name"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// SystemChecks 表示所有系统检查的结果
type SystemChecks struct {
	Checks []SystemCheck `json:"checks"`
	Errors []string      `json:"errors"`
}

// CheckSystemEnvironment 检查系统环境
func (i *Installer) CheckSystemEnvironment() SystemChecks {
	var checks SystemChecks
	var errors []string

	// 检查操作系统
	osCheck := SystemCheck{Name: "os",
		Status:  true,
		Message: "操作系统正常",
	}
	if _, err := os.Hostname(); err != nil {
		osCheck.Status = false
		osCheck.Message = "操作系统检查失败"
		errors = append(errors, err.Error())
	}
	checks.Checks = append(checks.Checks, osCheck)

	// 检查CPU
	cpuCheck := SystemCheck{Name: "cpu",
		Status:  runtime.NumCPU() >= 2,
		Message: fmt.Sprintf("CPU核心数: %d", runtime.NumCPU()),
	}
	if !cpuCheck.Status {
		cpuCheck.Message = "CPU核心数不足"
		errors = append(errors, "需要至少2个CPU核心")
	}
	checks.Checks = append(checks.Checks, cpuCheck)

	// 检查内存
	memCheck := SystemCheck{Name: "memory",
		Status:  true,
		Message: "",
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 添加调试日志
	log.Printf("Total memory: %.2f GB", float64(m.Sys)/(1024*1024*1024))
	log.Printf("Memory details: Total = %d, Sys = %d", m.TotalAlloc, m.Sys)

	// 使用 syscall 获取系统总内存
	var si syscall.Sysinfo_t
	if err := syscall.Sysinfo(&si); err != nil {
		memCheck.Status = false
		memCheck.Message = "内存检查失败"
		errors = append(errors, fmt.Sprintf("获取系统内存信息失败: %v", err))
	} else {
		totalRAM := uint64(si.Totalram) * uint64(si.Unit)
		availRAM := uint64(si.Freeram) * uint64(si.Unit)

		// 转换为GB
		totalRAMGB := float64(totalRAM) / (1024 * 1024 * 1024)
		availRAMGB := float64(availRAM) / (1024 * 1024 * 1024)

		log.Printf("System memory: Total = %.2f GB, Available = %.2f GB", totalRAMGB, availRAMGB)

		if availRAM > 1024*1024*1024 { // 1GB
			memCheck.Status = true
			memCheck.Message = fmt.Sprintf("可用内存: %.2f GB (总内存: %.2f GB)", availRAMGB, totalRAMGB)
		} else {
			memCheck.Status = false
			memCheck.Message = fmt.Sprintf("内存不足 (可用: %.2f GB, 总共: %.2f GB)", availRAMGB, totalRAMGB)
			errors = append(errors, fmt.Sprintf("需要至少1GB可用内存，当前可用: %.2f GB", availRAMGB))
		}
	}
	checks.Checks = append(checks.Checks, memCheck)

	// 检查磁盘空间
	diskCheck := SystemCheck{Name: "disk",
		Status:  true,
		Message: "",
	}
	if free, err := getDiskFreeSpace(i.config.BasePath); err == nil {
		if free > 1024*1024*1024 { // 1GB
			diskCheck.Message = fmt.Sprintf("可用空间: %.2f GB", float64(free)/(1024*1024*1024))
		} else {
			diskCheck.Status = false
			diskCheck.Message = "磁盘空间不足"
			errors = append(errors, "需要至少1GB可用磁盘空间")
		}
	} else {
		diskCheck.Status = false
		diskCheck.Message = "磁盘空间检查失败"
		errors = append(errors, err.Error())
	}
	checks.Checks = append(checks.Checks, diskCheck)

	checks.Errors = errors
	return checks
}

// TestDatabaseConnection 测试数据库连接并更新配置
func (i *Installer) TestDatabaseConnection(config config.DatabaseConfig) error {
	// 添加调试日志
	log.Printf("正在测试数据库连接，配置信息: %+v", config)

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// 尝试连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 更新配置文件
	i.config.Database = config

	// 添加调试日志
	log.Printf("数据库连接测试成功，正在保存配置...")

	// 保存配置到文件
	if err := i.SaveConfig(); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	log.Printf("配置保存成功")
	return nil
}

// SaveConfig 保存配置到文件
func (i *Installer) SaveConfig() error {
	// 添加调试日志
	log.Printf("正在保存配置: %+v", i.config)

	configData, err := yaml.Marshal(i.config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 修正配置文件路径 (使用项目根目录)
	configPath := filepath.Join(i.config.BasePath, "config", "config.yaml")
	log.Printf("项目根目录: %s", i.config.BasePath)
	log.Printf("保存配置文件到: %s", configPath)

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 如果文件已存在，先读取现有配置
	var existingConfig config.Config
	if _, err := os.Stat(configPath); err == nil {
		existingData, err := os.ReadFile(configPath)
		if err == nil {
			yaml.Unmarshal(existingData, &existingConfig)
		}
	}

	// 合并配置
	if existingConfig.Server.Host != "" {
		i.config.Server = existingConfig.Server
	}
	if existingConfig.Redis.Host != "" {
		i.config.Redis = existingConfig.Redis
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	log.Printf("配置文件保存成功")
	return nil
}

// TestRedisConnection 测试Redis连接
func (i *Installer) TestRedisConnection(config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}) error {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})
	defer rdb.Close()

	// 设置上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接测试失败: %v", err)
	}

	return nil
}

// getDiskFreeSpace 获取磁盘可用空间
func getDiskFreeSpace(path string) (uint64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}

// SaveDatabaseConfig 保存数据库配置到数据库
func (i *Installer) SaveDatabaseConfig(config config.DatabaseConfig) error {
	// 确保数据库连接
	if err := i.ensureDBConnection(); err != nil {
		return err
	}

	// 使用事务保存所有配置
	return i.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧表（如果存在）
		if tx.Migrator().HasTable(&model.SystemSettings{}) {
			if err := tx.Migrator().DropTable(&model.SystemSettings{}); err != nil {
				return fmt.Errorf("删除旧表失败: %v", err)
			}
		}

		// 创建新表
		if err := tx.Migrator().CreateTable(&model.SystemSettings{}); err != nil {
			return fmt.Errorf("创建系统设置表失败: %v", err)
		}

		// 保存数据库连接状态
		setting := model.SystemSettings{
			ConfigKey:   "database.connected",
			Value:       "true",
			Description: "数据库连接状态",
		}

		if err := tx.Create(&setting).Error; err != nil {
			return err
		}

		return nil
	})
}

// SaveRedisConfig 保存Redis配置到数据库
func (i *Installer) SaveRedisConfig(config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}) error {
	settings := []model.SystemSettings{
		{
			ConfigKey:   model.KeyRedisHost,
			Value:       config.Host,
			Description: model.DefaultSettingsDescription[model.KeyRedisHost],
		},
		{
			ConfigKey:   model.KeyRedisPort,
			Value:       fmt.Sprintf("%d", config.Port),
			Description: model.DefaultSettingsDescription[model.KeyRedisPort],
		},
		{
			ConfigKey:   model.KeyRedisPassword,
			Value:       config.Password,
			Description: model.DefaultSettingsDescription[model.KeyRedisPassword],
		},
		{
			ConfigKey:   model.KeyRedisDB,
			Value:       fmt.Sprintf("%d", config.DB),
			Description: model.DefaultSettingsDescription[model.KeyRedisDB],
		},
	}

	return i.db.Transaction(func(tx *gorm.DB) error {
		for _, setting := range settings {
			if err := tx.Where("config_key = ?", setting.ConfigKey).
				Assign(setting).
				FirstOrCreate(&setting).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
