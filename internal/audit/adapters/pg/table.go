package pg

import (
	"context"

	"github.com/C5rogers/G-Synch/internal/audit/core"
	pg_db "github.com/C5rogers/G-Synch/internal/audit/engines/pg/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func (p *Adapter) GetColumns(dsn string, table *core.Table) ([]core.Column, error) {

	queries := pg_db.New(p.db)
	ctx := context.Background()

	columns, err := queries.GetColumns(ctx, pg_db.GetColumnsParams{
		SchemaName: pgtype.Text{String: dsn, Valid: true},
		TableName:  pgtype.Text{String: table.Name, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	var cols []core.Column
	for _, c := range columns {
		column := core.Column{
			Name:         c.ColumnName.(string),
			DataType:     c.DataType.(string),
			IsNullable:   c.IsNullable.(string) == "YES",
			DefaultValue: c.ColumnDefault.(*string),
		}
		cols = append(cols, column)
	}
	return cols, nil
}

func (p *Adapter) GetForeignKeys(dsn string, table *core.Table) ([]core.ForeignKey, error) {

	queries := pg_db.New(p.db)
	ctx := context.Background()
	foreignKeys, err := queries.GetForeignKeys(ctx, pg_db.GetForeignKeysParams{
		SchemaName: pgtype.Text{String: dsn, Valid: true},
		TableName:  pgtype.Text{String: table.Name, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	var fks []core.ForeignKey
	for _, fk := range foreignKeys {
		foreignKey := core.ForeignKey{
			Column:           fk.ColumnName.(string),
			ReferencedTable:  fk.ForeignTableName.(string),
			ReferencedColumn: fk.ForeignColumnName.(string),
		}
		fks = append(fks, foreignKey)
	}
	return fks, nil
}

func (p *Adapter) GetPrimaryKeys(dsn string, table *core.Table) ([]string, error) {

	queries := pg_db.New(p.db)
	ctx := context.Background()

	primaryKeys, err := queries.GetPrimaryKeys(ctx, pg_db.GetPrimaryKeysParams{
		SchemaName: pgtype.Text{String: dsn, Valid: true},
		TableName:  pgtype.Text{String: table.Name, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	var pks []string
	for _, pk := range primaryKeys {
		pkName := pk.(string)
		pks = append(pks, pkName)
	}
	return pks, nil
}

func (p *Adapter) CopyTableData(srcDSN, dstDSN, table string) error {
	// TODO: implement to copy table data from srcDSN to dstDSN
	// 	COPY table_name TO 'another table';
	return nil
}
