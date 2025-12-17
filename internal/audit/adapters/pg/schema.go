package pg

import (
	"context"
	"fmt"

	"github.com/C5rogers/G-Synch/internal/audit/core"
	pg_db "github.com/C5rogers/G-Synch/internal/audit/engines/pg/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func (p *Adapter) LoadSchema(dsn string) (*core.Schema, error) {

	queries := pg_db.New(p.db)
	ctx := context.Background()

	var tables []interface{}
	tables, err := queries.LoadSchema(ctx, pgtype.Text{String: dsn, Valid: true})
	if err != nil {
		// here it is the error for database connection
		fmt.Println("We are here:", err)
		return nil, err
	}
	schema := &core.Schema{
		Name:   dsn,
		Tables: make([]core.Table, 0, len(tables)),
	}
	for _, t := range tables {
		tableName := t.(string)
		table := core.Table{
			Name:        tableName,
			Columns:     []core.Column{},
			PrimaryKey:  []string{},
			ForeignKeys: []core.ForeignKey{},
		}
		// for each table get the columns, primary keys, foreign keys
		columns, err := queries.GetColumns(ctx, pg_db.GetColumnsParams{
			SchemaName: pgtype.Text{String: dsn, Valid: true},
			TableName:  pgtype.Text{String: tableName, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		for _, c := range columns {
			column := core.Column{
				Name:         c.ColumnName.(string),
				DataType:     c.DataType.(string),
				IsNullable:   c.IsNullable.(string) == "YES",
				DefaultValue: c.ColumnDefault.(*string),
			}
			table.Columns = append(table.Columns, column)
		}
		// for each table get foreign keys
		foreignKeys, err := queries.GetForeignKeys(ctx, pg_db.GetForeignKeysParams{
			SchemaName: pgtype.Text{String: dsn, Valid: true},
			TableName:  pgtype.Text{String: tableName, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		for _, fk := range foreignKeys {
			foreignKey := core.ForeignKey{
				Column:           fk.ColumnName.(string),
				ReferencedTable:  fk.ForeignTableName.(string),
				ReferencedColumn: fk.ForeignColumnName.(string),
			}
			table.ForeignKeys = append(table.ForeignKeys, foreignKey)
		}
		primaryKeys, err := queries.GetPrimaryKeys(ctx, pg_db.GetPrimaryKeysParams{
			SchemaName: pgtype.Text{String: dsn, Valid: true},
			TableName:  pgtype.Text{String: tableName, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		for _, pk := range primaryKeys {
			pkName := pk.(string)
			table.PrimaryKey = append(table.PrimaryKey, pkName)
		}
		schema.Tables = append(schema.Tables, table)
	}

	return schema, nil
}
