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
| POST   | `/maps`    | `201` with `{"id","name","contentType","createdAt"}` |
| GET    | `/maps/{id}/image` | raw image bytes (`Content-Type` of the image) |

`POST /maps` takes `multipart/form-data` with an `image` file field (max 10 MB) and an optional
`name` field (defaults to the file name).

Data is stored in a single JSON file (`./data/db.json`, mounted into the container). Images are
base64-encoded inside it. The file is re-read on every request, so you can edit it by hand.

A sample map is included: an aerial photo of a park, with id `sample-park`.

Every response is JSON. Errors have the shape `{"error":"<message>"}`, for example an unknown
route returns `404` with `{"error":"not found"}`.

## Examples

```
# health check
curl http://localhost:8080/healthz

# upload a map
curl -F "image=@map.png" -F "name=Office floor 1" http://localhost:8080/maps

# download a map image
curl -o map.png http://localhost:8080/maps/<id>/image

# download the sample map
curl -o park.jpg http://localhost:8080/maps/sample-park/image

# unknown route -> 404
curl -i http://localhost:8080/nope
```
