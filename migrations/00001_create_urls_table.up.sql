-- +goose Up
CREATE TABLE urls (
    id uuid PRIMARY KEY,
    short_url text NOT NULL UNIQUE,
    original_url text NOT NULL UNIQUE,
    user_id uuid NOT NULL,
    correlation_id text,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_urls_short_url ON urls(short_url);
CREATE UNIQUE INDEX idx_urls_user_original ON urls(user_id, original_url) WHERE deleted_at IS NULL;
