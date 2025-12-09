-- name: LoadSchema :exec
SELECT table_name
  FROM information_schema.tables
  WHERE table_schema = sqlc.arg(schema_name)
ORDER BY table_name;