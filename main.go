package main

import (
	"cleansync/actions/adclear"
	"cleansync/actions/sync"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "s3sync",
		Usage: "Sync with the provided s3 bucket",
		Commands: []*cli.Command{
			{
				Name:   "adclear",
				Usage:  "Removes adds from the source and copies the resulting video to the destination",
				Action: adclear.Clear,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:     "source",
						Usage:    "The source file or folder, if it is a folder, it will attempt to process all video files. (currently mp4, mkv)",
						Required: true,
					},
					&cli.PathFlag{
						Name:     "dest",
						Usage:    "The destination file or folder",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "skip_first",
						Usage:    "Skips the first chapter, thus omiting it from the final product. Usefull for removing that 'Recorded by...' at the begining of playon videos",
						Required: false,
					},
				},
			},
			{
				Name:   "sync",
				Usage:  "upload new files to the provided bucket",
				Action: sync.Sync,
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
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
