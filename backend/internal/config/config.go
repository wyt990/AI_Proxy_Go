package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// 全局常量
const (
	SiteName = "五月天AI代理" // 网站标题
)

// Config 系统配置结构体
type Config struct {
	BasePath string         `yaml:"-"` // 不序列化到配置文件
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// 默认配置
var defaultConfig = Config{
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8341,
	},
	Database: DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "",
		Database: "AI_Proxy_Go",
	},
	Redis: RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	},
}

func LoadConfig(path string) (*Config, error) {
	// 添加调试日志
	log.Printf("加载配置文件: %s", path)

	cfg := defaultConfig

	// 获取程序根目录
	rootDir, err := os.Getwd() // 直接使用当前工作目录作为项目根目录
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %v", err)
	}
	cfg.BasePath = rootDir

	log.Printf("项目根目录: %s", rootDir)

	// 如果配置文件存在，则加载配置文件
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}

		// 添加调试日志
		log.Printf("配置文件内容: %s", string(data))

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %v", err)
		}

		// 确保 BasePath 不被配置文件覆盖
		cfg.BasePath = rootDir

		// 添加调试日志
		log.Printf("加载的配置: %+v", cfg)
	} else {
		log.Printf("配置文件不存在，使用默认配置")
	}

	return &cfg, nil
}
