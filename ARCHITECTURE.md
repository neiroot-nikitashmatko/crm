# ARCHITECTURE

## Стек

Vue SPA (`frontend/`) <-> Go API (`backend/`) <-> PostgreSQL

## Frontend

`views/` -> `components/<domain>/` -> `composables/` -> `api/` -> Go

Env: `VITE_API_BASE_URL`

## Backend (Go)

```
cmd/server/main.go
internal/handler -> service -> repository -> PostgreSQL
migrations/*.sql
```

Env: `DATABASE_URL`, `HTTP_ADDR`, `CORS_ORIGINS`, `JWT_SECRET`, `JWT_TTL_HOURS`

## API

REST JSON `/api/v1`: auth, leads, deals, tasks, catalog-products, users (admin).

## БД

- UUID id
- TIMESTAMPTZ даты
- миграции только через `backend/migrations`
- soft-delete в лидах/сделках через `deleted_at`

## Принципы

- UI не ходит напрямую в БД
- Composables работают только через `frontend/src/api/`
- Бизнес-правила в service-слое backend
