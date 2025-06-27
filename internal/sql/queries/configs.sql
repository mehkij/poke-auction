-- name: UpsertConfig :exec
INSERT INTO configs (server_id, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (server_id, key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = now();

-- name: GetServerConfig :many
SELECT * FROM configs
WHERE server_id = $1;

-- name: DeleteConfig :exec
DELETE FROM configs
WHERE server_id = $1 AND key = $2;
