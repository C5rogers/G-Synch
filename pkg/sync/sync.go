package sync

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sync struct {
	GivenDB  *pgxpool.Pool
	TargetDB *pgxpool.Pool
}

func NewSyncAPI(GivenDB, TargetDB *pgxpool.Pool) (*Sync, error) {
	s := &Sync{
		GivenDB:  GivenDB,
		TargetDB: TargetDB,
	}

	return s, nil
}

func (s *Sync) Synch(targetDB string, givenDB string, activityID *string, activityType *string, schema string) {
}
