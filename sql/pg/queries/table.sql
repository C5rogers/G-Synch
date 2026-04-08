-- name: GetColumns :many
  SELECT
	     column_name as column_name,
	     data_type as data_type,
	     is_nullable as is_nullable,
	     column_default as column_default
	 FROM information_schema.columns
	 WHERE table_schema = sqlc.arg(schema_name) AND table_name = sqlc.arg(table_name)
	 ORDER BY ordinal_position;

-- name: GetForeignKeys :many
  SELECT
      kcu.column_name AS column_name,
      ccu.table_name AS foreign_table_name,
      ccu.column_name AS foreign_column_name,
      tc.table_schema AS foreign_table_schema
  FROM
      information_schema.table_constraints AS tc
      JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
      JOIN information_schema.constraint_column_usage AS ccu
        ON ccu.constraint_name = tc.constraint_name
        AND ccu.table_schema = tc.table_schema
  WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name = sqlc.arg(table_name);

-- name: GetForeignKeyDeferrability :many
  SELECT
      a.attname AS column_name,
      bool_and(c.condeferrable) AS is_deferrable
  FROM pg_constraint c
  JOIN pg_class rel ON rel.oid = c.conrelid
  JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
  JOIN unnest(c.conkey) AS keycols(attnum) ON true
  JOIN pg_attribute a
    ON a.attrelid = rel.oid
   AND a.attnum = keycols.attnum
  WHERE c.contype = 'f'
    AND nsp.nspname = sqlc.arg(schema_name)
    AND rel.relname = sqlc.arg(table_name)
  GROUP BY a.attname;

-- name: GetPrimaryKeys :many
  SELECT
      kcu.column_name as column_name
  FROM
      information_schema.table_constraints AS tc
      JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
  WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_schema = sqlc.arg(schema_name) AND tc.table_name = sqlc.arg(table_name)
  ORDER BY kcu.ordinal_position;

-- name: GetTables :many
  SELECT
      table_name as table_name
  FROM
      information_schema.tables
  WHERE table_schema = sqlc.arg(schema_name);
