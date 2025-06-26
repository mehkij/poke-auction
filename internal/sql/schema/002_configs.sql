-- +goose Up
CREATE TABLE configs(
    server_id BIGINT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (server_id, key),
    CONSTRAINT fk_server FOREIGN KEY (server_id)
        REFERENCES servers(id)
        ON DELETE CASCADE
)

-- +goose Down
DROP TABLE configs;
