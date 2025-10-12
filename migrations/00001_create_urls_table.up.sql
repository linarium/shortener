-- +goose Up
CREATE TABLE urls (
    id uuid PRIMARY KEY,
    short_url varchar(100) NOT NULL UNIQUE,
    original_url varchar(2048) NOT NULL UNIQUE,
    user_id uuid NOT NULL,
    correlation_id varchar(100),
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_urls_short_url ON urls(short_url);
CREATE UNIQUE INDEX idx_urls_user_original ON urls(user_id, original_url) WHERE deleted_at IS NULL;
CREATE INDEX idx_urls_is_deleted ON urls(is_deleted);