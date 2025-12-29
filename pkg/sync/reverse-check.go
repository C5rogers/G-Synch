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
	"github.com/C5rogers/G-Synch/internal/models"
)

func (s *Sync) ReverseCheck(targetDB string, givenDB string, activityID *string, activityType *string, schema string, logInFile bool) {
	var writer *bufio.Writer
	if logInFile && activityID != nil && activityType != nil {
		file, err := os.Create("logs/" + *activityID + "_" + *activityType + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(file)
	} else if logInFile {
		// create the below file if it does not exist
		if err := os.MkdirAll("logs", os.ModePerm); err != nil {
			log.Fatal(err)
		}
		file, err := os.Create("logs/audit_check_" + time.Now().Format("20060102150405") + ".txt")
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

	targetDBAdapter := pg.New(s.GivenDB)
	givenDBAdapter := pg.New(s.TargetDB)

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
	if len(warnings) == 0 {
		warning := models.CheckReturn{
			Message: fmt.Sprintf("No differences found of %s relative to %s.", targetDB, givenDB),
			Type:    "INFO",
			Label:   "SUCCESS",
		}
		warnings = append(warnings, warning)
	}
	for _, warning := range warnings {
		if writer != nil {
			messageToLog := fmt.Sprintf("%s (%s): %s\n", warning.Label, warning.Type, warning.Message)
			fmt.Fprintf(writer, "%s\n", messageToLog)
		} else {
			fmt.Println(warning.GetColoredMessage())
		}
	}
	if writer != nil {
		writer.Flush()
	}
	fmt.Println("Audit reverse check completed.")
	time.Sleep(2 * time.Second)
}
