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

- [x] Add room status constants (`waiting`, `ready`, `playing`, `finished`), room/player records, and protocol event DTOs in `entity/room.go`.
- [x] Add `RoomRepository` with Redis JSON state, a two-hour TTL, WATCH retries, and duplicate/limit validation through the service.
- [x] Add `github.com/gorilla/websocket` and run `go mod tidy`.
- [x] Run `go test ./...` from `backend`.
- [x] Commit `feat: add redis room repository`.

### Task 2: Add room application service

**Files:**
- Create: `backend/internal/service/room_service.go`

- [x] Define a narrow repository interface in the service package so the service does not depend on a concrete Redis client.
- [x] Implement `Create`, `Join`, `Ready`, `Start`, `Answer`, and disconnect state transitions with room-state validation.
- [x] Use the existing GameService to validate the current puzzle answer and return a serializable answer result.
- [x] Run `go test ./...`.
- [x] Commit `feat: add multiplayer room service`.

### Task 3: Add WebSocket handler and HTTP routes

**Files:**
- Create: `backend/internal/handler/room.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [x] Add authenticated `POST /api/v1/rooms` and `POST /api/v1/rooms/:id/join` endpoints with request/response DTOs.
- [x] Add authenticated `GET /api/v1/rooms/:id/ws` WebSocket upgrade with JWT validation and connection lifecycle management.
- [x] Decode `{type,payload}` messages, pass them to the service with `c.Request.Context()`, and broadcast room events.
- [x] Send `ping`/`pong`, refresh read deadlines, and close idle connections safely.
- [x] Register the Redis repository independently and return 503 for missing room dependencies.
- [x] Run `go test ./...`.
- [x] Commit `feat: add multiplayer room websocket api`.

### Task 4: Add player room UI

**Files:**
- Create: `frontend/src/api/rooms.ts`
- Create: `frontend/src/stores/room.ts`
- Create: `frontend/src/views/RoomLobbyView.vue`
- Create: `frontend/src/views/RoomView.vue`
- Modify: `frontend/src/router/index.ts` and `frontend/src/views/HomeView.vue`

- [x] Add room create/join API calls and a WebSocket client that sends the player JWT during the handshake query.
- [x] Implement lobby controls for room ID, join, ready, start, and visible player state.
- [x] Implement room game state, answer submission, result display, and heartbeat reconnect behavior.
- [x] Run `npx vue-tsc --noEmit` and `npm run build` from `frontend`.
- [x] Commit `feat: add multiplayer room ui`.

### Task 5: Verify the M4 baseline

- [x] Start Redis, PostgreSQL, and MinIO containers.
- [x] Start the backend with `REDIS_ADDR`, `DATABASE_URL`, and `JWT_SECRET`.
- [x] Use two authenticated clients to create/join a room, verify ready/start events, submit answers, and verify both finish events.
- [x] Stop one client, reconnect, and verify the same player remains present in room state.
- [x] Re-run backend tests and both frontend/admin builds. Race tests remain unavailable because the machine has no GCC/CGO compiler.
- [ ] Commit a separate verification-only commit; behavioral implementation commits already exist.
