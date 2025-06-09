package redisutil

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}
}

func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(context.Background(), key, jsonData, expiration).Err()
}

func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisClient) HSet(key, field string, value interface{}) error {
	return r.client.HSet(context.Background(), key, field, value).Err()
}

func (r *RedisClient) HIncrBy(key, field string, incr int64) (int64, error) {
	return r.client.HIncrBy(context.Background(), key, field, incr).Result()
}

func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(context.Background(), key).Result()
}

func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	return r.client.Expire(context.Background(), key, expiration).Err()
}
