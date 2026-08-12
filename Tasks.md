# Tasks

Ordered checklist — each item is meant to be roughly one focused work session.

## Phase 0 — Scaffolding

- [x] Repo structure (`backend/`, `frontend/`), `.gitignore`
- [x] Go module + Gin boilerplate, config/env loading
- [x] Next.js app init (TypeScript, App Router, Tailwind)
- [x] `docker-compose.yml` for local Postgres + Mongo
- [x] `.env.example`

## Phase 1 — Data layer

- [x] Postgres migration for `users` (via golang-migrate or goose)
- [x] Mongo indexes for `user_logs` (on `user_id`, `event`, `created_at`)
- [x] `UserRepository` / `LogRepository` interfaces + Postgres/Mongo implementations

## Phase 2 — Auth

- [x] Admin credentials from env/config
- [x] Password hashing + `hmac.Equal` comparison utility
- [x] `POST /auth/login`
- [x] JWT issuing + auth middleware
- [x] Unit tests: hash/compare correctness, login success/failure/unknown-email paths

## Phase 3 — Async logging mechanism

- [x] `LogEvent` type + buffered channel
- [x] Background worker: drains channel → writes to Mongo
- [x] Non-blocking helper to emit events from handlers
- [x] Unit tests: event enqueued and consumed correctly, using a mocked writer

## Phase 4 — User CRUD API

- [x] `POST /users` — create, emit `user.created`
- [x] `GET /users` — paginated list
- [x] `GET /users/:id`
- [x] `PUT /users/:id` — update, emit `user.updated` with a diff in `data`
- [x] `DELETE /users/:id` — soft delete, emit `user.deleted`
- [x] Unit tests: happy path + validation errors + not-found for each handler

## Phase 5 — Frontend auth

- [x] Login page
- [x] Auth state (httpOnly cookie–based) + protected route wrapper

## Phase 6 — Frontend user management

- [x] User list — data table (TanStack Table), pagination
- [x] Create user form/modal
- [x] Edit user form/modal
- [x] Delete confirmation
- [x] Loading / error states

## Phase 7 — Testing pass

- [x] Backend coverage across auth, CRUD, async logging
- [x] Frontend component tests: login form, user table, create/edit forms

## Phase 8 — Optional / stretch

- [ ] `GET /users/:id/logs` read endpoint, simple log viewer in the UI

## Phase 9 — Wrap-up

- [ ] Update spec.md / plan.md to reflect what was actually built (spec drift check)
- [ ] Final progress.md entries
- [ ] Write README.md
- [ ] Final commit, push after approval
