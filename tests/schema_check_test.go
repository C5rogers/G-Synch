package tests

import (
	"testing"

	"github.com/C5rogers/G-Synch/internal/audit"
	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/stretchr/testify/assert"
)

func TestSerializeRow(t *testing.T) {
	row := []interface{}{"abc", 123, true}
	result := audit.SerializeRow(row)
	assert.Equal(t, "abc|123|true", result)
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
	assert.Equal(t, "<nil>|text|<nil>", result)
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
			{Name: "id", DataType: "bigint", IsNullable: false}, // type mismatch
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
