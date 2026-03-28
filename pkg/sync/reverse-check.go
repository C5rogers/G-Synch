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
	loader := NewLoader("Reading database metadata, dependencies, and rows...")
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
		defer writer.Flush()
	} else if logToFile {
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Printf("failed to create logs directory: %v", err)
			return
		}
		file, err := os.Create("logs/audit_reverse_check_" + time.Now().Format("20060102150405") + ".txt")
		if err != nil {
			log.Printf("failed to create reverse check log file: %v", err)
			return
		}
		defer file.Close()
		writer = bufio.NewWriter(file)
		defer writer.Flush()
	}

	Printf(writer, "Audit reverse check started between %s and %s of %s schema\n", targetDB, givenDB, schema)

	// Reverse check compares the given DB as target against the target DB as given.
	reverseTargetAdapter := pg.New(s.GivenDB)
	reverseGivenAdapter := pg.New(s.TargetDB)

	ctx := context.Background()
	auditor := audit.SchemaAudit{}

	warnings, err := auditor.Check(ctx, reverseTargetAdapter, reverseGivenAdapter, schema)
	if err != nil {
		Printf(writer, "Error during reverse audit check: %v\n", err)
	}

	for _, warning := range warnings {
		if writer != nil {
			messageToLog := fmt.Sprintf("%s (%s): %s", warning.Label, warning.Type, warning.Message)
			Printf(writer, "%s\n", messageToLog)
		} else {
			Println(writer, warning.GetColoredMessage())
		}
	}

	FlushWriter(writer)

	Println(nil, "Audit reverse check completed.")

}
