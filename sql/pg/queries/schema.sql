-- name: LoadSchema :many
SELECT table_name AS name
  FROM information_schema.tables
  WHERE table_schema = sqlc.arg(schema_name)
ORDER BY table_name;