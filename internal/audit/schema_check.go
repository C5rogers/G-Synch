package audit

import (
	"context"
	"fmt"

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

	}
	return warnings, nil
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
