# Production Compose 上线基线设计

## 目标

提供一套不影响现有开发环境的生产部署基线：单一入口通过 Nginx 暴露前端、管理端和后端 API，服务依赖使用健康检查控制启动顺序，密钥和跨域来源通过环境变量注入。

## 方案

- 新增 `docker-compose.prod.yml`，不修改现有 `docker-compose.yml`。
- PostgreSQL、Redis、MinIO 使用持久化命名卷，并增加健康检查。
- 后端增加 `/api/v1/ready` 健康检查，Compose 以该接口作为 backend 健康条件。
- 新增 `deploy/nginx.prod.conf` 和 `deploy/edge.Dockerfile`，edge 服务提供 80 端口：`/api/` 转发到 backend，`/socket/` 转发 WebSocket，其余路径按前端或管理端静态目录分流。
- frontend 与 admin 构建镜像内置 Nginx 配置，生产 Compose 只启动 edge，不再直接暴露两个静态服务端口。
- 后端 Compose 显式注入 `JWT_SECRET`、MinIO 凭据、`CORS_ORIGINS`，不在生产配置中依赖开发默认值。

## 运行与故障行为

依赖未 ready 时 backend 不进入 healthy；edge 仍可启动，但 API 返回由 backend 提供的 503。WebSocket 使用 `/api/v1/rooms/:id/ws`，Nginx 必须转发 Upgrade/Connection 头。所有服务使用 restart 策略，数据写入命名卷。

## 验收

使用 `docker compose -f docker-compose.prod.yml config` 校验配置；启动后访问 `/api/v1/ready` 应返回 200 且 postgres、redis、minio 均 ready，访问 edge 的 `/`、`/admin/`、`/api/v1/health` 均可达。
