-- name: GetColumns :exec
  SELECT
	     column_name,
	     data_type,
	     is_nullable,
	     column_default
	 FROM information_schema.columns
	 WHERE table_schema = sqlc.arg(schema_name) AND table_name = sqlc.arg(table_name)
	 ORDER BY ordinal_position;

-- name: GetForeignKeys :exec
  SELECT
      kcu.column_name,
      ccu.table_name AS foreign_table_name,
      ccu.column_name AS foreign_column_name
  FROM
      information_schema.table_constraints AS tc
      JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
      JOIN information_schema.constraint_column_usage AS ccu
        ON ccu.constraint_name = tc.constraint_name
        AND ccu.table_schema = tc.table_schema
  WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = sqlc.arg(schema_name) AND tc.table_name = sqlc.arg(table_name);

-- name: GetPrimaryKeys :exec
  SELECT
      kcu.column_name
  FROM
      information_schema.table_constraints AS tc
      JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
  WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_schema = sqlc.arg(schema_name) AND tc.table_name = sqlc.arg(table_name)
  ORDER BY kcu.ordinal_position;

-- name: GetTables :exec
  SELECT
      table_name
  FROM
      information_schema.tables
  WHERE table_schema = sqlc.arg(schema_name);