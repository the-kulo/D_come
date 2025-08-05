package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"D_come/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{client: rdb}, nil
}

// SetStockData 存储股票数据到Redis
func (r *RedisClient) SetStockData(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set data to Redis: %w", err)
	}

	return nil
}

// GetStockData 从Redis获取股票数据
func (r *RedisClient) GetStockData(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return fmt.Errorf("failed to get data from Redis: %w", err)
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// DeleteStockData 删除Redis中的股票数据
func (r *RedisClient) DeleteStockData(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete data from Redis: %w", err)
	}
	return nil
}

// SetMultipleStockData 批量存储股票数据
func (r *RedisClient) SetMultipleStockData(ctx context.Context, dataMap map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()

	for key, data := range dataMap {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, jsonData, expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	return nil
}

// KeyExists 检查键是否存在
func (r *RedisClient) KeyExists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return exists > 0, nil
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}