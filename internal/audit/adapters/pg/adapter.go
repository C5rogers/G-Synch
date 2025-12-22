package pg

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Adapter struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Adapter {
	return &Adapter{db: db}
}

func (a *Adapter) Engine() string {
	return "postgres"
}
