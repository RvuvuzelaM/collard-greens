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
| GET    | `/maps`    | `200` with `[{"id","name","contentType","createdAt","areas"}, ...]` (no image data) |
| POST   | `/maps`    | `201` with `{"id","name","contentType","createdAt","areas"}` |
| GET    | `/maps/{id}/image` | raw image bytes (`Content-Type` of the image) |
| GET    | `/maps/{id}/areas` | `200` with `[area, ...]` |
| POST   | `/maps/{id}/areas` | `201` with the new area |
| PATCH  | `/areas/{id}` | `200` with the updated area |
| DELETE | `/areas/{id}` | `204` |

`POST /maps` takes `multipart/form-data` with an `image` file field (max 10 MB) and an optional
`name` field (defaults to the file name).

An area is a marked region on a map:

```json
{
  "id": "…",
  "mapId": "sample-park",
  "title": "Broken bench",
  "status": "TODO",
  "color": [255, 128, 0],
  "coords": [[120, 340], [180, 360]],
  "createdAt": "2026-10-05T12:00:00Z"
}
```

- `coords` is a list of `[x, y]` pixel positions on the map image: whole numbers, not negative,
  at least one point.
- `color` is `[r, g, b]`, each `0`–`255`.
- `status` is `TODO` or `DONE`, and defaults to `TODO` on create.
- `title` is required.

`POST /maps/{id}/areas` takes `title`, `color`, `coords` and an optional `status` as JSON.
`PATCH /areas/{id}` takes any subset of those fields and leaves the rest unchanged. Invalid input
returns `400`; an unknown map or area returns `404`.

Data is stored in a single JSON file (`./data/db.json`, mounted into the container). Images are
base64-encoded inside it. The file is re-read on every request, so you can edit it by hand.

A sample map is included: an aerial photo of a park (626 × 582 px), with id `sample-park`. It has
one sample area, `sample-out-of-bounds`, whose points deliberately go past the image's right and
bottom edges, so you can check how the frontend handles coordinates outside the image.

Every response is JSON. Errors have the shape `{"error":"<message>"}`, for example an unknown
route returns `404` with `{"error":"not found"}`.

## Examples

```
# health check
curl http://localhost:8080/healthz

# list all maps
curl http://localhost:8080/maps

# upload a map
curl -F "image=@map.png" -F "name=Office floor 1" http://localhost:8080/maps

# download a map image
curl -o map.png http://localhost:8080/maps/<id>/image

# download the sample map
curl -o park.jpg http://localhost:8080/maps/sample-park/image

# list the areas of a map
curl http://localhost:8080/maps/sample-park/areas

# add an area to a map
curl -X POST -H "Content-Type: application/json" \
  -d '{"title":"Broken bench","color":[255,128,0],"coords":[[120,340],[180,360]]}' \
  http://localhost:8080/maps/sample-park/areas

# mark an area as done
curl -X PATCH -H "Content-Type: application/json" -d '{"status":"DONE"}' \
  http://localhost:8080/areas/<area-id>

# delete an area
curl -X DELETE http://localhost:8080/areas/<area-id>

# unknown route -> 404
curl -i http://localhost:8080/nope
```
