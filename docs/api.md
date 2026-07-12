# API Reference

Base URL: `/api/v1`

所有成功响应使用统一结构：

```json
{
  "data": {}
}
```

所有错误响应使用统一结构：

```json
{
  "error": "invalid request body"
}
```

常见状态码：

- `200 OK`：请求成功。
- `400 Bad Request`：请求体、路径参数或业务输入不合法。
- `401 Unauthorized`：缺少或无效的管理员 token。
- `404 Not Found`：资源不存在。
- `409 Conflict`：资源已存在或状态冲突。
- `500 Internal Server Error`：未预期服务端错误。
- `501 Not Implemented`：路线图占位模块尚未实现。

管理员接口需要请求头：

```http
Authorization: Bearer <admin-token>
```

## Health

### `GET /health`

响应：

```json
{
  "data": {
    "status": "ok",
    "service": "this-is-pun"
  }
}
```

## Puzzle

### `GET /puzzle-sets`

返回所有题集。

### `GET /puzzle-sets/:id`

返回单个题集。

### `GET /puzzle-sets/:id/puzzles`

返回题集下的公开题目列表。公开题目不会返回 `answer`、`answer_pinyin`、`answer_aliases`、`explanation`。

### `GET /puzzles/:id`

返回单个公开题目，并生成一次性的 `attempt_id`。提示和提交答案必须带同一个 `attempt_id`，后端用它记录实际提示次数。

### `GET /puzzles/:id/explanation`

返回题目的答案和解释，适合结果页或答题结束后回看使用。

响应：

```json
{
  "data": {
    "answer": "小鸟依人",
    "explanation": "小鸟 + 依人，组成成语小鸟依人。"
  }
}
```
### `POST /puzzles/:id/check`

请求体：

```json
{
  "attempt_id": "f2b0e6b3a0f84b2eb0ff0af6f5fd44c1",
  "answer": "小鸟依人",
  "answer_mode": "tiles",
  "elapsed_ms": 12000,
  "hints_used": 0
}
```

`hints_used` 当前仅作为兼容字段，实际计分以后端 attempt 中记录的提示次数为准。

响应：

```json
{
  "data": {
    "correct": true,
    "score": 100,
    "answer": "小鸟依人",
    "answer_mode": "tiles",
    "normalized": "小鸟依人",
    "expected_chars": 4,
    "elapsed_ms": 12000,
    "explanation": "小鸟 + 依人，组成成语小鸟依人。",
    "message": "答对了"
  }
}
```

### `POST /puzzles/:id/hint`

请求体：

```json
{
  "attempt_id": "f2b0e6b3a0f84b2eb0ff0af6f5fd44c1",
  "level": 1
}
```

响应：

```json
{
  "data": {
    "level": 1,
    "text": "答案是 4 个字",
    "score_if_correct": 90
  }
}
```

## Submission

### `POST /submissions`

玩家投稿。当前图片字段只接收 URL 数据，真正的文件上传将在 M3 接入 MinIO。

请求体：

```json
{
  "creator_name": "玩家A",
  "contact": "player@example.com",
  "puzzle_set_id": 1,
  "author_name": "玩家A",
  "hint_images": [
    {
      "id": "img-1",
      "url": "https://example.com/image.png",
      "label": "图一",
      "alt": "提示图"
    }
  ],
  "hint_text": "小鸟和人",
  "answer": "小鸟依人",
  "answer_pinyin": "xiao niao yi ren",
  "answer_aliases": ["小鸟依人"],
  "candidate_chars": [],
  "default_answer_mode": "manual",
  "supported_answer_modes": ["manual", "tiles"],
  "blank_template": "____",
  "category": "成语",
  "difficulty": 1,
  "explanation": "小鸟 + 依人。",
  "sort_order": 100
}
```

响应返回创建后的投稿，初始状态为 `pending`。

## Admin Auth

### `POST /admin/auth/register`

请求体：

```json
{
  "username": "admin",
  "password": "secret123"
}
```

响应：

```json
{
  "data": {
    "token": "...",
    "user": {
      "id": 1,
      "username": "admin",
      "created_at": "2026-07-12T00:00:00Z"
    },
    "expires_at": "2026-07-13T00:00:00Z"
  }
}
```

### `POST /admin/auth/login`

请求体同注册接口。响应同注册接口。

管理端可调用下方的登出接口撤销当前服务端会话。

### `POST /admin/auth/logout`

需要管理员 token。删除当前服务端 session，响应：

```json
{
  "data": {
    "logged_out": true
  }
}
```
## Admin Puzzle

以下接口均需要管理员 token。

- `GET /admin/puzzle-sets`：返回题集列表。
- `GET /admin/puzzles`：返回全部正式题目。
- `POST /admin/puzzles`：创建正式题目，请求体同投稿中的题目字段。
- `PUT /admin/puzzles/:id`：更新正式题目，请求体同创建题目。
- `DELETE /admin/puzzles/:id`：删除正式题目。

## Admin Submission

以下接口均需要管理员 token。

### `GET /admin/submissions?status=pending`

返回投稿列表。`status` 可选值：`pending`、`approved`、`rejected`。不传则返回全部。

### `POST /admin/submissions/:id/approve`

通过投稿并写入正式题库。请求体可为空对象：

```json
{}
```

### `POST /admin/submissions/:id/reject`

拒绝投稿。

请求体：

```json
{
  "review_note": "题面不够清晰"
}
```

## Roadmap Placeholders

以下路径当前是路线图占位，统一返回 `501 Not Implemented`：

- `/progress/*`：M2 用户进度同步。
- `/workshop/*`：M3 创意工坊。
- `/rooms/*`：M4 多人房间。

## Player Account

- POST /players/register��{ username, password }�������˺Ų����� JWT access/refresh token��
- POST /players/login��������ͬע�ᣬ���� JWT access/refresh token��
- GET /players/me/progress����Ҫ Authorization: Bearer <access_token>��������ͨ����Ŀ��¼��
- PUT /players/me/progress/:id����Ҫ��� token������ָ����Ŀ��ͨ��״̬��

