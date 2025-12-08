package pg

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAdapter struct {
	db *pgxpool.Pool
}

func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	return &PostgresAdapter{db: pool}
}

// here also implement the audit functionality using schema level
func (p *PostgresAdapter) Name() string {
	return "PostgresAdapter"
}
