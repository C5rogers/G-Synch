package audit

import (
	"context"
	"fmt"
	"strings"

	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/C5rogers/G-Synch/internal/models"
)

type SchemaAudit struct{}

func (a *SchemaAudit) Name() string {
	return "schema-check"
}

func (a *SchemaAudit) Check(ctx context.Context, target core.SchemaAdapter, given core.SchemaAdapter, schemaName string) ([]models.CheckReturn, error) {
	targetSchema, err := target.LoadSchema(ctx, schemaName)
	if err != nil {
		targetSchemaError := models.CheckReturn{
			Message: fmt.Sprintf("Error loading target schema: %v", err),
			Type:    "ERROR",
			Label:   "ERROR",
		}
		return []models.CheckReturn{targetSchemaError}, err
	}
	givenSchema, err := given.LoadSchema(ctx, schemaName)
	if err != nil {
		givenSchemaError := models.CheckReturn{
			Message: fmt.Sprintf("Error loading given schema: %v", err),
			Type:    "ERROR",
			Label:   "ERROR",
		}
		return []models.CheckReturn{givenSchemaError}, err
	}
	var warnings []models.CheckReturn

	targetTables := mapTables(targetSchema.Tables)
	givenTables := mapTables(givenSchema.Tables)

	for name, tTable := range targetTables {
		gTable, exists := givenTables[name]
		if !exists {
			if name == "compare_table" {
				continue
			}
			newCheck := models.CheckReturn{
				Message: fmt.Sprintf("MISSING TABLE: table %s is missing in %s schema", name, givenSchema.Name),
				Type:    "MISSING",
				Label:   "WARNING",
			}
			warnings = append(warnings, newCheck)
			continue
		}
		warnings = append(warnings, compareColumns(name, tTable, gTable)...)

		pkDiff, err := comparePrimaryKeyValuesUsingTempTable(ctx, target, given, schemaName, tTable)
		if err != nil {
			warnings = append(warnings, models.CheckReturn{
				Message: fmt.Sprintf("TABLE ERROR %s: error comparing rows: %v", name, err),
				Type:    "ERROR",
				Label:   "ERROR",
			})
			continue
		}
		if pkDiff.Message != "" {
			warnings = append(warnings, pkDiff)
		}

		fkIssues, err := compareForeignKeys(ctx, given, schemaName, tTable)
		if err != nil {
			warnings = append(warnings, models.CheckReturn{
				Message: fmt.Sprintf("TABLE ERROR %s: error comparing foreign keys: %v", name, err),
				Type:    "ERROR",
				Label:   "ERROR",
			})
			continue
		}
		warnings = append(warnings, fkIssues...)

	}
	return warnings, nil
}

func compareForeignKeys(ctx context.Context, given core.SchemaAdapter, schemaName string, table core.Table) ([]models.CheckReturn, error) {
	var issues []models.CheckReturn
	for _, fk := range table.ForeignKeys {
		if fk.ReferencedTableSchema != schemaName {
			dependencySchema, err := given.LoadSchema(ctx, fk.ReferencedTableSchema)
			if err != nil {
				issues = append(issues, models.CheckReturn{
					Message: fmt.Sprintf("MISSING DEPENDENCY SCHEMA: table %s depends on schema %s which is missing", table.Name, fk.ReferencedTableSchema),
					Type:    "MISSING_DEPENDENCY",
					Label:   "DEPENDENCY",
				})
				continue
			}
			if dependencySchema == nil || len(dependencySchema.Tables) == 0 {
				issues = append(issues, models.CheckReturn{
					Message: fmt.Sprintf("MISSING DEPENDENCY SCHEMA: table %s depends on schema %s which is missing", table.Name, fk.ReferencedTableSchema),
					Type:    "MISSING_DEPENDENCY",
					Label:   "DEPENDENCY",
				})
				continue
			}
			dependencyTable, exists := mapTables(dependencySchema.Tables)[fk.ReferencedTable]
			if !exists {
				issues = append(issues, models.CheckReturn{
					Message: fmt.Sprintf("MISSING DEPENDENCY TABLE: table %s depends on table %s in schema %s which is missing", table.Name, fk.ReferencedTable, fk.ReferencedTableSchema),
					Type:    "MISSING_DEPENDENCY",
					Label:   "DEPENDENCY",
				})
				continue
			}
			dependencyColumnExists := false
			for _, col := range dependencyTable.Columns {
				if col.Name == fk.ReferencedColumn {
					dependencyColumnExists = true
					break
				}
			}
			if !dependencyColumnExists {
				issues = append(issues, models.CheckReturn{
					Message: fmt.Sprintf("MISSING DEPENDENCY COLUMN: table %s depends on column %s in table %s in schema %s which is missing", table.Name, fk.ReferencedColumn, fk.ReferencedTable, fk.ReferencedTableSchema),
					Type:    "MISSING_DEPENDENCY",
					Label:   "DEPENDENCY",
				})
			}
		}
	}
	return issues, nil
}

