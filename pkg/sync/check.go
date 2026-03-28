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

func (s *Sync) Check(targetDB string, givenDB string, activityID *string, activityType *string, schema string, logInFile bool) {

	var writer *bufio.Writer
	if logInFile && activityID != nil && activityType != nil {
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
	} else if logInFile {
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Printf("failed to create logs directory: %v", err)
			return
		}
		file, err := os.Create("logs/audit_check_" + time.Now().Format("20060102150405") + ".txt")
		if err != nil {
			log.Printf("failed to create audit check log file: %v", err)
			return
		}
		defer file.Close()
		writer = bufio.NewWriter(file)
		defer writer.Flush()
	}
	if writer != nil {
		fmt.Fprintf(writer, "Audit check started between %s and %s of %s schema\n", targetDB, givenDB, schema)
	} else {
		fmt.Printf("Audit check started between %s and %s of %s schema\n", targetDB, givenDB, schema)
	}

	targetDBAdapter := pg.New(s.TargetDB)
	givenDBAdapter := pg.New(s.GivenDB)

	ctx := context.Background()

	auditor := audit.SchemaAudit{}

	warnings, err := auditor.Check(ctx, targetDBAdapter, givenDBAdapter, schema)
	if err != nil {
		if writer != nil {
			fmt.Fprintf(writer, "Error during audit check: %v\n", err)
		} else {
			fmt.Printf("Error during audit check: %v\n", err)
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
	fmt.Println("Audit check completed.")
}
