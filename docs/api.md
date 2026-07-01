# API 草案

## Health

- `GET /api/v1/health`

## Puzzle

- `GET /api/v1/puzzle-sets`
- `GET /api/v1/puzzle-sets/:id`
- `GET /api/v1/puzzle-sets/:id/puzzles`
- `GET /api/v1/puzzles/:id`
- `POST /api/v1/puzzles/:id/check`
- `POST /api/v1/puzzles/:id/hint`

`GET /api/v1/puzzles/:id` 不返回 `answer`、`answer_pinyin`、`answer_aliases`，但会返回一次性的 `attempt_id`。提示和提交答案都必须带同一个 `attempt_id`，后端用它记录实际提示次数。

`POST /api/v1/puzzles/:id/check` 请求体：

```json
{
  "attempt_id": "f2b0e6b3a0f84b2eb0ff0af6f5fd44c1",
  "answer": "小鸟依人",
  "answer_mode": "tiles",
  "elapsed_ms": 12000,
  "hints_used": 0
}
```

M1 阶段 `hints_used` 仅作为兼容字段，实际计分以后端 attempt 中记录的提示次数为准。

响应体：

```json
{
  "data": {
    "correct": true,
    "score": 100,
    "answer": "小鸟依人",
    "explanation": "小鸟 + 依人，组成成语小鸟依人。"
  }
}
```

`POST /api/v1/puzzles/:id/hint` 请求体：

```json
{
  "attempt_id": "f2b0e6b3a0f84b2eb0ff0af6f5fd44c1",
  "level": 1
}
```
