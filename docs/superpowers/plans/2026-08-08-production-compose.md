# Production Compose 上线基线 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an isolated production Compose deployment with one Nginx edge, healthy dependencies, and explicit runtime secrets.

**Architecture:** Keep the current development Compose untouched. Build frontend and admin static assets independently, expose them through one edge Nginx, and proxy `/api/` (including WebSocket upgrades) to the backend. PostgreSQL, Redis, MinIO, and backend use health checks and persistent named volumes.

**Tech Stack:** Docker Compose, Nginx, existing Go/Gin backend, existing Vue/Vite frontend and admin images.

---

### Task 1: Add edge image and routing configuration

**Files:**
- Create: `deploy/edge.Dockerfile`
- Create: `deploy/nginx.prod.conf`

- [ ] Create an Nginx edge image that copies frontend and admin build outputs into `/usr/share/nginx/html` and uses the production config.
- [ ] Route `/api/` to `backend:8080` with WebSocket upgrade headers and route `/admin/` to the admin static directory; serve the player SPA for all other paths.
- [ ] Run `docker build` for frontend, admin, and edge after Compose wiring is present.

### Task 2: Add isolated production Compose

**Files:**
- Create: `docker-compose.prod.yml`

- [ ] Define postgres, redis, minio, backend, and edge services with named volumes; let the edge multi-stage image build both Vue applications.
- [ ] Add health checks for postgres, redis, minio, and backend readiness; make backend depend on healthy data services and edge depend on healthy backend.
- [ ] Pass `JWT_SECRET`, MinIO credentials, and `CORS_ORIGINS` through `${...}` variables with safe compose-time required checks.
- [ ] Keep only edge port `80:80` public while data services stay internal to the Compose network.

### Task 3: Verify production configuration and runtime

**Files:**
- Modify: `backend/.env.example` (document production variables only if missing)

- [ ] Run `docker compose -f docker-compose.prod.yml config` and verify no unresolved required variables.
- [ ] Build and start the stack with an isolated project name and local test secrets.
- [ ] Query edge `/api/v1/ready`, edge `/api/v1/health`, player `/`, and admin `/admin/`; verify backend readiness reports all three dependencies ready.
- [ ] Run `go test ./...`, `npm run build` in `frontend`, `npm run build` in `admin`, and `git diff --check`.
- [ ] Commit with `feat: add production compose deployment`.
