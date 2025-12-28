-- +goose Up
-- +goose StatementBegin
CREATE TEMP TABLE compare_table (
  id   TEXT PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS compare_table;
-- +goose StatementEnd
