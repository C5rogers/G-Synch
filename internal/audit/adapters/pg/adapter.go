package pg

import (
	pg_db "github.com/C5rogers/G-Synch/internal/audit/engines/pg/db"
)

type Adapter struct {
	db pg_db.DBTX
}

func New(db pg_db.DBTX) *Adapter {
	return &Adapter{db: db}
}

func (a *Adapter) Engine() string {
	return "postgres"
}
