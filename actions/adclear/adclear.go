package adclear

import (
	"cleansync/filesystem"
	"cleansync/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

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

	vid := NewVideo(source, dest, skip, progressor)
	prog := tea.NewProgram(&vid)
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

	_, err := prog.Run()
	if err != nil {
		return err
	}

	return nil
}
