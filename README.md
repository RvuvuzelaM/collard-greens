# green-field API

A small REST API for the frontend exercise. You only need Docker.

## Start

From the `backend/` folder:

```
docker compose up --build
```

The API is then available at **http://localhost:8080**. Stop it with `Ctrl+C`.

CORS is open, so you can call it from any local dev server (Vite, Next.js, Angular, plain HTML, ...).

## Endpoints

| Method | Path       | Response          |
|--------|------------|-------------------|
| GET    | `/healthz` | `{"status":"ok"}` |
| GET    | `/maps`    | list of maps: `[{"id","name","imageUrl"}]` |
| GET    | `/maps/{id}` | single map, `404` if not found |

Every response is JSON. Errors have the shape `{"error":"<message>"}`, for example an unknown
route returns `404` with `{"error":"not found"}`.
