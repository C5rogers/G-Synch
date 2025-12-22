package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/lib/pq"
)

func (a *Adapter) GetPrimaryKeyValues(ctx context.Context, schemaName, tableName string) ([][]interface{}, error) {
	pkCols, err := a.GetPrimaryKeys(ctx, schemaName, &core.Table{Name: tableName})
	if err != nil {
		return nil, err
	}
	if len(pkCols) == 0 {
		return nil, fmt.Errorf("table %s.%s has no primary key", schemaName, tableName)
	}
	cols := strings.Join(pkCols, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s.%s", cols, pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(tableName))
	rows, err := a.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results [][]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(pkCols))
		for i := range vals {
			vals[i] = new(interface{})
		}
		if err := rows.Scan(vals...); err != nil {
			return nil, err
		}
		row := make([]interface{}, len(pkCols))
		for i, v := range vals {
			row[i] = *(v.(*interface{}))
		}
		results = append(results, row)
	}
	return results, nil
}
