package priorityqueue

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// RedisPriorityQueue implements PriorityQueuer using Redis
type RedisPriorityQueue struct {
	client *redis.Client
	ctx    context.Context
	mutex  sync.Mutex
}

// NewRedisPriorityQueue creates a new Redis-based priority queue
func NewRedisPriorityQueue(addr, password string, db int) PriorityQueuer {
	rpq := &RedisPriorityQueue{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ctx: context.Background(),
	}
	// Verify connection
	if err := rpq.client.Ping(rpq.ctx).Err(); err != nil {
		panic(fmt.Sprintf("failed to connect to Redis at %s: %v", addr, err))
	}
	return rpq
}

// ClearQueues removes specified queues from Redis
func (rpq *RedisPriorityQueue) ClearQueues(queues ...string) error {
	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	if len(queues) == 0 {
		return nil
	}
	_, err := rpq.client.Del(rpq.ctx, queues...).Result()
	if err != nil {
		return fmt.Errorf("redis error clearing queues: %v", err)
	}
	return nil
}

func (rpq *RedisPriorityQueue) AddQueue(name string) error {
	return nil
}

func (rpq *RedisPriorityQueue) Enqueue(queueName string, value interface{}, priority int) error {
	if priority < 0 || priority > 9 {
		return fmt.Errorf("priority must be between 0 and 9")
	}

	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	err := rpq.client.ZAdd(rpq.ctx, queueName, redis.Z{
		Score:  float64(priority),
		Member: fmt.Sprintf("%v", value),
	}).Err()
	return err
}

func (rpq *RedisPriorityQueue) Dequeue(queueName string) (interface{}, error) {
	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	result, err := rpq.client.ZPopMin(rpq.ctx, queueName, 1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %v", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("queue '%s' is empty", queueName)
	}
	return result[0].Member, nil
}

func (rpq *RedisPriorityQueue) IsEmpty(queueName string) (bool, error) {
	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	count, err := rpq.client.ZCard(rpq.ctx, queueName).Result()
	if err != nil {
		return false, fmt.Errorf("redis error: %v", err)
	}
	return count == 0, nil
}

func (rpq *RedisPriorityQueue) ListContents(queueName string) (map[int][]interface{}, error) {
	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	members, err := rpq.client.ZRangeWithScores(rpq.ctx, queueName, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %v", err)
	}

	contents := make(map[int][]interface{})
	for _, member := range members {
		priority := int(member.Score + 0.5) // Round to handle micro-decrements
		if priority >= 0 && priority <= 9 {
			contents[priority] = append(contents[priority], member.Member)
		}
	}
	return contents, nil
}

func (rpq *RedisPriorityQueue) GetPosition(queueName string, value interface{}) (int, int, error) {
	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	members, err := rpq.client.ZRangeWithScores(rpq.ctx, queueName, 0, -1).Result()
	if err != nil {
		return -1, -1, fmt.Errorf("redis error: %v", err)
	}

	valueStr := fmt.Sprintf("%v", value)
	for i, member := range members {
		if member.Member == valueStr {
			priority := int(member.Score + 0.5) // Round to nearest integer
			pos := 0
			for j := 0; j < i; j++ {
				if int(members[j].Score+0.5) == priority {
					pos++
				}
			}
			return priority, pos, nil
		}
	}
	return -1, -1, fmt.Errorf("value '%v' not found in queue '%s'", value, queueName)
}

func (rpq *RedisPriorityQueue) InsertAtTop(queueName string, value interface{}, priority int) error {
	if priority < 0 || priority > 9 {
		return fmt.Errorf("priority must be between 0 and 9")
	}

	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	valueStr := fmt.Sprintf("%v", value)
	rpq.client.ZRem(rpq.ctx, queueName, valueStr)

	score := float64(priority) - 0.000001
	return rpq.client.ZAdd(rpq.ctx, queueName, redis.Z{
		Score:  score,
		Member: valueStr,
	}).Err()
}

func (rpq *RedisPriorityQueue) DeleteItem(queueName string, value interface{}) error {
	rpq.mutex.Lock()
	defer rpq.mutex.Unlock()

	valueStr := fmt.Sprintf("%v", value)
	count, err := rpq.client.ZRem(rpq.ctx, queueName, valueStr).Result()
	if err != nil {
		return fmt.Errorf("redis error: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("value '%v' not found in queue '%s'", value, queueName)
	}
	return nil
}
