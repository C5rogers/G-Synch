package tests

import (
	"testing"

	"github.com/C5rogers/G-Synch/internal/audit"
	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/stretchr/testify/assert"
)

func TestMapTables(t *testing.T) {
	tables := []core.Table{
		{Name: "Users"},
		{Name: "Orders"},
		{Name: "PRODUCTS"},
	}

	result := audit.MapTables(tables)

	assert.Len(t, result, 3)
	assert.Equal(t, "Users", result["users"].Name)
	assert.Equal(t, "Orders", result["orders"].Name)
	assert.Equal(t, "PRODUCTS", result["products"].Name)
}

func TestMapTables_Empty(t *testing.T) {
	result := audit.MapTables([]core.Table{})
	assert.Len(t, result, 0)
}

func TestMapTables_DuplicateNames(t *testing.T) {
	tables := []core.Table{
		{Name: "users", Columns: []core.Column{{Name: "first"}}},
		{Name: "Users", Columns: []core.Column{{Name: "second"}}},
	}

	result := audit.MapTables(tables)
	assert.Len(t, result, 1)
	assert.Equal(t, "Users", result["users"].Name)
}

func TestTableExists(t *testing.T) {
	schema := &core.Schema{
		Tables: []core.Table{
			{Name: "users"},
			{Name: "orders"},
		},
	}

	assert.True(t, audit.TableExists(schema, "users"))
	assert.True(t, audit.TableExists(schema, "orders"))
	assert.False(t, audit.TableExists(schema, "products"))
	assert.False(t, audit.TableExists(schema, "Users"))
}

func TestTableExists_EmptySchema(t *testing.T) {
	schema := &core.Schema{Tables: []core.Table{}}
	assert.False(t, audit.TableExists(schema, "anything"))
}

func TestMapColumns(t *testing.T) {
	cols := []core.Column{
		{Name: "ID", DataType: "integer", IsNullable: false},
		{Name: "Email", DataType: "varchar", IsNullable: true},
		{Name: "NAME", DataType: "text", IsNullable: false},
	}

	result := audit.MapColumns(cols)

	assert.Len(t, result, 3)
	assert.Equal(t, "ID", result["id"].Name)
	assert.Equal(t, "integer", result["id"].DataType)
	assert.Equal(t, "Email", result["email"].Name)
	assert.True(t, result["email"].IsNullable)
	assert.Equal(t, "NAME", result["name"].Name)
}

func TestMapColumns_Empty(t *testing.T) {
	result := audit.MapColumns([]core.Column{})
	assert.Len(t, result, 0)
}
