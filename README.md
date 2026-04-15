# Scheduler API

The request specifically asked for "seperation of concerns". So I isolated domain, data, and handler layers. One could argue for a project of this size it adds unneccessary boilerplate BUT given the original concern it helps isolate things. Changes in the request/response structs don't demand changes to the underlying data structs.

Other specific callouts exist in comments of the code where appropriate.

## Prerequisites

- Go `1.22+`
- Docker (for local Postgres)
- GNU Make

## Run Locally

Run Full Stack via Docker Compose

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

## Postman

Import `scheduler.postman_collection.json` into Postman.

It includes ready-to-run requests for all endpoints and uses collection variables for inputs.

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
