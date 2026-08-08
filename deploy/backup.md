# 生产数据备份

先复制 `.env.prod.example` 为 `.env.prod`，填入生产密钥，并确保生产 Compose 已启动：

```powershell
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
.\deploy\backup.ps1 -EnvFile .env.prod
```

脚本会在 `.data/backups` 生成 PostgreSQL custom dump、Redis RDB 和 MinIO 数据卷压缩包。恢复前应停止 backend/edge，先用 `pg_restore` 导入 PostgreSQL，再替换 Redis 数据文件和 MinIO 数据卷；恢复完成后重新启动 Compose 并检查 `/api/v1/ready`。恢复操作会覆盖现有数据，必须由运维人员确认备份文件和目标环境后执行。
