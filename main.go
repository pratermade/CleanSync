package main

import (
	"context"
	"os"
	"s3sync/filesystem"
	"s3sync/localsql"
	"s3sync/messages"
	"s3sync/ui"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "s3sync",
		Usage: "Sync with the provided s3 bucket",
		Commands: []*cli.Command{
			{
				Name:  "sync",
				Usage: "upload new files to the provided bucket",
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "The source (local) folder to sync with S3",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "bucket",
						Aliases:  []string{"b"},
						Usage:    "The name of the bucket to sysnc to",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     "filter",
						Aliases:  []string{"f"},
						Usage:    "file types to filter for. Can be specified multiple times for multiple file types.",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "deep",
						Aliases:  []string{"d"},
						Usage:    "deep archive in S3",
						Required: false,
					},
				},
				Action: func(c *cli.Context) error {
					err := sync(c.String("bucket"), c.String("path"), c.StringSlice("filter"), c.Bool("deep"))
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func sync(bucket string, folderPath string, filters []string, deep bool) error {
	ctx := context.Background()

	client, err := getAwsClient(ctx)
	if err != nil {
		return err
	}

	db, err := localsql.InitDb("manifest.db")
	if err != nil {
		return err
	}

	files, err := filesystem.WalkAndHash(filters, folderPath)
	if err != nil {
		return err
	}

	err = db.UpdateManifest(files)
	if err != nil {
		return err
	}

	uploads, err := db.GetUploadList()
	if err != nil {
		return err
	}

	// So we can monitor the progress of the file file writing
	progressor := &messages.ProgressReadWriter{}
	ch := make(chan messages.ProgressMsg)
	go progressor.GetProgress(ch)
	//

	// This should send it to the execution loop
	prog := tea.NewProgram(ui.NewModel(folderPath, client, uploads, bucket, db, filters, progressor, deep))

	//Sends progress status for video reads/writes
	go func() {
		for {
			update := <-ch
			prog.Send(update)
		}
	}()

	_, err = prog.Run()
	if err != nil {
		return err
	}

	return nil
}

func getAwsClient(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	return client, nil
}
