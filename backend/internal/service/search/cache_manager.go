package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	m "AI_Proxy_Go/backend/internal/model"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// CacheManager 缓存管理器
type CacheManager struct {
	redis     *redis.Client
	logger    *log.Logger
	stats     atomic.Value // 使用 atomic.Value 替代直接指针
	closeOnce sync.Once
	stopChan  chan struct{}
	db        *gorm.DB // 添加数据库访问
}

// CacheStats 缓存统计信息
type CacheStats struct {
	mu            sync.RWMutex
	TotalItems    int64     `json:"total_items"`
	TotalSize     int64     `json:"total_size"`
	HitCount      int64     `json:"hit_count"`
	MissCount     int64     `json:"miss_count"`
	LastCleanTime time.Time `json:"last_clean_time"`
}

// Copy 创建 CacheStats 的深拷贝
func (s *CacheStats) Copy() CacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return CacheStats{
		TotalItems:    s.TotalItems,
		TotalSize:     s.TotalSize,
		HitCount:      s.HitCount,
		MissCount:     s.MissCount,
		LastCleanTime: s.LastCleanTime,
	}
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(redis *redis.Client, logger *log.Logger, db *gorm.DB) *CacheManager {
	cm := &CacheManager{
		redis:    redis,
		logger:   logger,
		stopChan: make(chan struct{}),
		db:       db,
	}

	// 初始化 stats
	cm.stats.Store(&CacheStats{})

	// 启动清理任务，使用配置的清理间隔
	var cleanupSetting m.SystemSettings
	cleanupInterval := time.Hour // 默认每小时清理一次
	if err := db.Where("config_key = ?", m.KeySearchCacheCleanupInterval).First(&cleanupSetting).Error; err == nil {
		if interval, err := strconv.Atoi(cleanupSetting.Value); err == nil && interval > 0 {
			cleanupInterval = time.Duration(interval) * time.Second
		}
	}

	go cm.startCleanupTask(cleanupInterval)
	return cm
}

// Get 从缓存获取数据
func (cm *CacheManager) Get(ctx context.Context, key string) (*m.SearchResult, error) {
	if cm.redis == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err := cm.updateStats(func(s *CacheStats) {
			s.mu.Lock()
			s.MissCount++
			s.mu.Unlock()
		}); err != nil {
			cm.logger.Printf("更新缓存统计失败: %v", err)
		}
		return nil, err
	}

	var result m.SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if err := cm.updateStats(func(s *CacheStats) {
		s.mu.Lock()
		s.HitCount++
		s.mu.Unlock()
	}); err != nil {
		cm.logger.Printf("更新缓存统计失败: %v", err)
	}

	return &result, nil
}

// Set 将数据存入缓存
func (cm *CacheManager) Set(ctx context.Context, key string, value *m.SearchResult, expiration time.Duration) error {
	if cm.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// 检查缓存大小限制
	if exceeded, err := cm.checkSizeLimit(ctx); err != nil {
		cm.logger.Printf("检查缓存大小失败: %v", err)
	} else if exceeded {
		if err := cm.cleanup(ctx); err != nil {
			cm.logger.Printf("清理缓存失败: %v", err)
		}
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := cm.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return err
	}

	cm.updateStats(func(s *CacheStats) {
		s.mu.Lock()
		s.TotalItems++
		s.TotalSize += int64(len(data))
		s.mu.Unlock()
	})

	return nil
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() (CacheStats, error) {
	value := cm.stats.Load()
	if value == nil {
		return CacheStats{}, fmt.Errorf("stats not initialized")
	}

	stats, ok := value.(*CacheStats)
	if !ok {
		return CacheStats{}, fmt.Errorf("invalid stats type")
	}

	return stats.Copy(), nil
}

// updateStats 更新缓存统计信息
func (cm *CacheManager) updateStats(update func(*CacheStats)) error {
	for i := 0; i < 3; i++ { // 最多重试3次
		value := cm.stats.Load()
		if value == nil {
			return fmt.Errorf("stats not initialized")
		}

		oldStats, ok := value.(*CacheStats)
		if !ok {
			return fmt.Errorf("invalid stats type")
		}

		// 创建新的统计信息实例
		newStats := new(CacheStats)
		*newStats = oldStats.Copy() // 使用深拷贝

		// 在新实例上执行更新
		update(newStats)

		// 尝试原子替换
		if cm.stats.CompareAndSwap(value, newStats) {
			return nil
		}
	}
	return fmt.Errorf("failed to update stats after retries")
}

// startCleanupTask 启动定期清理任务
func (cm *CacheManager) startCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := cm.cleanup(ctx); err != nil {
				cm.logger.Printf("定期清理缓存失败: %v", err)
			}
		case <-cm.stopChan:
			return
		}
	}
}

// cleanup 清理过期和超量的缓存
func (cm *CacheManager) cleanup(ctx context.Context) error {
	// 获取所有缓存键
	keys, err := cm.redis.Keys(ctx, "search:*").Result()
	if err != nil {
		return fmt.Errorf("获取缓存键失败: %v", err)
	}

	// 清理过期的缓存
	for _, key := range keys {
		ttl, err := cm.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		if ttl < 0 {
			cm.redis.Del(ctx, key)
		}
	}

	cm.updateStats(func(s *CacheStats) {
		s.mu.Lock()
		s.LastCleanTime = time.Now()
		s.mu.Unlock()
	})

	return nil
}

// checkSizeLimit 检查缓存大小是否超过限制
func (cm *CacheManager) checkSizeLimit(ctx context.Context) (bool, error) {
	// 从数据库获取最大缓存大小配置
	var maxCacheSetting m.SystemSettings
	maxSize := int64(10000) // 默认最大缓存条目数

	if err := cm.db.Where("config_key = ?", m.KeySearchMaxCacheSize).First(&maxCacheSetting).Error; err == nil {
		if size, err := strconv.ParseInt(maxCacheSetting.Value, 10, 64); err == nil && size > 0 {
			maxSize = size
		}
	}

	if err := cm.redis.DBSize(ctx).Err(); err != nil {
		return false, err
	}

	value := cm.stats.Load()
	if value == nil {
		return false, fmt.Errorf("stats not initialized")
	}

	stats, ok := value.(*CacheStats)
	if !ok {
		return false, fmt.Errorf("invalid stats type")
	}

	// 获取当前统计信息的快照
	currentStats := stats.Copy()
	return currentStats.TotalItems >= maxSize, nil
}

// Close 关闭缓存管理器
func (cm *CacheManager) Close() {
	cm.closeOnce.Do(func() {
		close(cm.stopChan)
	})
}
