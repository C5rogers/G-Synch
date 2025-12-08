package pg

import "github.com/C5rogers/G-Synch/internal/audit/core"

func (p *PostgresAdapter) GetColumns(table string) ([]core.Column, error) {
	// TODO: implement to select columns of this query:
	// 	SELECT
	//     column_name,
	//     data_type,
	//     is_nullable,
	//     column_default
	// FROM information_schema.columns
	// WHERE table_schema = 'public' AND table_name = 'expected_budget_target_programs'
	// ORDER BY ordinal_position;
	return nil, nil
}

func (p *PostgresAdapter) GetForeignKeys(*core.Table) ([]core.ForeignKey, error) {
	// TODO: implement to select foreign keys of this query:
	// 	SELECT
	//     tc.table_name AS table_name,
	//     kcu.column_name AS column_name,
	//     ccu.table_name AS foreign_table,
	//     ccu.column_name AS foreign_column
	// FROM
	//     information_schema.table_constraints AS tc
	// JOIN information_schema.key_column_usage AS kcu
	//     ON tc.constraint_name = kcu.constraint_name
	// JOIN information_schema.constraint_column_usage AS ccu
	//     ON ccu.constraint_name = tc.constraint_name
	// WHERE
	//     tc.constraint_type = 'FOREIGN KEY' AND
	//     tc.table_schema = 'public';
	return nil, nil
}

func (p *PostgresAdapter) GetPrimaryKeys(*core.Table) ([]string, error) {
	// TODO: implement to select primary keys of this query:
	// 	SELECT
	//     kcu.column_name
	// FROM information_schema.table_constraints tc
	// JOIN information_schema.key_column_usage kcu
	//     ON tc.constraint_name = kcu.constraint_name
	//     AND tc.table_schema = kcu.table_schema
	// WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = 'expected_budget_target_programs';
	return nil, nil
}

func (p *PostgresAdapter) CopyTableData(srcDSN, dstDSN, table string) error {
	// TODO: implement to copy table data from srcDSN to dstDSN
	// 	COPY table_name TO 'another table';
	return nil
}
