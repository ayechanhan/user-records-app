# Tasks

Ordered checklist — each item is meant to be roughly one focused work session.

## Phase 0 — Scaffolding

- [ ] Repo structure (`backend/`, `frontend/`), `.gitignore`
- [ ] Go module + Gin boilerplate, config/env loading
- [ ] Next.js app init (TypeScript, App Router, Tailwind)
- [ ] `docker-compose.yml` for local Postgres + Mongo
- [ ] `.env.example`

## Phase 1 — Data layer

- [ ] Postgres migration for `users` (via golang-migrate or goose)
- [ ] Mongo indexes for `user_logs` (on `user_id`, `event`, `created_at`)
- [ ] `UserRepository` / `LogRepository` interfaces + Postgres/Mongo implementations

## Phase 2 — Auth

- [ ] Admin credentials from env/config
- [ ] Password hashing + `hmac.Equal` comparison utility
- [ ] `POST /auth/login`
- [ ] JWT issuing + auth middleware
- [ ] Unit tests: hash/compare correctness, login success/failure/unknown-email paths

## Phase 3 — Async logging mechanism

- [ ] `LogEvent` type + buffered channel
- [ ] Background worker: drains channel → writes to Mongo
- [ ] Non-blocking helper to emit events from handlers
- [ ] Unit tests: event enqueued and consumed correctly, using a mocked writer

## Phase 4 — User CRUD API

- [ ] `POST /users` — create, emit `user.created`
- [ ] `GET /users` — paginated list
- [ ] `GET /users/:id`
- [ ] `PUT /users/:id` — update, emit `user.updated` with a diff in `data`
- [ ] `DELETE /users/:id` — soft delete, emit `user.deleted`
- [ ] Unit tests: happy path + validation errors + not-found for each handler

## Phase 5 — Frontend auth

- [ ] Login page
- [ ] Auth state (httpOnly cookie–based) + protected route wrapper

## Phase 6 — Frontend user management

- [ ] User list — data table (TanStack Table), pagination
- [ ] Create user form/modal
- [ ] Edit user form/modal
- [ ] Delete confirmation
- [ ] Loading / error states

## Phase 7 — Testing pass

- [ ] Backend coverage across auth, CRUD, async logging
- [ ] Frontend component tests: login form, user table, create/edit forms

## Phase 8 — Optional / stretch

- [ ] `GET /users/:id/logs` read endpoint, simple log viewer in the UI

## Phase 9 — Wrap-up

- [ ] Update spec.md / plan.md to reflect what was actually built (spec drift check)
- [ ] Final progress.md entries
- [ ] Write README.md
- [ ] Final commit, push after approval
