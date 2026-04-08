package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/C5rogers/G-Synch/internal/audit/adapters/pg"
	"github.com/C5rogers/G-Synch/internal/audit/core"
	"github.com/C5rogers/G-Synch/internal/config"
	"github.com/C5rogers/G-Synch/internal/models"
	"github.com/C5rogers/G-Synch/pkg/sync"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v2"
)

// command should look like: go run cmd/main.go audit check|synch|reverse-check <given_db> <target_db> --config=config.yml --env=development

func main() {
	app := &cli.App{
		Name:  "G-Synch Server",
		Usage: "g-sync command --config=config.yml",
		Commands: []*cli.Command{
			{
				Name:  "audit",
				Usage: "audit between given db and target db",
				Subcommands: []*cli.Command{
					{
						Name:  "check",
						Usage: "check discrepancy at the moment between the given and the target db",
						Action: func(cliCtx *cli.Context) error {
							givenDB := cliCtx.Args().First()
							targetDB := cliCtx.Args().Get(1)
							schemaFlag := cliCtx.String("schema")
							logToFile := cliCtx.Bool("log-to-file")
							return run(cliCtx.String("config"), cliCtx.String("env"), "check", givenDB, targetDB, schemaFlag, logToFile)
						},
					},
					{
						Name:  "synch",
						Usage: "start synchronization between the given and the target db and fix discrepancy",
						Action: func(cliCtx *cli.Context) error {
							givenDB := cliCtx.Args().First()
							targetDB := cliCtx.Args().Get(1)
							schemaFlag := cliCtx.String("schema")
							logToFile := cliCtx.Bool("log-to-file")
							return run(cliCtx.String("config"), cliCtx.String("env"), "synch", givenDB, targetDB, schemaFlag, logToFile)
						},
					},
					{
						Name:  "reverse-check",
						Usage: "check discrepancy at the moment between the target and the given db",
						Action: func(cliCtx *cli.Context) error {
							givenDB := cliCtx.Args().First()
							targetDB := cliCtx.Args().Get(1)
							schemaFlag := cliCtx.String("schema")
							logToFile := cliCtx.Bool("log-to-file")
							return run(cliCtx.String("config"), cliCtx.String("env"), "reverse-check", givenDB, targetDB, schemaFlag, logToFile)
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "pass configuration path",
			},
			&cli.StringFlag{
				Name:  "env",
				Usage: "pass environment name",
			},
		&cli.StringFlag{
			Name:  "schema",
			Usage: "pass schema name; if omitted you will be prompted to choose one",
		},
			&cli.BoolFlag{
				Name:  "log-to-file",
				Usage: "log output to a file",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(configPath, env, cmd, givenDB string, targetDB string, schema string, logToFile bool) error {
	config, err := config.Load(configPath)
	if err != nil {
		return err
	}
	ctx := context.Background()

	target := config[targetDB]
	given := config[givenDB]

	if env != "" {
		given = fmt.Sprintf("%s_%s", env, givenDB)
		target = fmt.Sprintf("%s_%s", env, targetDB)
	}

	targetConn, err := pgxpool.New(ctx, config[target])
	if err != nil {
		slog.With("error", err).Error("error connecting to target db")
		panic(err)
	}

	givenConn, err := pgxpool.New(ctx, config[given])
	if err != nil {
		slog.With("error", err).Error("error connecting to given database")
		panic(err)
	}

	targetAdapter := pg.New(targetConn)
	givenAdapter := pg.New(givenConn)

	if schema == "" {
		schema, err = resolveSchema(ctx, targetAdapter, givenAdapter)
		if err != nil {
			slog.With("error", err).Error("error selecting schema")
			return err
		}
		fmt.Printf("Selected schema: %s\n", schema)
	}

	s, err := sync.NewSyncAPI(givenConn, targetConn)
	if err != nil {
		slog.With("error", err).Error("error setting up synch api")
		panic(err)
	}

	command := models.CMDMapper[models.CMD(cmd)]

	switch command {
	case string(models.CHECK):
		s.Check(targetDB, givenDB, nil, nil, schema, logToFile)
	case string(models.SYNCH):
		s.Synch(targetDB, givenDB, nil, nil, schema, logToFile)
	case string(models.REVERSE_CHECK):
		s.ReverseCheck(targetDB, givenDB, nil, nil, schema, logToFile)
	default:
		slog.With("cmd", cmd).Error("unknown command")
	}
	return nil
}

func resolveSchema(ctx context.Context, targetAdapter, givenAdapter core.SchemaAdapter) (string, error) {
	targetSchemas, err := targetAdapter.ListSchemas(ctx)
	if err != nil {
		return "", err
	}

	givenSchemas, err := givenAdapter.ListSchemas(ctx)
	if err != nil {
		return "", err
	}

	common := commonSchemas(targetSchemas, givenSchemas)
	if len(common) == 0 {
		return "", fmt.Errorf("no common schemas found between the given and target databases")
	}

	if len(common) == 1 {
		return common[0], nil
	}

	fmt.Println("Select a schema to act on:")
	for i, schemaName := range common {
		fmt.Printf("  %d) %s\n", i+1, schemaName)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter schema number: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			return common[0], nil
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 1 || selected > len(common) {
			fmt.Println("Invalid selection, try again.")
			continue
		}

		return common[selected-1], nil
	}
}

func commonSchemas(targetSchemas, givenSchemas []string) []string {
	targetSet := make(map[string]string, len(targetSchemas))
	for _, schemaName := range targetSchemas {
		targetSet[strings.ToLower(schemaName)] = schemaName
	}

	common := make([]string, 0)
	for _, schemaName := range givenSchemas {
		if original, ok := targetSet[strings.ToLower(schemaName)]; ok {
			common = append(common, original)
		}
	}

	sort.Strings(common)
	return common
}
