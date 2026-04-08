package tests

import (
	"context"
	"testing"

	"github.com/C5rogers/G-Synch/internal/audit"
	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/stretchr/testify/assert"
)

type mockSchemaAdapter struct {
	schema *core.Schema
}

func (m mockSchemaAdapter) LoadSchema(ctx context.Context, dsn string) (*core.Schema, error) {
	return m.schema, nil
}

func (m mockSchemaAdapter) ListSchemas(ctx context.Context) ([]string, error) {
	return []string{"public"}, nil
}

func (m mockSchemaAdapter) GetColumns(ctx context.Context, dsn string, table *core.Table) ([]core.Column, error) {
	return nil, nil
}

func (m mockSchemaAdapter) GetForeignKeys(ctx context.Context, dsn string, table *core.Table) ([]core.ForeignKey, error) {
	return nil, nil
}

func (m mockSchemaAdapter) GetPrimaryKeys(ctx context.Context, dsn string, table *core.Table) ([]string, error) {
	return nil, nil
}

func (m mockSchemaAdapter) CopyTableData(ctx context.Context, srcDSN, dstDSN, table string) error {
	return nil
}

func (m mockSchemaAdapter) MigrateMissingRowsFrom(ctx context.Context, source core.SchemaAdapter, schema string, table core.Table) (int, int, []string, error) {
	return 0, 0, nil, nil
}

func (m mockSchemaAdapter) GetPrimaryKeyValues(ctx context.Context, dsn, table string) ([][]interface{}, error) {
	return nil, nil
}

func (m mockSchemaAdapter) GetUnsyncedPrimaryKeyValues(ctx context.Context, dsn, table string) ([]string, error) {
	return nil, nil
}

func (m mockSchemaAdapter) CreateTemporaryTable(ctx context.Context) error {
	return nil
}

func (m mockSchemaAdapter) CreateTempRecords(ctx context.Context, values []string) (int64, error) {
	return 0, nil
}

func (m mockSchemaAdapter) TruncateTemporaryTable(ctx context.Context) error {
	return nil
}

func (m mockSchemaAdapter) Engine() string {
	return "postgres"
}

func TestSerializeRow(t *testing.T) {
	row := []interface{}{"abc", 123, true}
	result := audit.SerializeRow(row)
	assert.Equal(t, "abc::123::true", result)
}

func TestSerializeRow_SingleValue(t *testing.T) {
	row := []interface{}{42}
	result := audit.SerializeRow(row)
	assert.Equal(t, "42", result)
}

func TestSerializeRow_Empty(t *testing.T) {
	row := []interface{}{}
	result := audit.SerializeRow(row)
	assert.Equal(t, "", result)
}

func TestSerializeRow_NilValues(t *testing.T) {
	row := []interface{}{nil, "text", nil}
	result := audit.SerializeRow(row)
	assert.Equal(t, "<nil>::text::<nil>", result)
}

func TestSerializeRow_UUIDArray(t *testing.T) {
	row := []interface{}{[16]uint8{118, 88, 77, 15, 84, 235, 67, 211, 187, 100, 140, 155, 227, 6, 185, 238}, "value"}
	result := audit.SerializeRow(row)
	assert.Equal(t, "76584d0f-54eb-43d3-bb64-8c9be306b9ee::value", result)
}

func TestCompareColumns_Identical(t *testing.T) {
	target := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: true},
		},
	}
	given := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: true},
		},
	}

	issues := audit.CompareColumns("users", target, given)
	assert.Empty(t, issues)
}

func TestCompareColumns_MissingColumn(t *testing.T) {
	target := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: true},
		},
	}
	given := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
	}

	issues := audit.CompareColumns("users", target, given)
	assert.Len(t, issues, 1)
	assert.Equal(t, "MISSING", issues[0].Type)
	assert.Contains(t, issues[0].Message, "email")
}

func TestCompareColumns_DataTypeMismatch(t *testing.T) {
	target := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "age", DataType: "integer", IsNullable: false},
		},
	}
	given := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "age", DataType: "text", IsNullable: false},
		},
	}

	issues := audit.CompareColumns("users", target, given)
	assert.Len(t, issues, 1)
	assert.Equal(t, "MISMATCH", issues[0].Type)
	assert.Contains(t, issues[0].Message, "integer")
	assert.Contains(t, issues[0].Message, "text")
}

func TestCompareColumns_NullabilityMismatch(t *testing.T) {
	target := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "name", DataType: "text", IsNullable: false},
		},
	}
	given := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "name", DataType: "text", IsNullable: true},
		},
	}

	issues := audit.CompareColumns("users", target, given)
	assert.Len(t, issues, 1)
	assert.Equal(t, "MISMATCH", issues[0].Type)
	assert.Contains(t, issues[0].Message, "nullable")
}

func TestCompareColumns_MultipleIssues(t *testing.T) {
	target := core.Table{
		Name: "products",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "price", DataType: "numeric", IsNullable: false},
			{Name: "sku", DataType: "varchar", IsNullable: true},
		},
	}
	given := core.Table{
		Name: "products",
		Columns: []core.Column{
			{Name: "id", DataType: "bigint", IsNullable: false},    // type mismatch
			{Name: "price", DataType: "numeric", IsNullable: true}, // nullable mismatch
			// sku is missing
		},
	}

	issues := audit.CompareColumns("products", target, given)
	assert.Len(t, issues, 3)
}

func TestCompareColumns_ExtraColumnsInGivenIgnored(t *testing.T) {
	target := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
	}
	given := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "extra", DataType: "text", IsNullable: true},
		},
	}

	issues := audit.CompareColumns("users", target, given)
	assert.Empty(t, issues)
}

func TestSchemaAudit_Name(t *testing.T) {
	a := &audit.SchemaAudit{}
	assert.Equal(t, "schema-check", a.Name())
}

func TestSchemaAudit_Check_SkipsKeylessTables(t *testing.T) {
	schema := &core.Schema{
		Name: "public",
		Tables: []core.Table{
			{
				Name:    "job_applications",
				Columns: []core.Column{{Name: "id", DataType: "integer", IsNullable: false}},
			},
		},
	}

	auditor := &audit.SchemaAudit{}
	warnings, err := auditor.Check(context.Background(), mockSchemaAdapter{schema: schema}, mockSchemaAdapter{schema: schema}, "public")
	assert.NoError(t, err)
	assert.NotEmpty(t, warnings)
	assert.Contains(t, warnings[0].Message, "no primary key")
	assert.Equal(t, "UNSUPPORTED", warnings[0].Type)
}
