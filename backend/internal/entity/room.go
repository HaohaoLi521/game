package entity

import "time"

// RoomStatus 是房间生命周期状态。
type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusReady    RoomStatus = "ready"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

// RoomPlayer 是房间内玩家状态。
type RoomPlayer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ready    bool   `json:"ready"`
	Score    int    `json:"score"`
	Answered bool   `json:"answered"`
	Present  bool   `json:"present"`
}

// Room 是 Redis 持久化的房间状态。
type Room struct {
	ID            string       `json:"id"`
	HostID        string       `json:"host_id"`
	Status        RoomStatus   `json:"status"`
	Players       []RoomPlayer `json:"players"`
	PuzzleIDs     []int64      `json:"puzzle_ids"`
	CurrentPuzzle int64        `json:"current_puzzle"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// RoomEvent 是 WebSocket 广播事件。
type RoomEvent struct {
	Type    string `json:"type"`
	Room    *Room  `json:"room,omitempty"`
	Payload any    `json:"payload,omitempty"`
}

// RoomMessage 是客户端发送的 WebSocket 消息。
type RoomMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}
