# 这是谐音梗

完整全栈版 Web 小游戏工程。当前实现范围是 M0 + M1：单人模式可玩、双答题模式、Go API、Vue 玩家端、独立 Vue 管理后台、PostgreSQL 题库和投稿审核。

## 启动

### 后端

```bash
docker compose up -d postgres
cd backend
go mod tidy
DATABASE_URL="postgres://pun:pun123@localhost:15435/this_is_pun_admin?sslmode=disable" go run ./cmd/server
```

默认监听 `http://localhost:8080`。数据库使用全新的 `this_is_pun_admin`，宿主机端口是 `15435`，首次启动会自动建表并初始化 20 道题。

### 前端

```bash
cd frontend
npm install
npm run dev
```

默认访问 `http://localhost:5173`。前端通过 Vite 代理访问 `/api`。

### 独立管理后台

```bash
cd admin
npm install
npm run dev
```

默认访问 `http://localhost:5174`。第一次进入可以创建管理员账号和密码，登录后可以审核玩家投稿，也可以在“题库管理”里新增、编辑、删除正式题目。

玩家端点击“我要出题”会进入 `http://localhost:5173/submit`。玩家提交后进入待审核列表，管理员在后台点击“通过并入库”后，题目才会写入正式题库；点击“拒绝”则不会影响玩家题库。

## 当前能力

- 首页、答题页、答对反馈页
- 手动输入 / 候选字格子两种答题模式
- 玩家可在同一题中切换模式，答案文本保持同步
- 候选字支持点击选字、拖拽到答案槽、点击撤回、拖回候选区撤回
- 后端统一判题，支持答案别名和拼音相同算对
- 提示系统：分类字数、第一个字、拼音
- 玩家投稿页：`/submit`
- 独立管理后台：`admin/` 项目，支持注册/登录、投稿审核、通过入库、拒绝备注、题库新增/编辑/删除
- 游客进度保存在 `localStorage`

## 后续里程碑

- M2：用户登录、进度同步、成就
- M3：创意工坊、出题编辑器、图片上传
- M4：多人 WebSocket 房间
- M5：部署、性能优化、移动端适配
