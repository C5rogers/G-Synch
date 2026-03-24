package tests

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/C5rogers/G-Synch/pkg/sync"
	"github.com/stretchr/testify/assert"
)

func TestSchemaCompatible_Identical(t *testing.T) {
	source := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: true},
		},
		PrimaryKey: []string{"id"},
	}
	dest := core.Table{
		Name: "users",
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: true},
		},
		PrimaryKey: []string{"id"},
	}

	assert.True(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_DifferentColumnCount(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: true},
		},
		PrimaryKey: []string{"id"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}

	assert.False(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_MissingColumn(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "name", DataType: "text", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "text", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}

	assert.False(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_DataTypeMismatch(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "bigint", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}

	assert.False(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_NullabilityMismatch(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: true},
		},
		PrimaryKey: []string{"id"},
	}

	assert.False(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_DifferentPKCount(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"id", "name"},
	}

	assert.False(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_DifferentPKColumns(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "name", DataType: "text", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "name", DataType: "text", IsNullable: false},
		},
		PrimaryKey: []string{"name"},
	}

	assert.False(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_CompositePK(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "user_id", DataType: "integer", IsNullable: false},
			{Name: "role_id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"user_id", "role_id"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "user_id", DataType: "integer", IsNullable: false},
			{Name: "role_id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"role_id", "user_id"},
	}

	assert.True(t, sync.SchemaCompatible(source, dest))
}

func TestSchemaCompatible_CaseInsensitiveMatch(t *testing.T) {
	source := core.Table{
		Columns: []core.Column{
			{Name: "ID", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"ID"},
	}
	dest := core.Table{
		Columns: []core.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: []string{"id"},
	}

	assert.True(t, sync.SchemaCompatible(source, dest))
}

func TestLogf_WithWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	sync.Logf(writer, "hello %s %d", "world", 42)
	writer.Flush()

	assert.Equal(t, "hello world 42\n", buf.String())
}

func TestLogf_NilWriter(t *testing.T) {
	sync.Logf(nil, "test %s", "message")
}

func TestFlushWriter_NilWriter(t *testing.T) {
	sync.FlushWriter(nil)
}

func TestFlushWriter_WithWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	writer.WriteString("test")
	sync.FlushWriter(writer)
	assert.Equal(t, "test", buf.String())
}
