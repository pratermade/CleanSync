package sync

import (
	"cleansync/filesystem"
	"cleansync/localsql"
	"cleansync/messages"
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

func Sync(c *cli.Context) error {

	bucket := c.String("bucket")
	folderPath := c.String("path")
	filters := c.StringSlice("filter")
	deep := c.Bool("deep")

	ctx := c.Context
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
	progressor := &filesystem.ProgressReadWriter{}
	ch := make(chan messages.ProgressMsg)
	go progressor.GetProgress(ch)
	//

	// This should send it to the execution loop
	prog := tea.NewProgram(NewModel(folderPath, client, uploads, bucket, db, filters, progressor, deep))

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
