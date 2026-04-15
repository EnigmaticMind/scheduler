# Scheduler API

Small Go API for managing trainer appointments.

## Prerequisites

- Go `1.22+`
- Docker (for local Postgres)
- GNU Make

## Run Locally

### 1) Start database

```bash
make db
```

This starts Postgres on `localhost:5432` with defaults:

- user: `scheduler`
- password: `scheduler`
- database: `scheduler`

Schema and seed data are loaded from `docker/postgres/init`.

### 2) Start API

```bash
make run
```

API starts on `http://localhost:8080`.

Health check:

```bash
curl http://localhost:8080/health
```

## Alternative: Run Full Stack via Docker Compose

```bash
make up
```

This runs both `db` and `app` services.

Stop services:

```bash
make down
```

Follow logs:

```bash
make logs
```

## Environment Variables

- `PORT` (default: `8080`)
- `DATABASE_URL` (default: `postgres://scheduler:scheduler@localhost:5432/scheduler?sslmode=disable`)
- `POSTGRES_USER` (default: `scheduler`) - used by Docker Compose
- `POSTGRES_PASSWORD` (default: `scheduler`) - used by Docker Compose
- `POSTGRES_DB` (default: `scheduler`) - used by Docker Compose
- `POSTGRES_PORT` (default: `5432`) - used by Docker Compose

## API Endpoints

- `GET /health`
- `GET /appointments`
- `POST /appointments`
- `GET /appointments/schedules`

All list responses are wrapped as:

```json
{
  "data": []
}
```

Error responses are wrapped as:

```json
{
  "error": "message"
}
```

## Postman

Import `scheduler.postman_collection.json` into Postman.

It includes ready-to-run requests for all endpoints and uses collection variables for inputs.
