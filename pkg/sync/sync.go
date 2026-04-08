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
	"github.com/jackc/pgx/v5"
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
	loader := NewLoader("Reading database metadata and synchronizing rows...")
	loader.Start()
	defer loader.Stop()

	var writer *bufio.Writer
	if logToFile && activityID != nil && activityType != nil {
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Printf("failed to create logs directory: %v", err)
			return
		}
		file, err := os.Create("logs/" + *activityID + "_" + *activityType + ".txt")
		if err != nil {
			log.Printf("failed to create log file for activity %s (%s): %v", *activityID, *activityType, err)
			return
		}
		defer file.Close()
		writer = bufio.NewWriter(file)
	} else if logToFile {
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Printf("failed to create logs directory: %v", err)
			return
		}
		file, err := os.Create("logs/audit_synch_" + time.Now().Format("20060102150405") + ".txt")
		if err != nil {
			log.Printf("failed to create synch log file: %v", err)
			return
		}
		defer file.Close()
		writer = bufio.NewWriter(file)
		defer writer.Flush()
	}

	Logf(writer, "Synchronization started between %s (target/source) and %s (given/destination) of %s schema", targetDB, givenDB, schema)

	targetAdapter := pg.New(s.TargetDB)
	givenAdapter := pg.New(s.GivenDB)
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

	plan, planErr := core.BuildDependencyPlan(targetSchema.Tables)
	if planErr != nil {
		Logf(writer, "Error building dependency plan: %v", planErr)
		FlushWriter(writer)
		return
	}
	cyclicGroups := 0
	for _, group := range plan {
		if group.Cyclic {
			cyclicGroups++
		}
	}
	Logf(writer, "Dependency plan built with %d groups (%d cyclic)", len(plan), cyclicGroups)

	givenTables := make(map[string]core.Table, len(givenSchema.Tables))
	for _, table := range givenSchema.Tables {
		givenTables[strings.ToLower(table.Name)] = table
	}

	totalMigrated := 0
	totalDenied := 0
	totalSkipped := 0
	deniedByTable := map[string][]string{}

	for _, group := range plan {
		if group.Cyclic {
			migrated, denied, skipped := s.syncCyclicGroup(ctx, writer, targetAdapter, givenAdapter, schema, group, givenTables)
			totalMigrated += migrated
			totalDenied += denied
			totalSkipped += skipped
			continue
		}

		sourceTable := group.Tables[0]
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
}

type foreignKeyDeferrabilityInspector interface {
	ForeignKeyDeferrability(ctx context.Context, schema string, table core.Table) (map[string]bool, error)
}

