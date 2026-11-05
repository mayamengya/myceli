# Myceli API

Base URL: `EXPO_PUBLIC_API_URL` in the mobile app (local: `http://<your-laptop-ip>:8080`).
All bodies are JSON. Errors look like `{"error": "message"}`.
Auth = `Authorization: Bearer <Firebase ID token>`. Authenticated routes return `401` with `missing authorization header`, `invalid token`, or `user not registered`.

## GET /health
Auth: none. Checks the database connection.
- `200` `{"status": "healthy"}`
- `503` `{"status": "unhealthy", "error": "..."}`

## POST /users/register
Auth: Firebase ID token required, but the user does not need to be registered yet. Call once after Firebase sign-up. The Firebase UID is taken from the verified token, never the body. Upserts on `firebase_uid`, so calling it again is safe and updates `display_name`.

Request:
```json
{ "display_name": "Maya" }
```
- `201` `{"id": "<users.id uuid>"}`
- `400` invalid body, or missing `display_name`
- `401` `missing authorization header` or `invalid token`
- `500` `failed to register user`

## POST /messages
Auth: required. Drops a message at a location.

Request:
```json
{
  "content": "hi from the park",
  "latitude": 37.7694,
  "longitude": -122.4862,
  "radius_meters": 50,
  "is_public": false,
  "recipient_id": "<users.id uuid>"
}
```
- `content` required.
- `radius_meters` defaults to 50 if omitted or 0.
- `recipient_id` is ignored when `is_public` is true; stored as null.

Responses:
- `201` `{"id": "<message uuid>"}`
- `400` invalid body or missing `content`
- `500` `failed to create message`

## Planned
- Location update / unlock check (`ST_DWithin` against messages the user may see and hasn't unlocked)
- List the user's unlocked messages
- Push notification on unlock
