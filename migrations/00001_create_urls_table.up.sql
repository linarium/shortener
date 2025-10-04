-- +goose Up
CREATE TABLE urls (
    id uuid PRIMARY KEY,
    short_url text NOT NULL UNIQUE,
    original_url text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_urls_short_url ON urls(short_url);
