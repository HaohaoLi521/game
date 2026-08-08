# M4 Multiplayer Implementation Plan

> **For agentic workers:** Use the inline execution workflow and commit after each completed task.

**Goal:** Deliver the first usable multiplayer room flow with WebSocket events, Redis-backed room state, JWT identity, and basic reconnect support.

**Architecture:** Keep the existing `internal/handler`, `internal/service`, `internal/model`, and `internal/entity` layout. Handlers own HTTP/WebSocket transport, services own state-machine rules, and the model repository owns Redis reads/writes. WebSocket connections stay process-local; room state and player membership stay in Redis.

**Tech Stack:** Gin, Gorilla WebSocket, go-redis/v9, existing JWT AuthManager, existing GameService.

---

### Task 1: Add room entities and Redis repository

**Files:**
- Create: `backend/internal/entity/room.go`
- Create: `backend/internal/model/room_repository.go`
- Modify: `backend/go.mod` and `backend/go.sum`

- [ ] Add room status constants (`waiting`, `ready`, `playing`, `finished`), room/player records, and protocol event DTOs in `entity/room.go`.
- [ ] Add `RoomRepository` with `Create`, `Get`, `Join`, `SetReady`, `Start`, and `Save` methods. Store one JSON room record at `this-is-pun:room:<roomID>` with a two-hour TTL. Use `WithContext` on every Redis call and reject duplicate players and rooms over eight members.
- [ ] Add `github.com/gorilla/websocket` and run `go mod tidy`.
- [ ] Run `go test ./...` from `backend`.
- [ ] Commit `feat: add redis room repository`.

### Task 2: Add room application service

**Files:**
- Create: `backend/internal/service/room_service.go`

- [ ] Define a narrow repository interface in the service package so the service does not depend on a concrete Redis client.
- [ ] Implement `Create`, `Join`, `Ready`, `Start`, and `Answer` with room-state validation. Only the host may start; every member must be ready; answers are accepted only in `playing`.
- [ ] Use the existing GameService to validate the current puzzle answer and return a serializable answer result. Map expected errors to existing service sentinels.
- [ ] Run `go test ./...`.
- [ ] Commit `feat: add multiplayer room service`.

### Task 3: Add WebSocket handler and HTTP routes

**Files:**
- Create: `backend/internal/handler/room.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] Add authenticated `POST /api/v1/rooms` and `POST /api/v1/rooms/:id/join` endpoints with request/response DTOs.
- [ ] Add authenticated `GET /api/v1/rooms/:id/ws` WebSocket upgrade. Validate JWT before upgrading, register the local connection, and remove it on close.
- [ ] Decode `{type,payload}` messages, pass them to the service with `c.Request.Context()`, and broadcast `{type,room,payload}` events to room connections. Do not pass `gin.Context` into service code.
- [ ] Send `ping`/`pong`, write deadlines, and close idle connections safely.
- [ ] Register the concrete Redis repository, room service, and handler only when Redis is configured; otherwise return 503 from room HTTP/WebSocket routes.
- [ ] Run `go test ./...`.
- [ ] Commit `feat: add multiplayer room websocket api`.

### Task 4: Add player room UI

**Files:**
- Create: `frontend/src/api/rooms.ts`
- Create: `frontend/src/stores/room.ts`
- Create: `frontend/src/views/RoomLobbyView.vue`
- Create: `frontend/src/views/RoomView.vue`
- Modify: `frontend/src/router/index.ts` and `frontend/src/views/HomeView.vue`

- [ ] Add room create/join API calls and a WebSocket client that sends the player JWT during the handshake.
- [ ] Implement lobby controls for room ID, join, ready, start, and visible player state.
- [ ] Implement room game state, answer submission, result display, and reconnect button.
- [ ] Run `npx vue-tsc --noEmit` and `npm run build` from `frontend`.
- [ ] Commit `feat: add multiplayer room ui`.

### Task 5: Verify the M4 baseline

- [ ] Start Redis and PostgreSQL containers when Docker is available.
- [ ] Start the backend with `REDIS_ADDR`, `DATABASE_URL`, and `JWT_SECRET`.
- [ ] Use two authenticated clients to create/join a room, verify ready/start events, submit an answer, and verify the finish event.
- [ ] Stop one client, reconnect within 60 seconds, and verify the same player ID remains in room state.
- [ ] Re-run backend tests and both frontend/admin builds.
- [ ] Commit `chore: verify m4 multiplayer baseline`.
