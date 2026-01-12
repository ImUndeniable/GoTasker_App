package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gotasker/internal/models"

	"github.com/redis/go-redis/v9"
)

const taskksETL = 60 * time.Second

// Key - Value pair
func TasksCacheKey(userID int64) string {
	return fmt.Sprintf("tasks:user:%d", userID)
}

//Checking Redis first

func GetTasks(ctx context.Context, rdb *redis.Client, userID int64) ([]models.Task, bool, error) {
	if rdb == nil {
		return nil, false, nil
	}

	val, err := rdb.Get(ctx, TasksCacheKey(userID)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	var tasks []models.Task
	if err := json.Unmarshal([]byte(val), &tasks); err != nil {
		return nil, false, err
	}

	return tasks, true, nil
}

// Setting up new Redis

func SetTasks(ctx context.Context, rdb *redis.Client, userID int64, tasks []models.Task) error {
	if rdb == nil {
		return nil
	}

	b, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, TasksCacheKey(userID), b, taskksETL).Err()
}

// Delete Redis

func DeletTaks(ctx context.Context, rdb *redis.Client, userID int64) error {
	if rdb == nil {
		return nil
	}

	err := rdb.Del(ctx, TasksCacheKey(userID)).Err()
	if err != nil {
		return nil
	}

	return err
}
