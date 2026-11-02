-- +goose Up
-- +goose StatementBegin
CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    category TEXT UNIQUE NOT NULL,
    limit_amount NUMERIC(14,2) NOT NULL CHECK (limit_amount > 0)
);

CREATE TABLE expenses (
    id SERIAL PRIMARY KEY,
    amount NUMERIC(14,2) NOT NULL CHECK (amount <> 0),
    category TEXT NOT NULL,
    description TEXT,
    date DATE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS budgets;
-- +goose StatementEnd
