CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    password_salt TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Case-insensitive uniqueness, scoped to active rows so an email can be reused
-- after the row that held it is soft-deleted.
CREATE UNIQUE INDEX users_email_unique_idx ON users (lower(email)) WHERE deleted_at IS NULL;
