package sync

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/C5rogers/G-Synch/internal/audit/adapters/pg"
	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sync struct {
	GivenDB  *pgxpool.Pool
	TargetDB *pgxpool.Pool
}

func NewSyncAPI(GivenDB, TargetDB *pgxpool.Pool) (*Sync, error) {
	s := &Sync{
		GivenDB:  GivenDB,
		TargetDB: TargetDB,
	}

	return s, nil
}

func (s *Sync) Synch(targetDB string, givenDB string, activityID *string, activityType *string, schema string, logToFile bool) {
	var writer *bufio.Writer
	if logToFile && activityID != nil && activityType != nil {
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Fatal(err)
		}
		file, err := os.Create("logs/" + *activityID + "_" + *activityType + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(file)
	} else if logToFile {
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Fatal(err)
		}
		file, err := os.Create("logs/audit_synch_" + time.Now().Format("20060102150405") + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(file)
	}

	Logf(writer, "Synchronization started between %s (target/source) and %s (given/destination) of %s schema", targetDB, givenDB, schema)

	var targetAdapter core.SchemaAdapter = pg.New(s.TargetDB)
	var givenAdapter core.SchemaAdapter = pg.New(s.GivenDB)
	ctx := context.Background()

	targetSchema, err := targetAdapter.LoadSchema(ctx, schema)
	if err != nil {
		Logf(writer, "Error loading target schema: %v", err)
		FlushWriter(writer)
		return
	}

	givenSchema, err := givenAdapter.LoadSchema(ctx, schema)
	if err != nil {
		Logf(writer, "Error loading given schema: %v", err)
		FlushWriter(writer)
		return
	}

	// Sort tables by FK dependency order so parent tables are migrated first.
	sortedTables, sortErr := core.TopologicalSortTables(targetSchema.Tables)
	if sortErr != nil {
		Logf(writer, "WARNING: Could not resolve table dependency order: %v", sortErr)
		Logf(writer, "Falling back to unsorted table order.")
		sortedTables = targetSchema.Tables
	} else {
		Logf(writer, "Tables sorted by FK dependency order (%d tables)", len(sortedTables))
	}

	givenTables := make(map[string]core.Table, len(givenSchema.Tables))
	for _, table := range givenSchema.Tables {
		givenTables[strings.ToLower(table.Name)] = table
	}

	totalMigrated := 0
	totalDenied := 0
	totalSkipped := 0
	deniedByTable := map[string][]string{}

	for _, sourceTable := range sortedTables {
		tableName := sourceTable.Name
		if strings.EqualFold(tableName, "compare_table") {
			continue
		}

		destinationTable, exists := givenTables[strings.ToLower(tableName)]
		if !exists {
			totalSkipped++
			Logf(writer, "SKIP TABLE %s: missing in given database", tableName)
			continue
		}

		if !SchemaCompatible(sourceTable, destinationTable) {
			totalSkipped++
			Logf(writer, "SKIP TABLE %s: schema mismatch (columns/types/nullability/primary keys differ)", tableName)
			continue
		}

		migrated, denied, deniedPKs, migrateErr := givenAdapter.MigrateMissingRowsFrom(ctx, targetAdapter, schema, sourceTable)
		if migrateErr != nil {
			totalSkipped++
			Logf(writer, "SKIP TABLE %s: error during migration planning/execution: %v", tableName, migrateErr)
			continue
		}

		totalMigrated += migrated
		totalDenied += denied
		if len(deniedPKs) > 0 {
			deniedByTable[tableName] = append(deniedByTable[tableName], deniedPKs...)
		}

		Logf(writer, "TABLE %s: migrated=%d denied=%d", tableName, migrated, denied)
	}

	Logf(writer, "Synchronization completed.")
	Logf(writer, "Summary: migrated=%d denied=%d skipped_tables=%d", totalMigrated, totalDenied, totalSkipped)

	if len(deniedByTable) > 0 {
		tables := make([]string, 0, len(deniedByTable))
		for tableName := range deniedByTable {
			tables = append(tables, tableName)
		}
		sort.Strings(tables)

		for _, tableName := range tables {
			pkRefs := deniedByTable[tableName]
			sort.Strings(pkRefs)
			Logf(writer, "Denied PK references in given DB for table %s: %s", tableName, strings.Join(pkRefs, ", "))
		}
	}

	FlushWriter(writer)
	time.Sleep(2 * time.Second)
}

func SchemaCompatible(sourceTable core.Table, destinationTable core.Table) bool {
	if len(sourceTable.Columns) != len(destinationTable.Columns) {
		return false
	}

	destinationCols := map[string]core.Column{}
	for _, col := range destinationTable.Columns {
		destinationCols[strings.ToLower(col.Name)] = col
	}

	for _, sourceCol := range sourceTable.Columns {
		destinationCol, ok := destinationCols[strings.ToLower(sourceCol.Name)]
		if !ok {
			return false
		}
		if sourceCol.DataType != destinationCol.DataType {
			return false
		}
		if sourceCol.IsNullable != destinationCol.IsNullable {
			return false
		}
	}

	if len(sourceTable.PrimaryKey) != len(destinationTable.PrimaryKey) {
		return false
	}

	destinationPKSet := map[string]struct{}{}
	for _, pk := range destinationTable.PrimaryKey {
		destinationPKSet[strings.ToLower(pk)] = struct{}{}
	}

	for _, sourcePK := range sourceTable.PrimaryKey {
		if _, ok := destinationPKSet[strings.ToLower(sourcePK)]; !ok {
			return false
		}
	}

	return true
}

func Logf(writer *bufio.Writer, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if writer != nil {
		fmt.Fprintf(writer, "%s\n", message)
		return
	}
	fmt.Println(message)
}

func FlushWriter(writer *bufio.Writer) {
	if writer != nil {
		_ = writer.Flush()
	}
}
