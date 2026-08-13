# Plan: Architecture & Technical Decisions

## Tech Stack

| Layer            | Choice                             |
| ---------------- | ---------------------------------- |
| Backend          | Go + Gin                           |
| Frontend         | Next.js (App Router) + TypeScript  |
| Relational store | PostgreSQL                         |
| Document store   | MongoDB                            |
| Postgres access  | GORM                               |
| Mongo access     | official `mongo-go-driver`         |
| Auth             | JWT in an httpOnly cookie          |
| Validation       | go-playground/validator            |
| Data table       | TanStack Table                     |
| Styling          | Tailwind CSS                       |
| Backend tests    | `testing` + `testify` + `httptest` |
| Frontend tests   | Jest + React Testing Library       |
| Local infra      | Docker Compose (Postgres + Mongo)  |

## Request Flow

```
Client -> Gin handler -> service layer -> repository interface -> Postgres (sync)
                                                |
                                       log event -> buffered channel
                                                |
                                     background worker -> MongoDB (async)
```

The handler returns as soon as the Postgres write succeeds. The log event is enqueued (non-blocking) and drained by a worker goroutine started at boot; the client never waits on the Mongo write.

## API Design

| Method | Path                   | Auth       | Purpose                                        |
| ------ | ---------------------- | ---------- | ---------------------------------------------- |
| POST   | /api/v1/auth/login     | —          | Admin or User login                            |
| GET    | /api/v1/auth/me        | required   | Current identity from the session JWT (added Phase 5 — the frontend needs a cookie-based way to check auth state server-side without touching the token) |
| POST   | /api/v1/auth/logout    | —          | Clears the session cookie (added Phase 5)      |
| GET    | /api/v1/users          | required   | Paginated list                                 |
| POST   | /api/v1/users          | Admin only | Create user                                    |
| GET    | /api/v1/users/:id      | required   | Fetch one                                      |
| PUT    | /api/v1/users/:id      | Admin only | Update user                                    |
| DELETE | /api/v1/users/:id      | Admin only | Soft delete                                    |
| GET    | /api/v1/users/:id/logs | Admin only | _(optional/stretch)_ view a user's log history |

## Project Structure

```
user-records-app/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/          # hashing, JWT
│   │   ├── config/
│   │   ├── handler/       # Gin handlers
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   │   ├── postgres/
│   │   │   └── mongo/
│   │   ├── service/       # business logic
│   │   └── logging/       # async event bus + worker
│   ├── migrations/
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── login/            # login page
│   │   │   └── (protected)/      # server-side auth-gated route group
│   │   │       ├── layout.tsx    # redirects to /login if no session
│   │   │       ├── session-provider.tsx
│   │   │       └── users/        # data table + create/edit/delete/logs modals,
│   │   │                         # colocated since nothing here is shared
│   │   └── lib/                  # api.ts (client fetch), session.ts (server-side)
│   └── package.json
├── docker-compose.yml
├── spec.md / plan.md / tasks.md / AGENTS.md / progress.md
└── README.md              # written last
```

No `components/` directory exists — nothing needed sharing across routes, so everything stayed colocated per the Next.js convention in this file's Conventions section.

## Key Decisions & Rationale

1. **Password hashing** — HMAC-SHA256 over `salt + password`, keyed with a server secret from env, with a random salt generated per user at creation:
   ```
   hash = HMAC-SHA256(key=serverSecret, message=salt+password)
   ```
   Verification recomputes the HMAC and compares with `hmac.Equal()` — never `==` — which is the point of the "hmac.compare" requirement: it avoids timing attacks that a naive byte comparison is vulnerable to. The per-user salt is an addition on top of the literal requirement: a keyed HMAC alone still produces identical hashes for identical passwords across users, so the salt closes that gap.
2. **Async logging via in-process channel + worker**, not a message broker — satisfies "asynchronous communication style" without adding infrastructure the tech stack doesn't mention. Trade-off: not durable across a crash; a durable queue (Redis Streams, RabbitMQ) is the natural production upgrade — listed under Future Improvements, not built now.
3. **Soft delete** on Users so UserLogs stay meaningful after a user is removed.
4. **JWT in an httpOnly cookie**, not localStorage — keeps the token out of reach of XSS.
5. **UUID primary keys** — avoids sequential-ID enumeration on a panel that manages other people's records.
6. **Repository interfaces** for both stores — handlers/services depend on interfaces, not concrete drivers, so unit tests can mock the data layer instead of standing up real databases.
7. **Case-insensitive email uniqueness** — normalize/lower-case on write and index accordingly, so `Name@x.com` and `name@x.com` can't both be created.
8. **Two datastores, not one** — Users and UserLogs have different shapes and access patterns: Users is a small, fixed schema that needs strong consistency (unique email, atomic updates), while UserLogs is high-volume, append-only, and its `data` payload varies by event type. PostgreSQL fits the first; a document store fits the second without a migration every time a new event type's payload shape changes. A single Postgres instance with a JSONB log table would also work technically — but the requirements specify RDBMS for users and NoSQL for logs separately, so this isn't a preference to relax without changing scope.
9. **CORS restricted to a single configured origin** (`FRONTEND_ORIGIN`, `AllowCredentials: true`) — added in Phase 5 once the frontend needed to call the API cross-origin in dev (`:3000` → `:8080`). A wildcard origin can't be combined with credentialed cookies per the CORS spec, so this had to be an explicit origin from day one, not a "tighten later" item.

## Future Improvements (explicitly not building now)

- Durable queue for log events instead of an in-process channel
- Role-based access beyond Admin/User
- Password reset flow
