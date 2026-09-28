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
| GET    | `/maps/{id}/areas` | list of areas on the map: `[{"id","name","description","map_id","coordinates"}]`, `404` if the map doesn't exist |
| POST   | `/maps/{id}/areas` | create an area on the map, returns `201` with the created area |
| GET    | `/maps/{id}/areas/{areaId}` | single area, `404` if not found on that map |
| PATCH  | `/maps/{id}/areas/{areaId}` | partially update an area (any subset of fields), returns the updated area |

An area body looks like `{"name","description","coordinates":[[x,y],...]}`; `map_id` comes from
the URL. `name` is required and `coordinates` needs at least 3 `[x, y]` points. Invalid input
returns `400`.

Every response is JSON. Errors have the shape `{"error":"<message>"}`, for example an unknown
route returns `404` with `{"error":"not found"}`.

## Examples

```
# health check
curl http://localhost:8080/healthz

# list all maps
curl http://localhost:8080/maps

# get a single map by id
curl http://localhost:8080/maps/1

# unknown map id -> 404
curl -i http://localhost:8080/maps/999

# create an area on map 1
curl -X POST http://localhost:8080/maps/1/areas \
  -H 'Content-Type: application/json' \
  -d '{"name":"Field A","description":"North field","coordinates":[[0,0],[1,0],[1,1],[0,1]]}'

# list areas on map 1
curl http://localhost:8080/maps/1/areas

# get a single area on map 1
curl http://localhost:8080/maps/1/areas/1

# update only the name of area 1 on map 1
curl -X PATCH http://localhost:8080/maps/1/areas/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Field A (renamed)"}'
```
