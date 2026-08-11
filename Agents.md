# AGENTS.md

Rules for any AI assistant (Claude Code or otherwise) working in this repo. Claude Code specifically reads `CLAUDE.md`, not this file — `CLAUDE.md` at the repo root imports this one (`@AGENTS.md`) so both tools read the same rules from a single source.

## Before starting work

- Read `spec.md` and `plan.md` — they're the source of truth for scope and architecture.
- Check `tasks.md` for the next unchecked item. Work one task at a time.

## Hard rules

- **Never `git push` without explicit approval from Aye Chan** — ask every time, even after a clean commit.
- **Never write `README.md`** until every item in `tasks.md` is checked off.
- Commit messages follow Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`, `refactor:`) with a specific, non-generic summary.
- Never log, print, or return a raw or hashed password in any response or log line.
- Compare password hashes with `hmac.Equal()` — never `==` or `bytes.Compare()`.
- All SQL goes through parameterized queries / GORM — no string-concatenated SQL.

## Commands

- Backend dev: `go run ./cmd/server`
- Backend tests: `go test ./... -cover`
- Frontend dev: `npm run dev` (from `frontend/`)
- Frontend tests: `npm test`
- Local infra: `docker compose up -d`

_(update this section once scaffolding lands — these are the intended commands, not yet verified)_

## Conventions

- Go: `cmd/` + `internal/` layout, table-driven tests, errors wrapped with `fmt.Errorf("...: %w", err)`.
- Next.js: App Router, TypeScript, colocate components with the route that owns them unless shared.

## After finishing a task

- Check it off in `tasks.md`.
- Add one line to `progress.md`.
- Propose a commit message and wait for approval before pushing.
