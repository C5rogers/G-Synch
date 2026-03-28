package tests

import (
	"testing"

	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/stretchr/testify/assert"
)

func TestTopologicalSort_LinearChain(t *testing.T) {
	// A → B → C: insert A first, then B, then C
	tables := []core.Table{
		{Name: "C", ForeignKeys: []core.ForeignKey{{Column: "b_id", ReferencedTable: "B", ReferencedColumn: "id"}}},
		{Name: "A", ForeignKeys: []core.ForeignKey{}},
		{Name: "B", ForeignKeys: []core.ForeignKey{{Column: "a_id", ReferencedTable: "A", ReferencedColumn: "id"}}},
	}

	sorted, err := core.TopologicalSortTables(tables)
	assert.NoError(t, err)
	assert.Len(t, sorted, 3)

	indexOf := func(name string) int {
		for i, t := range sorted {
			if t.Name == name {
				return i
			}
		}
		return -1
	}

	assert.Less(t, indexOf("A"), indexOf("B"), "A must come before B")
	assert.Less(t, indexOf("B"), indexOf("C"), "B must come before C")
}

func TestTopologicalSort_DiamondDependency(t *testing.T) {
	// A ← B, A ← C, B ← D, C ← D
	tables := []core.Table{
		{Name: "D", ForeignKeys: []core.ForeignKey{
			{Column: "b_id", ReferencedTable: "B", ReferencedColumn: "id"},
			{Column: "c_id", ReferencedTable: "C", ReferencedColumn: "id"},
		}},
		{Name: "B", ForeignKeys: []core.ForeignKey{{Column: "a_id", ReferencedTable: "A", ReferencedColumn: "id"}}},
		{Name: "C", ForeignKeys: []core.ForeignKey{{Column: "a_id", ReferencedTable: "A", ReferencedColumn: "id"}}},
		{Name: "A", ForeignKeys: []core.ForeignKey{}},
	}

	sorted, err := core.TopologicalSortTables(tables)
	assert.NoError(t, err)
	assert.Len(t, sorted, 4)

	indexOf := func(name string) int {
		for i, t := range sorted {
			if t.Name == name {
				return i
			}
		}
		return -1
	}

	assert.Less(t, indexOf("A"), indexOf("B"), "A must come before B")
	assert.Less(t, indexOf("A"), indexOf("C"), "A must come before C")
	assert.Less(t, indexOf("B"), indexOf("D"), "B must come before D")
	assert.Less(t, indexOf("C"), indexOf("D"), "C must come before D")
}

func TestTopologicalSort_NoDependencies(t *testing.T) {
	tables := []core.Table{
		{Name: "products", ForeignKeys: []core.ForeignKey{}},
		{Name: "categories", ForeignKeys: []core.ForeignKey{}},
		{Name: "tags", ForeignKeys: []core.ForeignKey{}},
	}

	sorted, err := core.TopologicalSortTables(tables)
	assert.NoError(t, err)
	assert.Len(t, sorted, 3)
}

func TestTopologicalSort_CycleDetection(t *testing.T) {
	// A → B → A (cycle)
	tables := []core.Table{
		{Name: "A", ForeignKeys: []core.ForeignKey{{Column: "b_id", ReferencedTable: "B", ReferencedColumn: "id"}}},
		{Name: "B", ForeignKeys: []core.ForeignKey{{Column: "a_id", ReferencedTable: "A", ReferencedColumn: "id"}}},
	}

	_, err := core.TopologicalSortTables(tables)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular foreign key dependency")
}

func TestTopologicalSort_SelfReference(t *testing.T) {
	// Table referencing itself (e.g., categories with parent_id)
	tables := []core.Table{
		{Name: "categories", ForeignKeys: []core.ForeignKey{
			{Column: "parent_id", ReferencedTable: "categories", ReferencedColumn: "id"},
		}},
		{Name: "products", ForeignKeys: []core.ForeignKey{
			{Column: "category_id", ReferencedTable: "categories", ReferencedColumn: "id"},
		}},
	}

	sorted, err := core.TopologicalSortTables(tables)
	assert.NoError(t, err)
	assert.Len(t, sorted, 2)

	// categories must come before products
	assert.Equal(t, "categories", sorted[0].Name)
	assert.Equal(t, "products", sorted[1].Name)
}

func TestTopologicalSort_CrossSchemaFKIgnored(t *testing.T) {
	// FK to a table not in the provided set is ignored
	tables := []core.Table{
		{Name: "orders", ForeignKeys: []core.ForeignKey{
			{Column: "customer_id", ReferencedTable: "external_customers", ReferencedColumn: "id", ReferencedTableSchema: "other_schema"},
		}},
		{Name: "order_items", ForeignKeys: []core.ForeignKey{
			{Column: "order_id", ReferencedTable: "orders", ReferencedColumn: "id"},
		}},
	}

	sorted, err := core.TopologicalSortTables(tables)
	assert.NoError(t, err)
	assert.Len(t, sorted, 2)
	assert.Equal(t, "orders", sorted[0].Name)
	assert.Equal(t, "order_items", sorted[1].Name)
}

func TestTopologicalSort_UsersAndUserRoles(t *testing.T) {
	// Real-world scenario: users and users_role
	tables := []core.Table{
		{Name: "users_role", ForeignKeys: []core.ForeignKey{
			{Column: "user_id", ReferencedTable: "users", ReferencedColumn: "id"},
			{Column: "role_id", ReferencedTable: "roles", ReferencedColumn: "id"},
		}},
		{Name: "roles", ForeignKeys: []core.ForeignKey{}},
		{Name: "users", ForeignKeys: []core.ForeignKey{}},
	}

	sorted, err := core.TopologicalSortTables(tables)
	assert.NoError(t, err)
	assert.Len(t, sorted, 3)

	indexOf := func(name string) int {
		for i, t := range sorted {
			if t.Name == name {
				return i
			}
		}
		return -1
	}

	assert.Less(t, indexOf("users"), indexOf("users_role"), "users must come before users_role")
	assert.Less(t, indexOf("roles"), indexOf("users_role"), "roles must come before users_role")
}

func TestTopologicalSort_EmptyInput(t *testing.T) {
	sorted, err := core.TopologicalSortTables([]core.Table{})
	assert.NoError(t, err)
	assert.Len(t, sorted, 0)
}
