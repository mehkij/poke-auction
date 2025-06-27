-- +goose Up
CREATE TABLE servers(
    id TEXT PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
)

-- +goose Down
DROP TABLE servers;
