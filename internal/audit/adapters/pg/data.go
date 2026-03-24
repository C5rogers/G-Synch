package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/jackc/pgx/v5"
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

func (a *Adapter) GetUnsyncedPrimaryKeyValues(ctx context.Context, schemaName, tableName string) ([]string, error) {
	pkCols, err := a.GetPrimaryKeys(ctx, schemaName, &core.Table{Name: tableName})
	if err != nil {
		return []string{}, err
	}
	if len(pkCols) == 0 {
		return []string{}, fmt.Errorf("table %s.%s has no primary key", schemaName, tableName)
	}

	pkExprParts := make([]string, len(pkCols))
	for i, col := range pkCols {
		pkExprParts[i] = fmt.Sprintf("%s::text", pq.QuoteIdentifier(col))
	}

	pkSerializedExpr := fmt.Sprintf("concat_ws('::', %s)", strings.Join(pkExprParts, ", "))

	query := fmt.Sprintf("SELECT id FROM compare_table WHERE id NOT IN (SELECT %s::text FROM %s.%s);", pkSerializedExpr, pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(tableName))
	res, err := a.db.Query(ctx, query)
	if err != nil {
		return []string{}, err
	}
	defer res.Close()

	var values []string
	for res.Next() {
		var value string
		if err := res.Scan(&value); err != nil {
			return []string{}, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (a *Adapter) MigrateMissingRowsFrom(ctx context.Context, source core.SchemaAdapter, schema string, table core.Table) (int, int, []string, error) {
	sourceAdapter, ok := source.(*Adapter)
	if !ok {
		return 0, 0, nil, fmt.Errorf("adapter mismatch: postgres destination requires postgres source adapter")
	}

	pkCols := table.PrimaryKey
	if len(pkCols) == 0 {
		return 0, 0, nil, fmt.Errorf("table %s has no primary key", table.Name)
	}

	colNames := make([]string, 0, len(table.Columns))
	colIndex := map[string]int{}
	for idx, col := range table.Columns {
		colNames = append(colNames, col.Name)
		colIndex[strings.ToLower(col.Name)] = idx
	}

	destinationPKSet, err := a.fetchPrimaryKeySet(ctx, schema, table.Name, pkCols)
	if err != nil {
		return 0, 0, nil, err
	}

	sourceRows, err := sourceAdapter.fetchRows(ctx, schema, table.Name, colNames)
	if err != nil {
		return 0, 0, nil, err
	}

	// Collect rows that are missing in the destination.
	missingRows := make([][]interface{}, 0)
	for _, row := range sourceRows {
		pkRef, pkErr := SerializePrimaryKey(row, pkCols, colIndex)
		if pkErr != nil {
			continue
		}
		if _, exists := destinationPKSet[pkRef]; exists {
			continue
		}
		missingRows = append(missingRows, row)
	}

	if len(missingRows) == 0 {
		return 0, 0, nil, nil
	}

	// Attempt bulk insert using pgx CopyFrom (PostgreSQL COPY protocol).
	quotedCols := make([]string, len(colNames))
	for i, col := range colNames {
		quotedCols[i] = pq.QuoteIdentifier(col)
	}

	copyCount, copyErr := a.db.CopyFrom(
		ctx,
		pgx.Identifier{schema, table.Name},
		colNames,
		pgx.CopyFromRows(missingRows),
	)

	if copyErr == nil {
		return int(copyCount), 0, nil, nil
	}

	// CopyFrom failed (e.g., constraint violation on some rows).
	// Fall back to row-by-row insert to identify individual failures.
	migrated := 0
	denied := 0
	deniedPKs := make([]string, 0)

	for _, row := range missingRows {
		pkRef, pkErr := SerializePrimaryKey(row, pkCols, colIndex)
		if pkErr != nil {
			denied++
			deniedPKs = append(deniedPKs, fmt.Sprintf("<unresolved:%v>", pkErr))
			continue
		}

		if insertErr := a.insertRow(ctx, schema, table.Name, colNames, row); insertErr != nil {
			denied++
			deniedPKs = append(deniedPKs, pkRef)
			continue
		}

		migrated++
	}

	return migrated, denied, deniedPKs, nil
}

func (a *Adapter) fetchRows(ctx context.Context, schema string, table string, columns []string) ([][]interface{}, error) {
	quotedCols := make([]string, len(columns))
	for i, col := range columns {
		quotedCols[i] = pq.QuoteIdentifier(col)
	}

	query := fmt.Sprintf("SELECT %s FROM %s.%s", strings.Join(quotedCols, ", "), pq.QuoteIdentifier(schema), pq.QuoteIdentifier(table))
	res, err := a.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	allRows := make([][]interface{}, 0)
	for res.Next() {
		values, valueErr := res.Values()
		if valueErr != nil {
			return nil, valueErr
		}
		rowCopy := make([]interface{}, len(values))
		copy(rowCopy, values)
		allRows = append(allRows, rowCopy)
	}

	if rowsErr := res.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return allRows, nil
}

func (a *Adapter) fetchPrimaryKeySet(ctx context.Context, schema string, table string, pkCols []string) (map[string]struct{}, error) {
	quotedPKCols := make([]string, len(pkCols))
	for i, col := range pkCols {
		quotedPKCols[i] = pq.QuoteIdentifier(col)
	}

	query := fmt.Sprintf("SELECT %s FROM %s.%s", strings.Join(quotedPKCols, ", "), pq.QuoteIdentifier(schema), pq.QuoteIdentifier(table))
	res, err := a.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	pks := map[string]struct{}{}
	for res.Next() {
		values, valueErr := res.Values()
		if valueErr != nil {
			return nil, valueErr
		}
		parts := make([]string, len(values))
		for i, v := range values {
			parts[i] = fmt.Sprintf("%v", v)
		}
		pks[strings.Join(parts, "::")] = struct{}{}
	}

	if rowsErr := res.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return pks, nil
}

func (a *Adapter) insertRow(ctx context.Context, schema string, table string, columns []string, row []interface{}) error {
	quotedCols := make([]string, len(columns))
	placeholders := make([]string, len(columns))
	for i, col := range columns {
		quotedCols[i] = pq.QuoteIdentifier(col)
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s.%s (%s) VALUES (%s)",
		pq.QuoteIdentifier(schema),
		pq.QuoteIdentifier(table),
		strings.Join(quotedCols, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := a.db.Exec(ctx, query, row...)
	return err
}

func SerializePrimaryKey(row []interface{}, pkCols []string, colIndex map[string]int) (string, error) {
	parts := make([]string, len(pkCols))
	for i, pk := range pkCols {
		idx, ok := colIndex[strings.ToLower(pk)]
		if !ok || idx >= len(row) {
			return "", fmt.Errorf("primary key column %s is not found in row payload", pk)
		}
		parts[i] = fmt.Sprintf("%v", row[idx])
	}
	return strings.Join(parts, "::"), nil
}
