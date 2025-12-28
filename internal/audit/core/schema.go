package core

import (
	"context"
)

/*
 * Column represents a column in a database table.
 */
type Column struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
}

/*
 * Table represents a database table with its columns, primary keys, and foreign keys.
 */
type Table struct {
	Name        string
	Columns     []Column
	PrimaryKey  []string
	ForeignKeys []ForeignKey
}

/*
 * ForeignKey represents a foreign key relationship in a database table.
 */
type ForeignKey struct {
	Column                string
	ReferencedTable       string
	ReferencedColumn      string
	ReferencedTableSchema string
}

/*
 * Schema represents a database schema containing multiple tables.
 */
type Schema struct {
	Name   string
	Tables []Table
}

/*
 * AuditResult represents the result of an audit operation.
 */
type SchemaAdapter interface {
	/*
	 * LoadSchema loads the database schema from the given DSN.
	 */
	LoadSchema(ctx context.Context, dsn string) (*Schema, error)
	/*
	 * GetColumns retrieves the columns of a specified table in the schema.
	 */
	GetColumns(ctx context.Context, dsn string, table *Table) ([]Column, error)
	/*
	 * GetForeignKeys retrieves the foreign keys of a specified table in the schema.
	 */
	GetForeignKeys(ctx context.Context, dsn string, table *Table) ([]ForeignKey, error)
	/*
	 * GetPrimaryKeys retrieves the primary keys of a specified table in the schema.
	 */
	GetPrimaryKeys(ctx context.Context, dsn string, table *Table) ([]string, error)
	/*
	 * CopyTableData copies data from one table to another.
	 */
	CopyTableData(ctx context.Context, srcDSN, dstDSN, table string) error
	/*
	 * GetPrimaryKeyValues retrieves the primary key values of a specified table.
	 */
	GetPrimaryKeyValues(ctx context.Context, dsn, table string) ([][]interface{}, error)
	/*
	 * GetUnsyncedPrimaryKeyValues retrieves the unsynced primary key values of a specified table.
	 */
	GetUnsyncedPrimaryKeyValues(ctx context.Context, dsn, table string) ([]string, error)
	/*
	 * CreateTemporaryTable creates a temporary table for staging data.
	 */
	CreateTemporaryTable(ctx context.Context) error
	/*
	 * CreateTempRecords creates temporary records in the temporary table.
	 */
	CreateTempRecords(ctx context.Context, values []string) (int64, error)
	/*
	 * TruncateTemporaryTable truncates the temporary table.
	 */
	TruncateTemporaryTable(ctx context.Context) error
	/*
	 * Engine returns the database engine being used.
	 */
	Engine() string
}
