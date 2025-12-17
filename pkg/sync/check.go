package sync

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/C5rogers/G-Synch/internal/audit"
	"github.com/C5rogers/G-Synch/internal/audit/adapters/pg"
)

func (s *Sync) Check(targetDB string, givenDB string, activityID *string, activityType *string, schema string) {
	// here we will setup an audit because all commands are audit struct types and the audit will abstract us the check and other sub commands

	// var file *os.File
	var writer *bufio.Writer
	// TODO: implement the audit for each tables on the schema
	if activityID != nil && activityType != nil {
		file, err := os.Create("logs/" + *activityID + "_" + *activityType + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(file)
	}
	if writer != nil {
		fmt.Fprintf(writer, "Audit check started between %s and %s of %s schema\n", targetDB, givenDB, schema)
	} else {
		fmt.Printf("Audit check started between %s and %s of %s schema\n", targetDB, givenDB, schema)
	}

	// yellow := color.New(color.FgYellow).SprintFunc()
	// red := color.New(color.FgRed).SprintFunc()
	// blue := color.New(color.FgBlue).SprintFunc()
	// green := color.New(color.FgGreen).SprintFunc()

	// create the adapters for the target and given
	targetDBAdapter := pg.New(s.TargetDB)
	givenDBAdapter := pg.New(s.GivenDB)

	ctx := context.Background()
	targetSchemaAdapter, err := targetDBAdapter.LoadSchema(schema)
	if err != nil {
		if writer != nil {
			fmt.Fprintf(writer, "Error loading target schema: %v\n", err)
		} else {
			fmt.Printf("Error loading target schema: %v\n", err)
		}
		return
	}
	givenSchemaAdapter, err := givenDBAdapter.LoadSchema(schema)
	if err != nil {
		if writer != nil {
			fmt.Fprintf(writer, "Error loading given schema: %v\n", err)
		} else {
			fmt.Printf("Error loading given schema: %v\n", err)
		}
		return
	}

	auditor := audit.SchemaAudit{}

	// here create the pg audit to run audit.Check function

	warnings, err := auditor.Check(ctx, targetSchemaAdapter, givenSchemaAdapter, schema)
	if err != nil {
		if writer != nil {
			fmt.Fprintf(writer, "Error during audit check: %v\n", err)
		} else {
			fmt.Printf("Error during audit check: %v\n", err)
		}
	}
	fmt.Println(warnings)

}
