-- name: UpsertServer :exec
INSERT INTO servers (id, name)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    updated_at = NOW();
    
-- name: GetServer :one
SELECT * FROM servers
WHERE id = $1;
