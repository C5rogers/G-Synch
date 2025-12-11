package core

type SchemaAdapter interface {
	LoadSchema(dsn string) (*Schema, error)
	GetColumns(dsn string, table *Table) ([]Column, error)
	GetForeignKeys(dsn string, table *Table) ([]ForeignKey, error)
	GetPrimaryKeys(dsn string, table *Table) ([]string, error)
	CopyTableData(srcDSN, dstDSN, table string) error
}
