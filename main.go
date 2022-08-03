package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/salad-server/cli/cmd"
	"github.com/salad-server/cli/utils"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func main() {
	err := utils.LoadUtils()
	bad := errors.New("Invalid arguments!")
	app := &cli.App{
		Name:  "cli",
		Usage: "salad-cli, small time jobs for your server!",

		Commands: []*cli.Command{
			{
				Name:  "update",
				Usage: "Update status for beatmaps (default qualified)",
				Action: func(ctx *cli.Context) error {
					if id := ctx.Int("beatmap"); id != 0 {
						return cmd.UpdateSet(id)
					}

					if status := ctx.String("status"); slices.Contains([]string{"pending", "ranked", "approved", "qualified", "loved"}, status) {
						return cmd.UpdateSetStatus(status)
					}

					fmt.Println("Must be a beatmapset ID or status! {pending|ranked|approved|qualified|loved}")
					return bad
				},

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "status",
						Aliases: []string{"s"},
						Usage:   "Update status for {pending|ranked|approved|qualified|loved} beatmaps",
					},

					&cli.IntFlag{
						Name:    "beatmap",
						Aliases: []string{"b", "i"},
						Usage:   "Update status for <beatmap set id>",
					},
				},
			},

			{
				Name:  "backup",
				Usage: "Create backup .zip including replays, user data, and SQL dump",
				Action: func(ctx *cli.Context) error {
					return cmd.Backup(
						!ctx.Bool("sql"),
						!ctx.Bool("replays"),
						!ctx.Bool("data"),
					)
				},

				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "sql",
						Aliases: []string{"s"},
						Usage:   "Don't include SQL dump in backup process",
					},

					&cli.BoolFlag{
						Name:    "replays",
						Aliases: []string{"r"},
						Usage:   "Don't include replays in backup process",
					},

					&cli.BoolFlag{
						Name:    "data",
						Aliases: []string{"d"},
						Usage:   "Don't include data (profile pictures, etc...) in backup process",
					},
				},
			},

			{
				Name:  "pb",
				Usage: "Mark a score as a personal best",
				Action: func(ctx *cli.Context) error {
					id, err := strconv.Atoi(ctx.Args().Get(0))

					if err != nil {
						fmt.Println("Invalid score ID!", err)
						return bad
					}

					return cmd.PersonalBest(id)
				},
			},
		},
	}

	if appErr := app.Run(os.Args); appErr != nil || err != nil {
		fmt.Println(
			"Could not run salad-cli!",
			"Please open an issue here:",
			"https://github.com/salad-server/cli/issues",
		)
	}
}
