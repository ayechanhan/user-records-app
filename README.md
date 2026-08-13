# User Records Management System

An admin panel for creating, editing, and deleting user records. Any user created through the panel can independently log in. Every create, update, or delete against a user record produces an auditable log event, written asynchronously so it never blocks the request that triggered it.

See [`spec.md`](./Spec.md) for the full requirements and [`plan.md`](./Plan.md) for the architecture and key decisions. [`progress.md`](./Progress.md) is a running log of what was built, in what order, and why — useful if you want the "why" behind a decision rather than just the "what."

## Tech stack

| Layer            | Choice                             |
| ---------------- | ---------------------------------- |
| Backend          | Go + Gin                           |
| Frontend         | Next.js (App Router) + TypeScript  |
| Relational store | PostgreSQL                         |
| Document store   | MongoDB                            |
| Auth             | JWT in an httpOnly cookie          |
| Data table       | TanStack Table                     |
| Styling          | Tailwind CSS                       |
| Backend tests    | `testing` + `testify` + `httptest` |
| Frontend tests   | Jest + React Testing Library       |
| Local infra      | Docker Compose (Postgres + Mongo)  |

## Prerequisites

- Go 1.25+
- Node.js 20+
- Docker (for local Postgres + Mongo)

## Setup

1. **Start local infra:**

   ```bash
   docker compose up -d
   ```

2. **Configure environment variables:**

   ```bash
   cp .env.example backend/.env
   ```

   Edit `backend/.env` — the Postgres/Mongo values already match docker-compose's defaults. Set real values for `JWT_SECRET`, `HMAC_SECRET`, `ADMIN_EMAIL`, and `ADMIN_PASSWORD` (any values work for local use).

   ```bash
   echo "NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1" > frontend/.env.local
   ```

3. **Run the backend** (from `backend/`):

   ```bash
   go run ./cmd/server
   ```

   This applies the Postgres migration and ensures the Mongo indexes exist automatically on boot. The server listens on `:8080` by default.

4. **Run the frontend** (from `frontend/`):

   ```bash
   npm install
   npm run dev
   ```

   Open [http://localhost:3000](http://localhost:3000). You'll be redirected to `/login`. Sign in with the `ADMIN_EMAIL` / `ADMIN_PASSWORD` you set above.

## Commands

| Command                | Where       | What                                 |
| ---------------------- | ----------- | ------------------------------------ |
| `go run ./cmd/server`  | `backend/`  | Run the API server                   |
| `go test ./... -cover` | `backend/`  | Run backend unit tests with coverage |
| `npm run dev`          | `frontend/` | Run the Next.js dev server           |
| `npm test`             | `frontend/` | Run frontend component tests         |
| `npm run build`        | `frontend/` | Production build (also typechecks)   |
| `npm run lint`         | `frontend/` | Lint                                 |
| `docker compose up -d` | repo root   | Start Postgres + Mongo               |
| `docker compose down`  | repo root   | Stop them                            |

## API overview

All routes are under `/api/v1`. Auth uses an `httpOnly` session cookie set by `/auth/login` — there is no bearer token to pass manually.

| Method | Path              | Auth       | Purpose                                |
| ------ | ----------------- | ---------- | -------------------------------------- |
| POST   | `/auth/login`     | —          | Admin or User login                    |
| GET    | `/auth/me`        | required   | Current identity from the session      |
| POST   | `/auth/logout`    | —          | Clears the session cookie              |
| GET    | `/users`          | required   | Paginated user list                    |
| POST   | `/users`          | Admin only | Create a user                          |
| GET    | `/users/:id`      | required   | Fetch one user                         |
| PUT    | `/users/:id`      | Admin only | Update a user                          |
| DELETE | `/users/:id`      | Admin only | Soft-delete a user                     |
| GET    | `/users/:id/logs` | Admin only | Paginated audit log history for a user |

"Admin" is a single identity from configuration (`ADMIN_EMAIL` / `ADMIN_PASSWORD`), not a row in the `users` table. Any user created through the panel authenticates through the same `/auth/login` endpoint and gets a `role: "user"` session, which can read but not write.

## Project structure

```
user-records-app/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/          # password hashing, JWT
│   │   ├── config/        # env loading
│   │   ├── handler/       # Gin handlers
│   │   ├── middleware/    # auth/admin route guards
│   │   ├── model/
│   │   ├── repository/    # interfaces + postgres/ and mongo/ implementations
│   │   ├── service/       # business logic
│   │   └── logging/       # async audit-log event bus + worker
│   └── migrations/
├── frontend/
│   └── src/
│       ├── app/
│       │   ├── login/
│       │   └── (protected)/   # server-side auth-gated routes
│       │       └── users/     # data table, create/edit/delete/log-viewer modals
│       └── lib/                # api.ts (client), session.ts (server)
└── docker-compose.yml
```

## Testing

```bash
# backend
cd backend && go test ./... -cover

# frontend
cd frontend && npm test
```

The repository layer (`internal/repository/postgres`, `internal/repository/mongo`) is intentionally not unit-tested — handlers and services depend on repository _interfaces_, which are mocked in tests instead. The repository implementations themselves were verified against real Postgres/Mongo during development (see `progress.md`).
