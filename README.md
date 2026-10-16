# green-field API

A small REST API for the frontend exercise. You only need Docker.

## Start

From the `backend/` folder:

```
docker compose up --build
```

The API is then available at **http://localhost:8080**. Stop it with `Ctrl+C`.

CORS is open, so you can call it from any local dev server (Vite, Next.js, Angular, plain HTML, ...).

## Data and mock

Everything is stored in `./data/maps.json`, which is mounted into the container. It ships with one
mock map, id `sample-farm`: a 1600x1600 top-down drawing of a farm with four points on it (three
single pixels and one four-corner outline).

To reset to the mock state, run `git checkout -- data`.

## Endpoints

| Method | Path          | What                                                              |
|--------|---------------|-------------------------------------------------------------------|
| GET    | `/healthz`    | `{"status":"ok"}`                                                 |
| GET    | `/maps`       | `200` with a summary of every map (no image, no points)           |
| POST   | `/maps`       | add a map: `201` with the full map                                |
| GET    | `/maps/{id}`  | read a map: `200` with the full map, `404` if unknown             |
| PUT    | `/maps/{id}`  | edit a map (name, description, image and/or points): `200`        |

### Map

```json
{
  "id": "sample-farm",
  "name": "Sample Farm",
  "description": "Top-down view of a small farm ...",
  "image": "data:image/png;base64,iVBORw0KGgo...",
  "points": [
    { "name": "Farmyard", "description": "Barn and house next to the main road.", "coordinates": [[760, 470]] },
    { "name": "Orchard",  "description": "Fruit trees, outlined corner by corner.", "coordinates": [[1060, 655], [1560, 650], [1560, 1040], [1060, 1040]] }
  ],
  "createdAt": "2026-10-16T08:00:00Z",
  "updatedAt": "2026-10-16T08:00:00Z"
}
```

- `image` is the uploaded picture as base64. Send either a data URL (`data:image/png;base64,...`)
  or the bare base64 string; the API always stores and returns a data URL, so it can go straight
  into `<img src>`. PNG, JPEG, GIF and WebP are accepted. Request bodies may be up to 32 MiB.
- `points` are things drawn on the map. Each has a `name` (required), a `description` and
  `coordinates`: an array of `[x, y]` pixel pairs, so a single pin is `[[x, y]]` and an outline is
  several pairs. At least one pair is required.
- `POST /maps` needs `name` and `image`; `description` and `points` are optional.
- `PUT /maps/{id}` takes any subset of `name`, `description`, `image`, `points`. Fields you leave out
  stay as they are. `points` replaces the whole list, so to add a point send the existing points
  plus the new one.
- The list endpoint returns `{id, name, description, pointCount, createdAt, updatedAt}` per map.

Every response is JSON. Errors have the shape `{"error":"<message>"}`, for example an unknown route
returns `404` with `{"error":"not found"}`.

## Examples

```
# health check
curl http://localhost:8080/healthz

# list maps (contains the mock "Sample Farm")
curl http://localhost:8080/maps

# read the mock map (big: includes the base64 image)
curl http://localhost:8080/maps/sample-farm

# add a map from a local image (macOS base64; on Linux use: base64 -w0 map.png)
IMG=$(base64 -i map.png)
curl -X POST http://localhost:8080/maps \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"My field\",\"description\":\"Drone shot from May\",\"image\":\"$IMG\"}"

# edit a map: add points (replaces the whole points list)
curl -X PUT http://localhost:8080/maps/sample-farm \
  -H "Content-Type: application/json" \
  -d '{"points":[
        {"name":"Gate","description":"Main entrance","coordinates":[[120,480]]},
        {"name":"North field","description":"Wheat","coordinates":[[40,40],[600,30],[630,420],[70,440]]}
      ]}'

# edit a map: rename only
curl -X PUT http://localhost:8080/maps/sample-farm \
  -H "Content-Type: application/json" \
  -d '{"name":"Sample Farm (renamed)"}'

# validation errors -> 400
curl -i -X POST http://localhost:8080/maps -H "Content-Type: application/json" -d '{"name":"No image"}'
curl -i -X POST http://localhost:8080/maps -H "Content-Type: application/json" -d '{"name":"Bad image","image":"not-base64!"}'
curl -i -X PUT  http://localhost:8080/maps/sample-farm -H "Content-Type: application/json" -d '{"points":[{"name":"Broken","coordinates":[[1,2,3]]}]}'

# unknown map -> 404
curl -i http://localhost:8080/maps/nope
```
