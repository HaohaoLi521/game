# M4 联机房间设计

## 范围

首版实现创建房间、加入房间、准备、开始、答题、结算和基础断线重连。房间最多 8 名玩家，房间状态保留 2 小时，断线玩家保留 60 秒席位。

## 三层结构

- `internal/handler/room.go`：HTTP 创建/加入接口和 WebSocket 握手、消息编解码、连接生命周期。
- `internal/service/room_service.go`：房间状态机、玩家权限、答题规则和事件编排；不接触 Gin、Redis 客户端和 WebSocket 连接。
- `internal/model/room_repository.go`：Redis 房间状态读写、TTL、成员更新和并发保护。
- `internal/entity/room.go`：房间、玩家、事件 DTO 与状态常量。

## 状态和协议

房间状态为 `waiting`、`ready`、`playing`、`finished`。客户端消息统一为 `{type, payload}`，服务端事件统一为 `{type, room, payload}`。

- 客户端：`ready`、`start`、`answer`、`ping`
- 服务端：`room_state`、`player_joined`、`player_ready`、`game_started`、`answer_result`、`game_finished`、`error`、`pong`

服务端只允许房主执行 `start`；只有 `playing` 状态允许答题；所有状态变更先写 Redis，再广播事件。答题复用现有 GameService 的题目校验，房间状态只保存进度和结算摘要。

## 失败和安全

JWT 在 WebSocket 握手阶段验证，玩家 ID 作为 Redis 成员身份。房间不存在、已满、状态不允许、非房主操作和重复答题统一映射为明确的 4xx 错误。Redis 不可用时房间接口返回 503，不创建进程内临时房间。

## 验收

- 两个客户端可创建/加入同一房间并收到广播事件。
- 非房主不能开始，未准备玩家不能开始。
- 答题结果和结束事件能被所有在线客户端收到。
- 断开后 60 秒内重连保留玩家身份；Redis key 按 TTL 清理。
