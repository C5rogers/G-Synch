package sync

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/C5rogers/G-Synch/internal/audit"
	"github.com/C5rogers/G-Synch/internal/audit/adapters/pg"
)

func (s *Sync) ReverseCheck(targetDB string, givenDB string, activityID *string, activityType *string, schema string, logToFile bool) {
	var writer *bufio.Writer
	if logToFile && activityID != nil && activityType != nil {
		file, err := os.Create("logs/" + *activityID + "_" + *activityType + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(file)
	} else if logToFile {
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Fatal(err)
		}
		file, err := os.Create("logs/audit_reverse_check_" + time.Now().Format("20060102150405") + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(file)
	}

	if writer != nil {
		fmt.Fprintf(writer, "Audit reverse check started between %s and %s of %s schema\n", targetDB, givenDB, schema)
	} else {
		fmt.Printf("Audit reverse check started between %s and %s of %s schema\n", targetDB, givenDB, schema)
	}

	// Reverse check compares the given DB as target against the target DB as given.
	reverseTargetAdapter := pg.New(s.GivenDB)
	reverseGivenAdapter := pg.New(s.TargetDB)

	ctx := context.Background()
	auditor := audit.SchemaAudit{}

	warnings, err := auditor.Check(ctx, reverseTargetAdapter, reverseGivenAdapter, schema)
	if err != nil {
		if writer != nil {
			fmt.Fprintf(writer, "Error during reverse audit check: %v\n", err)
		} else {
			fmt.Printf("Error during reverse audit check: %v\n", err)
		}
	}

	for _, warning := range warnings {
		if writer != nil {
			messageToLog := fmt.Sprintf("%s (%s): %s\n", warning.Label, warning.Type, warning.Message)
			fmt.Fprintf(writer, "%s\n", messageToLog)
		} else {
			fmt.Println(warning.GetColoredMessage())
		}
	}

	FlushWriter(writer)

	fmt.Println("Audit reverse check completed.")

}
