-- name: CreateTempCompareTable :exec
CREATE TEMP TABLE IF NOT EXISTS compare_table (
  id   TEXT PRIMARY KEY
);

-- name: CreateTempRecords :copyfrom
INSERT INTO compare_table (id) VALUES ($1);

-- name: TruncateCompareTable :exec
TRUNCATE TABLE compare_table;