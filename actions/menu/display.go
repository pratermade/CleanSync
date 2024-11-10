package menu

import (
	"cleansync/filesystem"
	"cleansync/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

func Display(c *cli.Context) error {
	menu := NewMenu()

	progressor := &filesystem.ProgressReadWriter{}
	ch := make(chan messages.ProgressMsg)
	go progressor.GetProgress(ch)

	prog := tea.NewProgram(&menu)
	if menu.err != nil {
		return menu.err
	}

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
