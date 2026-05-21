# Backend Services Documentation

## Overview

The backend is a Go microservices architecture for a workout tracking platform. It consists of four services behind a single API Gateway entry point, using PostgreSQL for persistence and Redis for caching.

### Services

| Service | Port | Purpose |
|---------|------|---------|
| API Gateway | 8080 | Single entry point; JWT auth; routes to downstream services |
| User Service | 8081 | User profiles and workout splits |
| Exercise Service | 8083 | Exercise catalog and media uploads |
| Workout Service | 8082 | Workout sessions and set tracking (**HTTP layer not yet implemented**) |

### Technology Stack

- **Language**: Go (standard `net/http` + `chi` router in API Gateway)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Auth**: Auth0, RS256 JWT
- **Media storage**: Local filesystem

---

## Authentication & Authorization

### Mechanism

All non-health endpoints require an Auth0 JWT in the `Authorization: Bearer <token>` header. The API Gateway validates the token using Auth0's JWKS endpoint (5-minute key cache).

- Algorithm: RS256
- Required JWT claim: `sub` (used as `user_id` throughout the system)
- Environment: `AUTH0_DOMAIN`, `AUTH0_AUDIENCE`
- Issuer URL: `https://{AUTH0_DOMAIN}/`

### Inter-Service Authorization

Services trust the `X-User-ID` header for inbound calls from the API Gateway. The gateway sets this header to the JWT `sub` claim. Downstream services do **not** re-validate the JWT — they rely on the gateway to authenticate first.

### Ownership Checks

Resources are owned by the user whose `user_id` was set at creation time. Handlers enforce that the caller's `X-User-ID` matches the resource owner before allowing mutation or deletion. Mismatches return `403 Forbidden`.

---

## API Reference

Base URL: `http://localhost:8080` (gateway)

All endpoints return `application/json`. Mutation endpoints expect `Content-Type: application/json` unless noted.

### Health

#### `GET /health`

No auth required.

**Response 200**
```json
{"status": "ok"}
```

---

### Users

#### `POST /users`

Creates a user. The `user_id` is taken from the JWT subject — clients do not supply it.

**Body**
```json
{
  "email": "string (optional, unique)",
  "username": "string (required, unique)",
  "sport_goals": ["string"],
  "gender": "string (optional)",
  "birthdate": "string (optional)"
}
```

