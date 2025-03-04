package install

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// TestRedisConnection 测试Redis连接
func TestRedisConnection(host string, port int, password string, db int) error {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	// 设置上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	// 关闭连接
	if err := rdb.Close(); err != nil {
		return fmt.Errorf("关闭Redis连接失败: %w", err)
	}

	return nil
}
