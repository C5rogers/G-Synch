package core

type SchemaAdapter interface {
	LoadSchema(dsn string) (*Schema, error)
	GetColumns(*Table) ([]Column, error)
	GetForeignKeys(*Table) ([]ForeignKey, error)
	GetPrimaryKeys(*Table) ([]string, error)
	CopyTableData(srcDSN, dstDSN, table string) error
}
