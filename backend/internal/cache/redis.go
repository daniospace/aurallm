package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Key formats
func teamBudgetLimitKey(teamID string) string { return fmt.Sprintf("team:%s:limit", teamID) }
func teamBudgetUsedKey(teamID string) string  { return fmt.Sprintf("team:%s:used", teamID) }
func gatewayKeyHashKey(keyHash string) string { return fmt.Sprintf("key:%s", keyHash) }

func (r *RedisCache) GetTeamBudget(ctx context.Context, teamID string) (float64, float64, bool, error) {
	limitStr, err := r.client.Get(ctx, teamBudgetLimitKey(teamID)).Result()
	if err == redis.Nil {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}

	usedStr, err := r.client.Get(ctx, teamBudgetUsedKey(teamID)).Result()
	if err != nil && err != redis.Nil {
		return 0, 0, false, err
	}

	var limit, used float64
	fmt.Sscanf(limitStr, "%f", &limit)
	if usedStr != "" {
		fmt.Sscanf(usedStr, "%f", &used)
	}

	return limit, used, true, nil
}

func (r *RedisCache) SetTeamBudget(ctx context.Context, teamID string, limit float64, used float64) error {
	pipe := r.client.Pipeline()
	pipe.Set(ctx, teamBudgetLimitKey(teamID), limit, 1*time.Hour)
	pipe.Set(ctx, teamBudgetUsedKey(teamID), used, 1*time.Hour) // Used is technically persisted in PG, cache acts as fast-path
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisCache) IncrementTeamBudgetUsed(ctx context.Context, teamID string, amount float64) (float64, error) {
	// Redis natively supports atomic float increments!
	newUsed, err := r.client.IncrByFloat(ctx, teamBudgetUsedKey(teamID), amount).Result()
	if err != nil {
		return 0, err
	}
	return newUsed, nil
}

func (r *RedisCache) GetGatewayKey(ctx context.Context, keyHash string) (*model.GatewayKey, bool, error) {
	val, err := r.client.Get(ctx, gatewayKeyHashKey(keyHash)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var key model.GatewayKey
	if err := json.Unmarshal([]byte(val), &key); err != nil {
		return nil, false, err
	}

	return &key, true, nil
}

func (r *RedisCache) SetGatewayKey(ctx context.Context, keyHash string, key *model.GatewayKey) error {
	bytes, err := json.Marshal(key)
	if err != nil {
		return err
	}
	// Cache keys for a reasonable time to prevent stale permissions
	return r.client.Set(ctx, gatewayKeyHashKey(keyHash), bytes, 5*time.Minute).Err()
}

func (r *RedisCache) InvalidateGatewayKey(ctx context.Context, keyHash string) error {
	return r.client.Del(ctx, gatewayKeyHashKey(keyHash)).Err()
}

func (r *RedisCache) InvalidateTeam(ctx context.Context, teamID string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, teamBudgetLimitKey(teamID))
	pipe.Del(ctx, teamBudgetUsedKey(teamID))
	_, err := pipe.Exec(ctx)
	return err
}
