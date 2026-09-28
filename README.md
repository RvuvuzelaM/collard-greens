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
| DELETE | `/maps/{id}/areas/{areaId}` | delete an area, returns `204`; pins linked to it are kept with `area_id` set to `null`, its comments are deleted |
| GET    | `/maps/{id}/areas/{areaId}/comments` | list of comments on the area, oldest first: `[{"id","text","created_at"}]`, `404` if the area doesn't exist |
| POST   | `/maps/{id}/areas/{areaId}/comments` | add a comment to the area, returns `201` with the created comment |

| GET    | `/maps/{id}/pins` | list of pins on the map: `[{"id","name","description","status","map_id","area_id","coordinates"}]`, `404` if the map doesn't exist |
| POST   | `/maps/{id}/pins` | create a pin on the map, returns `201` with the created pin |
| GET    | `/maps/{id}/pins/{pinId}` | single pin, `404` if not found on that map |
| PATCH  | `/maps/{id}/pins/{pinId}` | partially update a pin (any subset of fields except `status`), returns the updated pin |
| DELETE | `/maps/{id}/pins/{pinId}` | delete a pin and its comments, returns `204`, `404` if not found on that map |
| PUT    | `/maps/{id}/pins/{pinId}/status` | set the pin's status with `{"status":"TODO"\|"DONE"}`, returns the updated pin, `404` if not found on that map |
| GET    | `/maps/{id}/pins/{pinId}/comments` | list of comments on the pin, oldest first: `[{"id","text","created_at"}]`, `404` if the pin doesn't exist |
| POST   | `/maps/{id}/pins/{pinId}/comments` | add a comment to the pin, returns `201` with the created comment |

An area body looks like `{"name","description","coordinates":[[x,y],...]}`; `map_id` comes from
the URL. `name` is required and `coordinates` needs at least 3 `[x, y]` points. Invalid input
returns `400`.

A pin body looks like `{"name","description","area_id","coordinates":[[x,y],...]}`; `map_id`
comes from the URL. `name` is required, `coordinates` is a list of at least 1 `[x, y]` point and
`area_id` is optional (`null` or omitted for no area). When set, `area_id` must be an area on the
same map. Send `"area_id": null` in a PATCH to unlink the area.

A pin's `status` is `TODO` or `DONE`. New pins start as `TODO`, and the status can only be changed
through `PUT /maps/{id}/pins/{pinId}/status`; sending `status` to POST or PATCH returns `400`.

A comment body looks like `{"text"}`; `text` is required. `id` and `created_at` (RFC 3339, UTC)
are set by the server.

Coordinates for areas and pins are pixels on the map image, origin top-left.

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

# list pins on map 1
curl http://localhost:8080/maps/1/pins

# create a pin on map 1 linked to area 1
curl -X POST http://localhost:8080/maps/1/pins \
  -H 'Content-Type: application/json' \
  -d '{"name":"Fix sign","description":"Hospital sign is crooked","area_id":"1","coordinates":[[60,400],[70,410]]}'

# create a pin on map 1 without an area
curl -X POST http://localhost:8080/maps/1/pins \
  -H 'Content-Type: application/json' \
  -d '{"name":"Pothole","coordinates":[[200,300]]}'

# get a single pin on map 1
curl http://localhost:8080/maps/1/pins/1

# mark pin 1 on map 1 as done
curl -X PUT http://localhost:8080/maps/1/pins/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"DONE"}'

# unlink pin 1 on map 1 from its area
curl -X PATCH http://localhost:8080/maps/1/pins/1 \
  -H 'Content-Type: application/json' \
  -d '{"area_id":null}'

# add a comment to area 1 on map 1
curl -X POST http://localhost:8080/maps/1/areas/1/comments \
  -H 'Content-Type: application/json' \
  -d '{"text":"Entrance is blocked by construction"}'

# list comments on area 1 on map 1
curl http://localhost:8080/maps/1/areas/1/comments

# add a comment to pin 1 on map 1
curl -X POST http://localhost:8080/maps/1/pins/1/comments \
  -H 'Content-Type: application/json' \
  -d '{"text":"Checked today, still blocked"}'

# list comments on pin 1 on map 1
curl http://localhost:8080/maps/1/pins/1/comments

# delete pin 1 on map 1
curl -i -X DELETE http://localhost:8080/maps/1/pins/1

# delete area 1 on map 1 (linked pins are unlinked, not deleted)
curl -i -X DELETE http://localhost:8080/maps/1/areas/1
```
