package main

import (
	"context"
	"fmt"
	"os"
	"s3sync/syncer"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
		fmt.Println(err)
	}
}

func sync(bucket string, p string, filters []string, deep bool) error {
	ctx := context.Background()

	client, err := getAwsClient(ctx)
	if err != nil {
		return err
	}

	app := syncer.Syncer{
		Bucket:     bucket,
		FolderPath: p,
		S3Client:   client,
	}

	app.InitDb("manifest.db")

	// get a list of the actual files in the folder
	fileMap, err := app.WalkAndHash(filters)
	if err != nil {
		return err
	}

	// Update the manifest with any new or updated files
	err = app.UpdateManifest(fileMap)
	if err != nil {
		return err
	}

	// Get any items that has not been set as uploaded
	uploads, err := app.GetUploadList()
	if err != nil {
		return err
	}

	err = app.UploadDiffs(ctx, uploads, deep)
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

func diffMaps(maniMap map[string]string, fileMap map[string]string) map[string]string {
	if maniMap == nil {
		return fileMap
	}
	resMap := make(map[string]string)
	for fp, h := range fileMap {
		val, ok := maniMap[fp]
		if ok {
			if val == h {
				continue
			}
		}
		resMap[fp] = h
	}
	return nil
}
