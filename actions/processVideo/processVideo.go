package processVideo

import (
	"cleansync/filesystem"
	"cleansync/messages"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

// Clear is a CLI command handler that processes video files by clearing
// or modifying them based on the provided options.
//
// Parameters:
//   - c: The CLI context containing the command-line arguments.
//
// Expected Flags:
//   - source: The file path to the source video file.
//   - skip_first: A boolean flag indicating whether to skip the first frame or section of the video.
//   - dest: The file path to the destination where the processed video will be saved.
//
// Returns:
//   - An error if the clearing process fails, otherwise nil.
func Clear(c *cli.Context) error {
	source := c.Path("source")
	skip := c.Bool("skip_first")
	dest := c.Path("dest")

	return clear(source, dest, skip)
}

// To ease testing
func clear(source string, dest string, skip bool) error {

	// So we can monitor the progress of the file file writing
	progressor := &filesystem.ProgressReadWriter{}
	ch := make(chan messages.ProgressMsg)
	go progressor.GetProgress(ch)
	//

	vid, err := NewVideo(source, dest, skip, progressor)
	if err != nil {
		return err
	}
	defer os.RemoveAll(vid.tempFolder)

	prog := tea.NewProgram(vid)
	if vid.err != nil {
		return vid.err
	}

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
