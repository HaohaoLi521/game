# 这是谐音梗

完整全栈版 Web 小游戏工程。当前实现范围是 M0 + M1：单人模式可玩、双答题模式、Go API、Vue 玩家端、独立 Vue 管理后台、PostgreSQL 题库和投稿审核。

## 项目结构

- `backend/`：Go + Gin API，负责题库、判题、提示、投稿审核和后台鉴权。
- `frontend/`：Vue 玩家端，负责首页、答题、结果和玩家投稿。
- `admin/`：独立 Vue 管理后台，负责管理员注册登录、题库管理和投稿审核。
- `docs/`：API、路线图和开发基线文档。
- `docker-compose.yml`：本地 PostgreSQL、Redis、MinIO、后端、玩家端和管理端编排。

## 本地启动

### 1. 启动数据库

```bash
docker compose up -d postgres
```

默认数据库为 `this_is_pun_admin`，宿主机端口为 `15435`。后端首次连接 PostgreSQL 时会自动建表并初始化 20 道种子题。

### 2. 启动后端

```bash
cd backend
go mod tidy
DATABASE_URL="postgres://pun:pun123@localhost:15435/this_is_pun_admin?sslmode=disable" go run ./cmd/server
```

默认监听 `http://localhost:8080`。如果没有设置 `DATABASE_URL`，后端会退回内存题库，适合快速试玩但数据不会持久化。

### 3. 启动玩家端

```bash
cd frontend
npm install
npm run dev
```

默认访问 `http://localhost:5173`。Vite 开发代理会把 `/api` 转发到 `http://localhost:8080`。

### 4. 启动管理后台

```bash
cd admin
npm install
npm run dev
```

默认访问 `http://localhost:5174`。第一次进入可以创建管理员账号和密码；登录后可以审核玩家投稿，也可以在“题库管理”里新增、编辑、删除正式题目。

## 当前能力

- 首页、答题页、答对反馈页。
- 手动输入 / 候选字格子两种答题模式。
- 玩家可在同一题中切换模式，答案文本保持同步。
- 候选字支持点击选字、拖拽到答案槽、点击撤回、拖回候选区撤回。
- 后端统一判题，支持答案别名和拼音相同算对。
- 提示系统：分类字数、第一个字、拼音。
- 玩家投稿页：`/submit`。
- 独立管理后台：支持注册/登录、投稿审核、通过入库、拒绝备注、题库新增/编辑/删除。
- 游客进度保存在 `localStorage`。

## 当前占位模块

以下接口已在后端路由中占位，但会返回 `501 Not Implemented`：

- `/api/v1/progress/*`：账号进度同步，计划在 M2 实现。
- `/api/v1/workshop/*`：创作工坊，计划在 M3 实现。
- `/api/v1/rooms/*`：多人房间，计划在 M4 实现。

## 验证命令

```bash
cd backend
go test ./...
```

```bash
cd frontend
npm ci
npm run build
```

```bash
cd admin
npm ci
npm run build
```

CI 会在 GitHub Actions 中执行同样的后端测试和前端构建。Windows 本地如果 Go 尝试写用户目录 telemetry，可参考 `docs/development.md` 使用项目内临时目录运行测试。

## 配置

根目录 `.env.example` 给出本地完整服务的默认配置；`backend/.env.example` 给出后端当前实际读取的配置项。Redis 和 MinIO 已在 Docker Compose 中准备好，但业务代码会在 M3/M4 阶段才接入。

## 后续里程碑

- M1.5：题目选择、设置页、解释路由、后台退出、基础错误态。
- M2：用户登录、JWT、进度同步、成就。
- M3：创意工坊、出题编辑器、图片上传。
- M4：多人 WebSocket 房间、经典模式、实时排行榜。
- M5：部署、性能优化、移动端适配、备份和监控。
