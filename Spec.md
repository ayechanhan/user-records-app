# Spec: User Records Management System

## Overview

A user management system with an admin panel for creating, editing, and deleting user records. Every user created through the panel can independently authenticate and log in. Every create, update, or delete against a user record produces an auditable log event, written asynchronously so it never blocks the request that triggered it.

## Goals

- A secure admin panel with full CRUD over user records, presented as a data table.
- Any user created via the panel can log in on their own.
- A reliable, queryable audit trail of every user-affecting action.
- A write path for user data that stays fast because log writes are decoupled from the request lifecycle.

## Non-Goals

- Role tiers beyond Admin and User — a logged-in User does not gain management access over other users (see Assumptions).
- Self-service signup — users are only created by the Admin.
- Password reset / forgot-password flow.
- Email verification or transactional email.
- Multi-tenancy.
- Real-time log streaming — logs are queried on demand, not pushed live.

## Actors

- **Admin** — authenticates with credentials provided via configuration.
- **User** — a record created by the Admin; can authenticate through the same login flow but has no management permissions.

## Data Model

### Users (relational)

| Field         | Type                | Notes                                                  |
| ------------- | ------------------- | ------------------------------------------------------ |
| id            | UUID                | primary key                                            |
| name          | string              | required                                               |
| email         | string              | required, unique (case-insensitive, active rows only)  |
| password_hash | string              | HMAC hash — never returned by the API                  |
| password_salt | string              | per-user random salt, see plan.md Key Decisions #1     |
| created_at    | timestamp           |                                                         |
| updated_at    | timestamp           |                                                         |
| deleted_at    | timestamp, nullable | soft delete, see Assumptions                           |

### UserLogs (document store)

| Field      | Type      | Notes                                                                             |
| ---------- | --------- | --------------------------------------------------------------------------------- |
| user_id    | string    | references Users.id                                                               |
| event      | string    | `user.created`, `user.updated`, `user.deleted`, `user.login`, `user.login_failed` |
| data       | object    | event-specific payload                                                            |
| created_at | timestamp | event time                                                                        |

## Functional Requirements

1. Admin authenticates with email + password; the password check uses a constant-time HMAC comparison, not `==`.
2. Admin can create, list (as a paginated data table), edit, and delete users.
3. Any user created by the Admin can independently log in through the same authentication flow.
4. Every create, update, or delete on a User produces exactly one UserLog event.
5. Log events are written asynchronously — the triggering request does not wait on the log write.
6. User data is stored in a relational database; log data is stored in a NoSQL document store.
7. Unit tests cover authentication, every CRUD path, and the async logging mechanism.

## Assumptions

- **Admin identity**: the requirements separate "admin login with provided credentials" from "any created user can log in," and the Users schema has no role field. Treating the Admin as a single identity from configuration, not a Users row — otherwise "admin panel" has no real boundary once any created user can also authenticate.
- **Deletion**: soft delete, so historical log entries stay meaningful after a user is removed, and a soft-deleted user can no longer log in.

## Definition of Done

- Users data working end-to-end against the relational store.
- UserLogs data working end-to-end against the document store, written asynchronously.
- Unit test suite passing.
- spec.md / plan.md / tasks.md / AGENTS.md / progress.md kept current as the AI-driven workflow record.
