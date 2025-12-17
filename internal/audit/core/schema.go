package core

import "context"

type Column struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
}

type Table struct {
	Name        string
	Columns     []Column
	PrimaryKey  []string
	ForeignKeys []ForeignKey
}

type ForeignKey struct {
	Column           string
	ReferencedTable  string
	ReferencedColumn string
}
type Schema struct {
	Name   string
	Tables []Table
}

// CopyTableData implements [SchemaAdapter].
func (s *Schema) CopyTableData(ctx context.Context, srcDSN string, dstDSN string, table string) error {
	panic("unimplemented")
}

// Engine implements [SchemaAdapter].
func (s *Schema) Engine() string {
	panic("unimplemented")
}

// GetColumns implements [SchemaAdapter].
func (s *Schema) GetColumns(ctx context.Context, dsn string, table *Table) ([]Column, error) {
	panic("unimplemented")
}

// GetForeignKeys implements [SchemaAdapter].
func (s *Schema) GetForeignKeys(ctx context.Context, dsn string, table *Table) ([]ForeignKey, error) {
	panic("unimplemented")
}

// GetPrimaryKeyValues implements [SchemaAdapter].
func (s *Schema) GetPrimaryKeyValues(ctx context.Context, dsn string, table string) ([][]interface{}, error) {
	panic("unimplemented")
}

// GetPrimaryKeys implements [SchemaAdapter].
func (s *Schema) GetPrimaryKeys(ctx context.Context, dsn string, table *Table) ([]string, error) {
	panic("unimplemented")
}

// LoadSchema implements [SchemaAdapter].
func (s *Schema) LoadSchema(ctx context.Context, dsn string) (*Schema, error) {
	panic("unimplemented")
}

type SchemaAdapter interface {
	LoadSchema(ctx context.Context, dsn string) (*Schema, error)
	GetColumns(ctx context.Context, dsn string, table *Table) ([]Column, error)
	GetForeignKeys(ctx context.Context, dsn string, table *Table) ([]ForeignKey, error)
	GetPrimaryKeys(ctx context.Context, dsn string, table *Table) ([]string, error)
	CopyTableData(ctx context.Context, srcDSN, dstDSN, table string) error
	GetPrimaryKeyValues(ctx context.Context, dsn, table string) ([][]interface{}, error)
	Engine() string
}
