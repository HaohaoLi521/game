package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"this-is-pun/backend/internal/entity"
	"this-is-pun/backend/internal/model"
	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

var roomUpgrader = websocket.Upgrader{ReadBufferSize: 4096, WriteBufferSize: 4096, CheckOrigin: func(*http.Request) bool { return true }}

// RoomHandler 是联机房间的 HTTP 与 WebSocket 入口。
type RoomHandler struct {
	service *service.RoomService
	mu      sync.RWMutex
	rooms   map[string]map[string]*roomConnection
}

type roomConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func NewRoomHandler(service *service.RoomService) *RoomHandler {
	return &RoomHandler{service: service, rooms: make(map[string]map[string]*roomConnection)}
}

// CreateRequest 是创建房间请求 DTO。
type CreateRoomRequest struct {
	PuzzleSetID int64  `json:"puzzle_set_id"`
	PlayerName  string `json:"player_name"`
}

// JoinRequest 是加入房间请求 DTO。
type JoinRoomRequest struct {
	PlayerName string `json:"player_name"`
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	room, err := h.service.Create(c.Request.Context(), service.CreateRoomInput{PlayerID: c.GetString("UserID"), PlayerName: req.PlayerName, PuzzleSetID: req.PuzzleSetID})
	if err != nil {
		handleError(c, err)
		return
	}
	response.Created(c, room)
}

func (h *RoomHandler) Join(c *gin.Context) {
	var req JoinRoomRequest
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}
	room, err := h.service.Join(c.Request.Context(), service.JoinRoomInput{RoomID: c.Param("id"), PlayerID: c.GetString("UserID"), PlayerName: req.PlayerName})
	if err != nil {
		handleError(c, err)
		return
	}
	h.broadcast(room.ID, entity.RoomEvent{Type: "player_joined", Room: &room})
	response.OK(c, room)
}

// WebSocket 升级后只保留连接对象，业务状态仍由 RoomService/Redis 管理。
func (h *RoomHandler) WebSocket(c *gin.Context) {
	roomID := c.Param("id")
	playerID := c.GetString("UserID")
	if roomID == "" || playerID == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	room, err := h.service.Join(c.Request.Context(), service.JoinRoomInput{RoomID: roomID, PlayerID: playerID, PlayerName: playerID})
	if err != nil {
		handleError(c, err)
		return
	}
	conn, err := roomUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &roomConnection{conn: conn}
	h.addConnection(roomID, playerID, client)
	defer func() {
		h.removeConnection(roomID, playerID, client)
		_, _ = h.service.Leave(c.Request.Context(), roomID, playerID)
		_ = conn.Close()
	}()
	_ = client.writeJSON(entity.RoomEvent{Type: "room_state", Room: &room})
	conn.SetReadLimit(16 << 10)
	_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(70 * time.Second)) })
	for {
		_, data, readErr := conn.ReadMessage()
		if readErr != nil {
			return
		}
		var msg roomMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			_ = client.writeJSON(entity.RoomEvent{Type: "error", Payload: "invalid message"})
			continue
		}
		h.handleMessage(c, client, roomID, playerID, msg)
	}
}

type roomMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (h *RoomHandler) handleMessage(c *gin.Context, client *roomConnection, roomID, playerID string, msg roomMessage) {
	switch msg.Type {
	case "ping":
		_ = client.writeJSON(entity.RoomEvent{Type: "pong"})
	case "ready":
		room, err := h.service.Ready(c.Request.Context(), service.ReadyRoomInput{RoomID: roomID, PlayerID: playerID})
		if err != nil {
			_ = client.writeJSON(errorEvent(err))
			return
		}
		h.broadcast(roomID, entity.RoomEvent{Type: "player_ready", Room: &room})
	case "start":
		room, err := h.service.Start(c.Request.Context(), service.StartRoomInput{RoomID: roomID, PlayerID: playerID})
		if err != nil {
			_ = client.writeJSON(errorEvent(err))
			return
		}
		h.broadcast(roomID, entity.RoomEvent{Type: "game_started", Room: &room})
	case "answer":
		var payload struct {
			Answer     string `json:"answer"`
			AnswerMode string `json:"answer_mode"`
			ElapsedMS  int64  `json:"elapsed_ms"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			_ = client.writeJSON(errorEvent(service.ErrInvalidRequest))
			return
		}
		room, result, err := h.service.Answer(c.Request.Context(), service.AnswerRoomInput{RoomID: roomID, PlayerID: playerID, Answer: payload.Answer, AnswerMode: model.AnswerMode(payload.AnswerMode), ElapsedMS: payload.ElapsedMS})
		if err != nil {
			_ = client.writeJSON(errorEvent(err))
			return
		}
		h.broadcast(roomID, entity.RoomEvent{Type: "answer_result", Room: &room, Payload: result})
		if room.Status == entity.RoomStatusFinished {
			h.broadcast(roomID, entity.RoomEvent{Type: "game_finished", Room: &room})
		}
	default:
		_ = client.writeJSON(errorEvent(service.ErrInvalidRequest))
	}
}

func (h *RoomHandler) addConnection(roomID, playerID string, connection *roomConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[string]*roomConnection)
	}
	if previous := h.rooms[roomID][playerID]; previous != nil {
		_ = previous.conn.Close()
	}
	h.rooms[roomID][playerID] = connection
}

func (h *RoomHandler) removeConnection(roomID, playerID string, connection *roomConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if current := h.rooms[roomID][playerID]; current == connection {
		delete(h.rooms[roomID], playerID)
	}
	if len(h.rooms[roomID]) == 0 {
		delete(h.rooms, roomID)
	}
}

func (h *RoomHandler) broadcast(roomID string, event entity.RoomEvent) {
	h.mu.RLock()
	connections := make([]*roomConnection, 0, len(h.rooms[roomID]))
	for _, connection := range h.rooms[roomID] {
		connections = append(connections, connection)
	}
	h.mu.RUnlock()
	for _, connection := range connections {
		_ = connection.writeJSON(event)
	}
}

func (c *roomConnection) writeJSON(value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(value)
}

func errorEvent(err error) entity.RoomEvent {
	return entity.RoomEvent{Type: "error", Payload: err.Error()}
}
