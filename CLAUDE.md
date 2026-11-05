# Myceli backend (Go API)

Myceli is a location-based messaging app: a user drops a message at a GPS point with a radius, and it unlocks when the recipient (or anyone, if public) physically walks into that radius. This repo is the Go API. The mobile app lives in `../myceli-mobile`.

**Goal of this project:** a working two-phone demo, and code the author can explain and defend in interviews. Keep things simple and readable. Explain non-obvious tradeoffs before implementing them.

## Stack (actual, not aspirational)
- Go, `net/http` + `go-chi/chi/v5` router
- Postgres on Supabase with PostGIS, via `jackc/pgx/v5` (`pgxpool`)
- Firebase Auth: the mobile app signs in with Firebase; this API verifies the Firebase ID token with `firebase.google.com/go/v4`
- **There is no Rust service.** It was deliberately cut; PostGIS does the geospatial work. Do not add one.
- No ORM. Raw SQL with pgx.

## Layout
```
cmd/server/main.go          entrypoint: loads .env, creates DB pool + Firebase client, wires routes
internal/db/db.go           NewPool: pgxpool from DATABASE_URL, pings on startup (fail fast)
internal/auth/firebase.go   Client wrapper; VerifyToken(idToken) -> Firebase UID
internal/middleware/auth.go VerifyToken: Bearer token -> Firebase UID in context; RequireUser: Firebase UID -> users.id in context
internal/handlers/          one file per resource (health, users, messages)
```

## Running
```
go run ./cmd/server        # listens on $PORT (default 8080)
go build ./... && go vet ./...
```
Env vars (in `.env`, loaded by godotenv): `DATABASE_URL`, `PORT`, `FIREBASE_CREDENTIALS` (path to the service account JSON).

`.env` and `firebase-service-account.json` are gitignored. Never print, commit, or copy their contents.

## Testing endpoints with curl
Get a Firebase ID token (expires after ~1 hour) via the Firebase REST API, using `EXPO_PUBLIC_FIREBASE_API_KEY` from `../myceli-mobile/.env`:
```
curl -s -X POST "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=$FIREBASE_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"...","returnSecureToken":true}'
```
Use the `idToken` field: `curl -H "Authorization: Bearer $TOKEN" ...`. See `API.md` for every endpoint.

## Database
The schema lives in Supabase and is **not** in this repo yet. Tables used by the code:
- `users`: `id` (uuid), `firebase_uid` (unique), `display_name`
- `messages`: `id`, `sender_id`, `recipient_id` (nullable; null for public), `content`, `location` (`geography(Point)`), `radius_meters`, `is_public`

Rules:
- Locations are `geography`, not lat/lng floats, so PostGIS can index them and do spherical distance.
- `ST_MakePoint` takes **longitude first**, then latitude. Easy to get wrong.
- Proximity checks use `ST_DWithin(location, point, radius_meters)` in SQL, never distance math in Go.
- Don't change the schema without saying so explicitly and giving the SQL to run in Supabase.

## Conventions
- Handlers are structs holding `DB *pgxpool.Pool`; methods are `func (h *X) Name(w, r)`.
- The authenticated user's internal ID is in context: `r.Context().Value(middleware.ContextKeyUserID).(string)`. Use the internal `users.id` for foreign keys, not the Firebase UID.
- Errors: wrap with `fmt.Errorf("...: %w", err)`. Return JSON `{"error": "..."}` with the right status; don't leak raw DB errors to clients.
- Always use parameterized queries (`$1`, `$2`). Never build SQL with string concatenation.
- New authenticated routes go inside the `r.Group` with `authMiddleware.VerifyToken, authMiddleware.RequireUser` in `main.go`. Only `/users/register` uses `VerifyToken` alone.

## Ground rules for Claude
- Don't add dependencies, services, or frameworks without asking first.
- Don't touch `internal/auth` or `internal/middleware` without explaining why first.
- Don't change request/response shapes without updating `API.md` and flagging that `../myceli-mobile/services/api.ts` must change too.
- For any non-trivial feature: propose a plan and the tradeoffs first, then implement after approval.
- Prefer small diffs the author can read and understand.

## Known issues / TODO
- `messages.go` uses an unchecked type assertion on the context user ID (panics if middleware is missing; Recoverer turns that into a 500).
- `POST /messages` doesn't validate lat/lng ranges, radius bounds, or that `recipient_id` exists.
- No tests yet.
- Not built yet: location-update/unlock endpoint, listing nearby/unlocked messages, push notifications.
