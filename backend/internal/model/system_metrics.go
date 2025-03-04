package model

import (
	"time"
)

// SystemMetrics 系统指标模型
type SystemMetrics struct {
	ID           int64     `gorm:"primarykey"`
	CPUUsage     float64   `gorm:"type:decimal(5,2);comment:CPU使用率"`
	MemoryUsage  float64   `gorm:"type:decimal(5,2);comment:内存使用率"`
	APIHealth    float64   `gorm:"type:decimal(5,2);comment:API健康度"`
	RequestCount int64     `gorm:"type:bigint;comment:请求数"`
	ErrorCount   int64     `gorm:"type:bigint;comment:错误数"`
	CreatedAt    time.Time `gorm:"type:datetime(3);index;comment:创建时间"`
}

// TableName 指定表名
func (SystemMetrics) TableName() string {
	return "system_metrics"
} 