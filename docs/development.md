# Development Baseline

## Local Verification

Run backend tests:

```bash
cd backend
go test ./...
```

Run player build:

```bash
cd frontend
npm ci
npm run build
```

Run admin build:

```bash
cd admin
npm ci
npm run build
```

On Windows sandboxes where Go cannot write user telemetry or temporary build files, run tests with project-local paths:

```powershell
New-Item -ItemType Directory -Force ..\.tmp, ..\.tmp\appdata, ..\.tmp\go-cache, ..\.tmp\go-tmp | Out-Null
$env:APPDATA = (Resolve-Path ..\.tmp\appdata).Path
$env:GOCACHE = (Resolve-Path ..\.tmp\go-cache).Path
$env:GOTMPDIR = (Resolve-Path ..\.tmp\go-tmp).Path
$env:TMP = (Resolve-Path ..\.tmp).Path
$env:TEMP = (Resolve-Path ..\.tmp).Path
go test ./...
```

## Data Strategy

The backend currently uses a lightweight startup initializer in the PostgreSQL repository:

- If `DATABASE_URL` is set and PostgreSQL is reachable, tables are created automatically and seed puzzles are inserted when the database is empty.
- If `DATABASE_URL` is missing or PostgreSQL is unavailable, the server falls back to the in-memory seed repository.
- Schema migrations are not implemented yet. Add a migration tool before production deployment in M5.

## Configuration

Current backend runtime variables:

- `BACKEND_PORT`: HTTP port. Defaults to `8080`.
- `DATABASE_URL`: PostgreSQL connection string. If omitted, in-memory data is used.

Prepared but not yet used by business code:

- `REDIS_ADDR`: reserved for M4 rooms and leaderboard state.
- `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`: reserved for M3 image upload.

## CI Scope

GitHub Actions runs three independent checks:

- Backend: `go test ./...` in `backend/`.
- Player frontend: `npm ci` and `npm run build` in `frontend/`.
- Admin frontend: `npm ci` and `npm run build` in `admin/`.

No lint step is configured yet because the repo has no lint scripts. Add lint scripts when code style rules are introduced.
