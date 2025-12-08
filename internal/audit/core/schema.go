package core

type Column struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
}

type Table struct {
	Name        string
	Schema      Schema
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
	Tables []Table
}
