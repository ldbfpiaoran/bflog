package db

import (
	"bflog/config"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"time"
)

var redisdb *RedisClient

// RedisClient 封装了 go-redis 客户端的操作
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient 创建一个新的 Redis 客户端实例，使用连接池
func NewRedisClient(addr string, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // 没有密码设置为 ""
		DB:       db,       // 使用的数据库编号

	})

	return &RedisClient{
		client: rdb,
		ctx:    context.Background(),
	}
}

// Set 设置一个键值对，并设置过期时间
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		log.Errorf("Failed to set key %s: %v", key, err)
		return err
	}
	log.Infof("Set key %s with value %v", key, value)
	return nil
}

// Get 获取一个键的值
func (r *RedisClient) Get(key string) (string, error) {
	value, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		log.Warnf("Key %s does not exist", key)
		return "", nil
	} else if err != nil {
		log.Errorf("Failed to get key %s: %v", key, err)
		return "", err
	}
	log.Infof("Get key %s with value %v", key, value)
	return value, nil
}

// Del 删除一个键
func (r *RedisClient) Del(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		log.Errorf("Failed to delete key %s: %v", key, err)
		return err
	}
	//log.Infof("Deleted key %s", key)
	return nil
}

// Close 关闭 Redis 客户端连接
func (r *RedisClient) Close() error {
	err := r.client.Close()
	if err != nil {
		log.Errorf("Failed to close Redis connection: %v", err)
		return err
	}
	//log.Info("Closed Redis connection")
	return nil
}

// LPush 将一个或多个值推入列表的左端
func (r *RedisClient) LPush(key string, values ...interface{}) error {
	err := r.client.LPush(r.ctx, key, values...).Err()
	if err != nil {
		log.Errorf("Failed to LPUSH to key %s: %v", key, err)
		return err
	}
	//log.Infof("LPUSH to key %s with values %v", key, values)
	return nil
}

// RPush 将一个或多个值推入列表的右端
func (r *RedisClient) RPush(key string, values ...interface{}) error {
	err := r.client.RPush(r.ctx, key, values...).Err()
	if err != nil {
		log.Errorf("Failed to RPUSH to key %s: %v", key, err)
		return err
	}
	//log.Infof("RPUSH to key %s with values %v", key, values)
	return nil
}

// LRange 获取列表中的指定范围的元素
func (r *RedisClient) LRange(key string, start, stop int64) ([]string, error) {
	values, err := r.client.LRange(r.ctx, key, start, stop).Result()
	if err != nil {
		log.Errorf("Failed to LRANGE key %s: %v", key, err)
		return nil, err
	}
	//log.Infof("LRANGE key %s from %d to %d, got %v", key, start, stop, values)
	return values, nil
}

func (r *RedisClient) LLen(key string) (int64, error) {
	values, err := r.client.LLen(r.ctx, key).Result()
	if err != nil {
		log.Errorf("Failed to LRANGE key %s: %v", key, err)
		return 0, err
	}
	//log.Infof("LRANGE key %s from %d to %d, got %v", key, start, stop, values)
	return values, nil
}

// LPop 从列表的左端弹出一个元素
func (r *RedisClient) LPop(key string) (string, error) {
	value, err := r.client.LPop(r.ctx, key).Result()
	if err == redis.Nil {
		//log.Warnf("Key %s does not exist or list is empty", key)
		return "", nil
	} else if err != nil {
		//log.Errorf("Failed to LPOP from key %s: %v", key, err)
		return "", err
	}
	//log.Infof("LPOP from key %s, got %v", key, value)
	return value, nil
}

// RPop 从列表的右端弹出一个元素
func (r *RedisClient) RPop(key string) (string, error) {
	value, err := r.client.RPop(r.ctx, key).Result()
	if err == redis.Nil {
		log.Warnf("Key %s does not exist or list is empty", key)
		return "", nil
	} else if err != nil {
		log.Errorf("Failed to RPOP from key %s: %v", key, err)
		return "", err
	}
	//log.Infof("RPOP from key %s, got %v", key, value)
	return value, nil
}

func InitRedisDB() {
	redisClient := NewRedisClient("localhost:6379", "", 0)
	redisdb = redisClient

}

func GetRedis() *RedisClient {
	return redisdb
}

// 使用示例
func main() {
	// 创建 Redis 客户端实例
	var password string
	if config.GetBase().Redispass == "123456" {
		password = ""
	} else {
		password = config.GetBase().Redispass
	}
	redisClient := NewRedisClient("localhost:6379", password, 0)

	defer redisClient.Close()

	// 设置一个键值对
	err := redisClient.Set("example_key", "example_value", 10*time.Second)
	if err != nil {
		log.Fatalf("Error setting value: %v", err)
	}

	// 获取一个键的值
	value, err := redisClient.Get("example_key")
	if err != nil {
		log.Fatalf("Error getting value: %v", err)
	}
	fmt.Printf("example_key: %s\n", value)

	// 删除一个键
	err = redisClient.Del("example_key")
	if err != nil {
		log.Fatalf("Error deleting key: %v", err)
	}
}