func serializeRow(row []interface{}) string {
	parts := make([]string, len(row))
	for i, v := range row {
		parts[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(parts, "|")
}

func comparePrimaryKeyValuesUsingTempTable(ctx context.Context, target core.SchemaAdapter, given core.SchemaAdapter, schemaName string, table core.Table) (models.CheckReturn, error) {
	tPks, err := target.GetPrimaryKeyValues(ctx, schemaName, table.Name)
	if err != nil {
		return models.CheckReturn{
			Message: fmt.Sprintf("Error getting primary key values for table %s: %v", table.Name, err),
			Type:    "ERROR",
			Label:   "ERROR",
		}, err
	}
	err = given.CreateTemporaryTable(ctx)
	if err != nil {
		return models.CheckReturn{
			Message: fmt.Sprintf("Error creating temporary table for comparison: %v", err),
			Type:    "ERROR",
			Label:   "ERROR",
		}, err
	}

	err = given.TruncateTemporaryTable(ctx)
	if err != nil {
		return models.CheckReturn{
			Message: fmt.Sprintf("Error truncating temporary table for comparison: %v", err),
			Type:    "ERROR",
			Label:   "ERROR",
		}, err
	}

	var serializedTPKs []string
	for _, row := range tPks {
		serializedTPKs = append(serializedTPKs, serializeRow(row))
	}

	_, err = given.CreateTempRecords(ctx, serializedTPKs)
	if err != nil {
		return models.CheckReturn{
			Message: fmt.Sprintf("Error inserting records into temporary table for comparison: %v", err),
			Type:    "ERROR",
			Label:   "ERROR",
		}, err
	}
	res, err := given.GetUnsyncedPrimaryKeyValues(ctx, schemaName, table.Name)
	if err != nil {
		return models.CheckReturn{
			Message: fmt.Sprintf("Error searching primary key values in temporary table for comparison: %v", err),
			Type:    "ERROR",
			Label:   "ERROR",
		}, err
	}

	var returnableCheckReturn models.CheckReturn
	if len(res) > 0 {
		returnableCheckReturn = models.CheckReturn{
			Message: fmt.Sprintf("MISMATCH TABLE %s: %d unsynced rows", table.Name, len(res)),
			Type:    "MISMATCH",
			Label:   "WARNING",
		}
	}
	return returnableCheckReturn, nil
}

func compareColumns(table string, target core.Table, given core.Table) []models.CheckReturn {
	var issues []models.CheckReturn

	tCols := mapColumns(target.Columns)
	gCols := mapColumns(given.Columns)

	for name, col := range tCols {
		gcol, ok := gCols[name]
		if !ok {
			newIssue := models.CheckReturn{
				Message: fmt.Sprintf("MISSING COLUMN: table %s of given database missing column %s of target database", table, col.Name),
				Type:    "MISSING",
				Label:   "WARNING",
			}
			issues = append(issues, newIssue)
			continue
		}
		if col.DataType != gcol.DataType {
			issues = append(issues, models.CheckReturn{
				Message: fmt.Sprintf("MISMATCH TABLE %s: column %s of type %s of given database mismatches with column %s of type %s of target database", table, name, col.DataType, name, gcol.DataType),
				Type:    "MISMATCH",
				Label:   "WARNING",
			})
		}
		if col.IsNullable != gcol.IsNullable {
			issues = append(issues, models.CheckReturn{
				Message: fmt.Sprintf("MISMATCH TABLE %s(GIVEN): column %s nullable mismatch", table, col.Name),
				Type:    "MISMATCH",
				Label:   "WARNING",
			})
		}
	}

	return issues
}
