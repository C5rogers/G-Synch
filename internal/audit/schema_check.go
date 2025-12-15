package audit

import (
	"context"
	"fmt"
	"strings"

	"github.com/C5rogers/G-Synch/internal/audit/core"
)

type SchemaAudit struct{}

func (a *SchemaAudit) Name() string {
	return "schema-check"
}

func (a *SchemaAudit) Check(ctx context.Context, target core.SchemaAdapter, given core.SchemaAdapter, schemaName string) ([]string, error) {
	targetSchema, err := target.LoadSchema(ctx, schemaName)
	if err != nil {
		return nil, err
	}
	givenSchema, err := given.LoadSchema(ctx, schemaName)
	if err != nil {
		return nil, err
	}
	var warnings []string

	targetTables := mapTables(targetSchema.Tables)
	givenTables := mapTables(givenSchema.Tables)

	for name, tTable := range targetTables {
		gTable, exists := givenTables[name]
		if !exists {
			warnings = append(warnings, "missing table: "+name)
			continue
		}
		warnings = append(warnings, compareColumns(name, tTable, gTable)...)

		pkDiff, err := comparePrimaryKeyValues(ctx, target, given, schemaName, tTable)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("table %s: error comparing rows: %v", name, err))
			continue
		}
		warnings = append(warnings, pkDiff)
	}
	return warnings, nil
}

func serializeRow(row []interface{}) string {
	parts := make([]string, len(row))
	for i, v := range row {
		parts[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(parts, "|")
}

func comparePrimaryKeyValues(ctx context.Context, target core.SchemaAdapter, given core.SchemaAdapter, schemaName string, table core.Table) (string, error) {
	tPks, err := target.GetPrimaryKeyValues(ctx, schemaName, table.Name)
	if err != nil {
		return "", err
	}
	gPks, err := given.GetPrimaryKeyValues(ctx, schemaName, table.Name)
	if err != nil {
		return "", err
	}

	// Convert slices to map for quick lookup
	tMap := make(map[string]struct{}, len(tPks))
	for _, row := range tPks {
		key := serializeRow(row)
		tMap[key] = struct{}{}
	}
	gMap := make(map[string]struct{}, len(gPks))
	for _, row := range gPks {
		key := serializeRow(row)
		gMap[key] = struct{}{}
	}

	// Count unsynced rows
	diffCount := 0
	for key := range tMap {
		if _, ok := gMap[key]; !ok {
			diffCount++
		}
	}
	return fmt.Sprintf("table %s: %d unsynced rows", table.Name, diffCount), nil
}

func compareColumns(table string, target core.Table, given core.Table) []string {
	var issues []string

	tCols := mapColumns(target.Columns)
	gCols := mapColumns(given.Columns)

	for name, col := range tCols {
		gcol, ok := gCols[name]
		if !ok {
			issues = append(issues,
				fmt.Sprintf("table %s: missing column %s", table, col.Name))
			continue
		}
		if col.DataType != gcol.DataType {
			issues = append(issues,
				fmt.Sprintf("table %s: column %s of type %s mismatches with column %s of type %s", table, name, col.DataType, name, gcol.DataType))
		}
		if col.IsNullable != gcol.IsNullable {
			issues = append(issues,
				fmt.Sprintf("table %s: column %s nullable mismatch", table, col.Name))
		}
	}

	return issues
}