func (s *Sync) syncCyclicGroup(ctx context.Context, writer *bufio.Writer, targetAdapter core.SchemaAdapter, givenAdapter *pg.Adapter, schema string, group core.DependencyGroup, givenTables map[string]core.Table) (int, int, int) {
	groupTableNames := make([]string, 0, len(group.Tables))
	for _, table := range group.Tables {
		groupTableNames = append(groupTableNames, table.Name)
	}
	sort.Strings(groupTableNames)
	Logf(writer, "CYCLIC GROUP: tables=%s", strings.Join(groupTableNames, ", "))

	schemaIssues := make([]string, 0)
	for _, table := range group.Tables {
		destinationTable, exists := givenTables[strings.ToLower(table.Name)]
		if !exists {
			schemaIssues = append(schemaIssues, fmt.Sprintf("table %s missing in given database", table.Name))
			continue
		}
		if !SchemaCompatible(table, destinationTable) {
			schemaIssues = append(schemaIssues, fmt.Sprintf("table %s schema mismatch (columns/types/nullability/primary keys differ)", table.Name))
		}
	}

	if len(schemaIssues) > 0 {
		sort.Strings(schemaIssues)
		Logf(writer, "SKIP CYCLIC GROUP %s: schema issues detected", strings.Join(groupTableNames, ", "))
		for _, issue := range schemaIssues {
			Logf(writer, "  - %s", issue)
		}
		return 0, 0, len(group.Tables)
	}

	inspector, ok := any(givenAdapter).(foreignKeyDeferrabilityInspector)
	if !ok {
		Logf(writer, "SKIP CYCLIC GROUP %s: given adapter cannot inspect foreign key deferrability", strings.Join(groupTableNames, ", "))
		return 0, 0, len(group.Tables)
	}

	groupSet := make(map[string]struct{}, len(group.Tables))
	for _, table := range group.Tables {
		groupSet[strings.ToLower(table.Name)] = struct{}{}
	}

	blockedReasons := make([]string, 0)
	for _, table := range group.Tables {
		deferrability, err := inspector.ForeignKeyDeferrability(ctx, schema, table)
		if err != nil {
			blockedReasons = append(blockedReasons, fmt.Sprintf("%s: %v", table.Name, err))
			continue
		}
		for _, fk := range table.ForeignKeys {
			if _, inside := groupSet[strings.ToLower(fk.ReferencedTable)]; !inside {
				continue
			}
			if ok, exists := deferrability[strings.ToLower(fk.Column)]; !exists || !ok {
				blockedReasons = append(blockedReasons, fmt.Sprintf("%s.%s -> %s.%s is not deferrable", table.Name, fk.Column, fk.ReferencedTable, fk.ReferencedColumn))
			}
		}
	}

	if len(blockedReasons) > 0 {
		sort.Strings(blockedReasons)
		Logf(writer, "SKIP CYCLIC GROUP %s: cannot defer internal foreign keys", strings.Join(groupTableNames, ", "))
		for _, reason := range blockedReasons {
			Logf(writer, "  - %s", reason)
		}
		return 0, 0, len(group.Tables)
	}

	tx, err := s.GivenDB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		Logf(writer, "SKIP CYCLIC GROUP %s: unable to start transaction: %v", strings.Join(groupTableNames, ", "), err)
		return 0, 0, len(group.Tables)
	}
	txAdapter := pg.New(tx)

	if _, err := tx.Exec(ctx, "SET CONSTRAINTS ALL DEFERRED"); err != nil {
		_ = tx.Rollback(ctx)
		Logf(writer, "SKIP CYCLIC GROUP %s: unable to defer constraints: %v", strings.Join(groupTableNames, ", "), err)
		return 0, 0, len(group.Tables)
	}

	migrated := 0
	denied := 0
	deniedPKs := make([]string, 0)

	for _, table := range group.Tables {
		tableMigrated, tableDenied, tableDeniedPKs, migrateErr := txAdapter.MigrateMissingRowsFrom(ctx, targetAdapter, schema, table)
		if migrateErr != nil {
			_ = tx.Rollback(ctx)
			Logf(writer, "SKIP CYCLIC GROUP %s: error during migration for table %s: %v", strings.Join(groupTableNames, ", "), table.Name, migrateErr)
			return 0, 0, len(group.Tables)
		}
		migrated += tableMigrated
		denied += tableDenied
		deniedPKs = append(deniedPKs, tableDeniedPKs...)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		Logf(writer, "SKIP CYCLIC GROUP %s: commit failed: %v", strings.Join(groupTableNames, ", "), err)
		return 0, 0, len(group.Tables)
	}

	Logf(writer, "CYCLIC GROUP %s: migrated=%d denied=%d", strings.Join(groupTableNames, ", "), migrated, denied)
	if len(deniedPKs) > 0 {
		sort.Strings(deniedPKs)
		Logf(writer, "CYCLIC GROUP %s denied PK references: %s", strings.Join(groupTableNames, ", "), strings.Join(deniedPKs, ", "))
	}

	return migrated, denied, 0
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
	pauseActiveLoader()
	loaderState.stdoutMu.Lock()
	fmt.Println(message)
	loaderState.stdoutMu.Unlock()
}

func FlushWriter(writer *bufio.Writer) {
	if writer != nil {
		_ = writer.Flush()
	}
}
