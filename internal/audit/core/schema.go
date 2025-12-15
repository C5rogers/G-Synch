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

type SchemaAdapter interface {
	LoadSchema(ctx context.Context, dsn string) (*Schema, error)
	GetColumns(ctx context.Context, dsn string, table *Table) ([]Column, error)
	GetForeignKeys(ctx context.Context, dsn string, table *Table) ([]ForeignKey, error)
	GetPrimaryKeys(ctx context.Context, dsn string, table *Table) ([]string, error)
	CopyTableData(ctx context.Context, srcDSN, dstDSN, table string) error
	GetPrimaryKeyValues(ctx context.Context, dsn, table string) ([][]interface{}, error)
	Engine() string
}
