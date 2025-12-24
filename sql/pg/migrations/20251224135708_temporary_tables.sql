-- +goose Up
-- +goose StatementBegin
CREATE TEMP TABLE compare_table (
  id   TEXT PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
