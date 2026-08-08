package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"this-is-pun/backend/internal/entity"
)

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrRoomExists      = errors.New("room already exists")
	ErrRoomUnavailable = errors.New("room store unavailable")
)

// RoomRepository 负责 Redis 房间状态读写和并发更新。
type RoomRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRoomRepository(client *redis.Client) *RoomRepository {
	return &RoomRepository{client: client, ttl: 2 * time.Hour}
}

func (r *RoomRepository) Create(ctx context.Context, room entity.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("marshal room: %w", err)
	}
	ok, err := r.client.SetNX(ctx, roomKey(room.ID), data, r.ttl).Result()
	if err != nil {
		return fmt.Errorf("%w: create room: %w", ErrRoomUnavailable, err)
	}
	if !ok {
		return ErrRoomExists
	}
	return nil
}

func (r *RoomRepository) Get(ctx context.Context, roomID string) (entity.Room, error) {
	data, err := r.client.Get(ctx, roomKey(roomID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return entity.Room{}, ErrRoomNotFound
	}
	if err != nil {
		return entity.Room{}, fmt.Errorf("%w: get room: %w", ErrRoomUnavailable, err)
	}
	var room entity.Room
	if err := json.Unmarshal(data, &room); err != nil {
		return entity.Room{}, fmt.Errorf("unmarshal room: %w", err)
	}
	return room, nil
}

// Update 使用 Redis WATCH 防止并发加入、准备和答题覆盖彼此状态。
func (r *RoomRepository) Update(ctx context.Context, roomID string, mutate func(*entity.Room) error) (entity.Room, error) {
	var updated entity.Room
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = r.client.Watch(ctx, func(tx *redis.Tx) error {
			data, err := tx.Get(ctx, roomKey(roomID)).Bytes()
			if errors.Is(err, redis.Nil) {
				return ErrRoomNotFound
			}
			if err != nil {
				return fmt.Errorf("%w: watch room: %w", ErrRoomUnavailable, err)
			}
			if err := json.Unmarshal(data, &updated); err != nil {
				return fmt.Errorf("unmarshal room: %w", err)
			}
			if err := mutate(&updated); err != nil {
				return err
			}
			updated.UpdatedAt = time.Now().UTC()
			encoded, err := json.Marshal(updated)
			if err != nil {
				return fmt.Errorf("marshal updated room: %w", err)
			}
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, roomKey(roomID), encoded, r.ttl)
				return nil
			})
			return err
		}, roomKey(roomID))
		if !errors.Is(err, redis.TxFailedErr) {
			break
		}
	}
	if err != nil {
		return entity.Room{}, err
	}
	return updated, nil
}

func roomKey(id string) string { return "this-is-pun:room:" + id }
