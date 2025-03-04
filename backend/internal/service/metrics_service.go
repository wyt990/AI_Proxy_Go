package service

import (
	"AI_Proxy_Go/backend/internal/model"
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"gorm.io/gorm"
)

type MetricsService struct {
	DB *gorm.DB
}

func NewMetricsService(db *gorm.DB) *MetricsService {
	return &MetricsService{DB: db}
}

// CollectMetrics 收集系统指标
func (s *MetricsService) CollectMetrics() error {
	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	// 计算API健康度 (这里用一个简单的计算方式)
	var errorCount int64
	if err := s.DB.Model(&model.MessageStats{}).
		Where("response_time > 5000 OR response_time = 0"). // 响应时间超过5秒或为0视为错误
		Where("created_at >= ?", time.Now().Add(-5*time.Minute)).
		Count(&errorCount).Error; err != nil {
		return err
	}

	var totalCount int64
	if err := s.DB.Model(&model.MessageStats{}).
		Where("created_at >= ?", time.Now().Add(-5*time.Minute)).
		Count(&totalCount).Error; err != nil {
		return err
	}

	apiHealth := 100.0
	if totalCount > 0 {
		apiHealth = float64(totalCount-errorCount) / float64(totalCount) * 100
	}

	// 保存指标
	metrics := &model.SystemMetrics{
		CPUUsage:     cpuPercent[0],
		MemoryUsage:  memInfo.UsedPercent,
		APIHealth:    apiHealth,
		RequestCount: totalCount,
		ErrorCount:   errorCount,
		CreatedAt:    time.Now(),
	}

	return s.DB.Create(metrics).Error
}

// StartMetricsCollection 启动定时采集
func (s *MetricsService) StartMetricsCollection(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			if err := s.CollectMetrics(); err != nil {
				log.Printf("采集系统指标失败: %v", err)
			}
		}
	}()
}

// GetLatestMetrics 获取最新的系统指标
func (s *MetricsService) GetLatestMetrics() (*model.SystemMetrics, error) {
	var metrics model.SystemMetrics
	err := s.DB.Order("created_at DESC").First(&metrics).Error
	return &metrics, err
}

// CleanOldMetrics 清理旧数据
func (s *MetricsService) CleanOldMetrics(retention time.Duration) error {
	return s.DB.Where("created_at < ?", time.Now().Add(-retention)).
		Delete(&model.SystemMetrics{}).Error
}