**Response 201** — User object (see [User model](#user))

---

#### `GET /users/{id}`

**Response 200** — User object, or **404**

---

#### `PUT /users/{id}`

Caller must be the owner (`X-User-ID == id`).

**Body** — Full User object (same shape as POST, `user_id` ignored)

**Response 200** — Updated User object, or **403/404**

---

#### `DELETE /users/{id}`

Caller must be the owner.

**Response 204**, or **403/404**

---

#### `PUT /users/{id}/split`

Updates the user's workout split. Caller must be the owner.

**Body**
```json
{
  "elements": [
    {
      "muscles": ["string"],
      "title": "string"
    }
  ]
}
```

**Response 200** — Updated User object, or **403/404**

---

### Exercises

#### `POST /exercises`

**Body**
```json
{
  "name": "string",
  "exercise_type": "string",
  "primary_muscle": "string",
  "secondary_muscles": ["string"],
  "description": "string",
  "user_id": "string",
  "is_private": false,
  "weight_direction": 0
}
```

**Response 201** — Exercise object (see [Exercise model](#exercise))

---

#### `GET /exercises?user_id={id}`

Lists exercises visible to the given user (their own + all public exercises).

**Query params**: `user_id` (required)

**Response 200** — Array of Exercise objects

---

#### `POST /exercises/recommendations`

Returns the top-N exercises ranked by how well they target a given muscle state.

**Scoring algorithm**: Each exercise is scored by matching its primary muscle (weight 1.2) and secondary muscles (weight 1.0) against the provided muscle scores. Final score = sum of matched weights / total possible weight. Results are sorted descending.

**Body**
```json
{
  "muscles": {
    "chest": 0.8,
    "triceps": 0.4
  },
  "user_id": "string",
  "limit": 10
}
```

Pass `limit: -1` to default to 10.

**Response 200** — Array of Exercise objects, sorted by relevance

---

#### `GET /exercises/{id}`

**Path param**: `id` (int64)

**Response 200** — Exercise object, or **404**

---

#### `POST /exercises/{id}/media`

Upload media for an exercise. Caller must own the exercise.

**Content-Type**: `multipart/form-data`

**Form field**: `file` (required)

**Allowed MIME types**: `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `video/mp4`, `video/quicktime`, `video/webm`

**Response 204**, or error

---

#### `GET /exercises/{id}/media`

**Response 200**
```json
[
  {
    "exercise_id": 1,
    "user_id": "string",
    "url": "string"
  }
]
```

---

### Workouts

> The workout service HTTP layer is not fully implemented. The API Gateway handlers proxy to the workout service using the client library, but the workout service itself has no running HTTP server (`cmd/main.go` is empty).

#### `GET /workouts/{workout_id}`

Returns a workout only if the caller owns it (otherwise 404).

**Response 200** — Workout object (see [Workout model](#workout))

---

#### `GET /workouts/{workout_id}/sets`

Returns all sets for a workout. Caller must own the workout.

**Response 200** — Array of Set objects (see [Set model](#set)), or empty array

---

### Sets

#### `PUT /sets/{set_id}`

Updates a set. Caller must own the workout that contains this set.

**Body** — Full Set object
```json
{
  "set_id": 1,
  "workout_id": 1,
  "exercise_id": 1,
  "reps": 10,
  "difficulty": 7,
  "weight": 80.0,
  "unit": "kg",
  "logged_at": "2024-01-01T10:00:00Z"
}
```

**Response 200** — Updated Set object, or **404**

---

#### `DELETE /sets/{set_id}`

Caller must own the workout that contains this set.

**Response 204**, or **404**

---

## Data Models

### User

```json
{
  "user_id": "string (Auth0 sub, primary key)",
  "email": "string (unique, nullable)",
  "username": "string (unique)",
  "sport_goals": ["string"],
  "gender": "string (nullable)",
  "birthdate": "string (nullable)",
  "split": {
    "elements": [
      {
        "muscles": ["string"],
        "title": "string"
      }
    ]
  }
}
```

Stored in PostgreSQL `users` table. `split` is a JSONB column.

### Exercise

```json
{
  "exercise_id": "int64 (auto-increment)",
  "name": "string",
  "exercise_type": "string",
  "primary_muscle": "string",
  "secondary_muscles": ["string"],
  "description": "string",
  "user_id": "string (owner)",
  "is_private": false,
  "weight_direction": "int64"
}
```

Stored in PostgreSQL `exercises` table. Media metadata stored in `exercise_media` table.

### Workout

```json
{
  "workout_id": "int64 (auto-increment)",
  "user_id": "string (owner)",
  "title": "string (optional)",
  "started_at": "timestamp (required, set by client)",
  "finished_at": "timestamp (optional, nullable)"
}
```

Stored in PostgreSQL `workouts` table.

### Set

```json
{
  "set_id": "int64 (auto-increment)",
  "workout_id": "int64",
  "exercise_id": "int64",
  "reps": "int",
  "difficulty": "int",
  "weight": "float64",
  "unit": "string",
  "logged_at": "timestamp (auto-set by service on create)"
}
```

Stored in PostgreSQL `workout_sets` table. `logged_at` is set to `NOW()` by the service on creation; the client value is ignored on create but accepted on update.

---

## Database Schema

```sql
-- users
CREATE TABLE users (
  user_id    TEXT PRIMARY KEY,
  email      TEXT UNIQUE,
  username   TEXT NOT NULL UNIQUE,
  sport_goals TEXT[] NOT NULL DEFAULT '{}',
  gender     TEXT,
  birthdate  TEXT,
  split      JSONB NOT NULL DEFAULT '{}'
);

-- exercises
CREATE TABLE exercises (
  exercise_id      SERIAL PRIMARY KEY,
  name             TEXT,
  exercise_type    TEXT,
  primary_muscle   TEXT,
  secondary_muscles TEXT[],
  description      TEXT,
  user_id          TEXT,
  is_private       BOOLEAN,
  weight_direction BIGINT
);

CREATE TABLE exercise_media (
  media_id    INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  exercise_id BIGINT,
  url         TEXT,
  user_id     TEXT
);

-- workouts
CREATE TABLE workouts (
  workout_id  INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  user_id     TEXT NOT NULL,
  title       TEXT,
  started_at  TIMESTAMP NOT NULL,
  finished_at TIMESTAMP
);

CREATE TABLE workout_sets (
  set_id      INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  workout_id  INTEGER,
  exercise_id INTEGER,
  reps        INTEGER,
  difficulty  INTEGER,
  weight      DOUBLE PRECISION,
  unit        TEXT,
  logged_at   TIMESTAMP
);
```

---

## Caching

Redis is used by the Exercise Service only.

| Cache key | TTL | Invalidated on |
|-----------|-----|----------------|
| `public_exercises` | 10 min | Any exercise update/delete |
| `exercise:{id}` | 10 min | Update or delete of that exercise |

User exercises and media endpoints are **not cached**.

---

## Middleware (API Gateway)

Applied in order to all requests:

1. **RequestID** — assigns a unique ID to each request
2. **RealIP** — extracts real client IP from proxy headers
3. **Logger** — standard access log
4. **Recoverer** — catches panics and returns 500
5. **CORS** — allows all origins; methods: GET, POST, PUT, DELETE, OPTIONS; exposes `Authorization` header
6. **RateLimiter** — token bucket per IP; configurable via env vars (see below); returns 429 on limit
7. **Auth** — validates Auth0 JWT; extracts `sub` claim as `user_id`; passes it downstream as `X-User-ID`

---

## Environment Variables

### API Gateway

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Listen port |
| `READ_TIMEOUT_MS` | `5000` | HTTP read timeout (ms) |
| `WRITE_TIMEOUT_MS` | `5000` | HTTP write timeout (ms) |
| `AUTH0_DOMAIN` | — | Auth0 tenant domain |
| `AUTH0_AUDIENCE` | — | Auth0 API audience |
| `USER_SERVICE_URL` | — | Base URL of user service |
| `EXERCISE_SERVICE_URL` | — | Base URL of exercise service |
| `WORKOUT_SERVICE_URL` | — | Base URL of workout service |
| `RATE_LIMIT_ENABLED` | `true` | Enable per-IP rate limiting |
| `RATE_LIMIT_RPS` | `100` | Requests per second per IP |

### User Service

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_DSN` | — | PostgreSQL connection string |
| `USER_SERVICE_PORT` | `8081` | Listen port |

### Exercise Service

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_DSN` | — | PostgreSQL connection string |
| `REDIS_DSN` | — | Redis connection string |
| `EXERCISE_SERVICE_PORT` | `8083` | Listen port |
| `MEDIA_BASE_ROOT` | — | Filesystem path for uploaded media |
| `MEDIA_BASE_URL` | — | Public base URL for serving media |

---

## HTTP Client Libraries

The API Gateway contains typed HTTP clients for each service under `services/api-gateway/internal/clients/`. All clients set `X-User-ID` on outbound requests.

| Client | Timeout | Package |
|--------|---------|---------|
| `userclient` | 5s | `clients/userclient` |
| `exerciseclient` | 10s | `clients/exerciseclient` |
| `workoutclient` | 10s | `clients/workoutclient` |

---

## Error Handling

Services return plain-text HTTP errors via `http.Error()`. There is no structured JSON error envelope.

| Status | Meaning |
|--------|---------|
| 400 | Malformed request body or missing required parameter |
| 401 | Missing or invalid JWT |
| 403 | Caller does not own the resource |
| 404 | Resource not found (also used when ownership check fails) |
| 429 | Rate limit exceeded |
| 500 | Internal service error |
| 502 | Upstream service unreachable or returned an error |

---

## Code Layout

```
services/
  api-gateway/
    cmd/main.go                         # Entry point
    internal/
      config/                           # Env-based config loading
      clients/
        userclient/                     # HTTP client → user service
        exerciseclient/                 # HTTP client → exercise service
        workoutclient/                  # HTTP client → workout service
      transport/http/
        handlers/                       # Route handlers (proxy to clients)
        middleware/                     # Auth, CORS, rate limiter, logger

  user-service/
    cmd/main.go
    internal/
      domain/                           # Models + repository/service interfaces
      service/                          # Business logic (thin pass-through)
      transport/http/handlers/          # HTTP handlers
      infrastructure/repository/        # PostgreSQL implementation

  exercise-service/
    cmd/main.go
    internal/
      domain/                           # Models + interfaces
      service/                          # Business logic + scoring algorithm
      transport/http/handlers/
      infrastructure/
        repository/                     # PostgreSQL + Redis cache
        storage/                        # Local media file storage

  workout-service/
    cmd/main.go                         # Empty — HTTP server not implemented
    internal/
      domain/                           # Models + interfaces
      service/                          # Business logic (thin pass-through)
      transport/http/handlers/          # Handlers defined but not wired
      infrastructure/repository/        # PostgreSQL implementation

schemas/                                # SQL DDL files
shared/
  env/                                  # GetString/GetInt/GetBool helpers
```

---

## Known Gaps

- **Workout service has no running HTTP server.** Domain, service, and repository layers exist but `cmd/main.go` is empty.
- **No `POST /workouts` endpoint.** There is no way to create a workout via the API.
- **`PUT /exercises/{id}` and `DELETE /exercises/{id}` exist in the service layer** but are not registered in the exercise service router.
- **User deletion does not cascade to workouts.** Deleting a user does not trigger deletion of their workouts or sets in the workout service.
- **No structured error responses.** Errors are plain text, not JSON.
