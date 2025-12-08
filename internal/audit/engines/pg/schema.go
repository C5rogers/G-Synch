package pg

import (
	"github.com/C5rogers/G-Synch/internal/audit/core"
)

func (p *PostgresAdapter) LoadSchema(dsn string) (*core.Schema, error) {
	// TODO: implement to load schema from the given dsn
	// 	SELECT table_name
	// FROM information_schema.tables
	// WHERE table_schema = 'public'
	// ORDER BY table_name;
	return nil, nil
}
