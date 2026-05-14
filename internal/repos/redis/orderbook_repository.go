package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisOrderBook struct {
	client *redis.Client
}

func NewRedisOrderBook(client *redis.Client) *RedisOrderBook {
	return &RedisOrderBook{client: client}
}

func (r *RedisOrderBook) AddOrder(ctx context.Context, order *domain.Order) error {
	// Определяем side на основе цены (bid если цена высокая, ask если низкая)
	side := "bids"
	if order.Price < 0.5 {
		side = "asks"
	}
	key := fmt.Sprintf("orderbook:%s:%s:%s", order.MarketID, order.Outcome, side)

	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	if err := r.client.ZAdd(ctx, key, redis.Z{Score: order.Price, Member: orderData}).Err(); err != nil {
		return fmt.Errorf("failed to add order to redis: %w", err)
	}
	return nil
}

func (r *RedisOrderBook) RemoveOrder(ctx context.Context, order *domain.Order) error {
	side := "bids"
	if order.Price < 0.5 {
		side = "asks"
	}
	key := fmt.Sprintf("orderbook:%s:%s:%s", order.MarketID, order.Outcome, side)
	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}
	if err := r.client.ZRem(ctx, key, orderData).Err(); err != nil {
		return fmt.Errorf("failed to remove order from redis: %w", err)
	}
	return nil
}

func (r *RedisOrderBook) GetBestBids(ctx context.Context, marketID, outcome string, limit int64) ([]domain.Order, error) {
	key := fmt.Sprintf("orderbook:%s:%s:bids", marketID, outcome)

	orderData, err := r.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  limit - 1,
		Rev:   true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get best bids from redis: %w", err)
	}

	var orders []domain.Order
	for _, data := range orderData {
		var order domain.Order
		if err := json.Unmarshal([]byte(data), &order); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order data: %w", err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *RedisOrderBook) GetBestAsks(ctx context.Context, marketID, outcome string, limit int64) ([]domain.Order, error) {
	key := fmt.Sprintf("orderbook:%s:%s:asks", marketID, outcome)

	orderData, err := r.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  limit - 1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get best asks from redis: %w", err)
	}

	var orders []domain.Order
	for _, data := range orderData {
		var order domain.Order
		if err := json.Unmarshal([]byte(data), &order); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order data: %w", err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}
