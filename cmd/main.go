package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

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
							return run(cliCtx.String("config"), cliCtx.String("env"), "check", givenDB, targetDB)
						},
					},
					{
						Name:  "synch",
						Usage: "start synchronization between the given and the target db and fix discrepancy",
						Action: func(cliCtx *cli.Context) error {
							givenDB := cliCtx.Args().First()
							targetDB := cliCtx.Args().Get(1)
							return run(cliCtx.String("config"), cliCtx.String("env"), "synch", givenDB, targetDB)
						},
					},
					{
						Name:  "reverse-check",
						Usage: "check discrepancy at the moment between the target and the given db",
						Action: func(cliCtx *cli.Context) error {
							givenDB := cliCtx.Args().First()
							targetDB := cliCtx.Args().Get(1)
							return run(cliCtx.String("config"), cliCtx.String("env"), "reverse-check", givenDB, targetDB)
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
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(configPath, env, cmd, givenDB string, targetDB string) error {
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

	fmt.Println("env:", env, given, givenDB, target, targetDB)
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

	s, err := sync.NewSyncAPI(givenConn, targetConn)

	if err != nil {
		slog.With("error", err).Error("error setting up synch api")
		panic(err)
	}

	// i want to check the cmd to be the specific pre defined struct types and execute against the commands

	command := models.CMDMapper[models.CMD(cmd)]

	switch command {
	case string(models.CHECK):
		s.Check(givenDB, nil, nil)
	case string(models.SYNCH):
		s.Synch(givenDB, nil, nil)
	case string(models.REVERSE_CHECK):
		s.ReverseCheck(givenDB, nil, nil)
	default:
		slog.With("cmd", cmd).Error("unknown command")
	}
	return nil
}
