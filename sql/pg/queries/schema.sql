-- name: LoadSchema :many
SELECT table_name AS table_name
  FROM information_schema.tables
  WHERE table_schema = sqlc.arg(schema_name)
ORDER BY table_name;

-- name: ListSchemas :many
SELECT schema_name AS schema_name
  FROM information_schema.schemata
  WHERE schema_name <> 'information_schema'
    AND schema_name NOT LIKE 'pg_%'
ORDER BY schema_name;
