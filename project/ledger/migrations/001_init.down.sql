-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS budgets;
-- +goose StatementEnd
