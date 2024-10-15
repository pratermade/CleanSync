package adclear

import (
	"cleansync/messages"
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m VideoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case RemoveAdsMsg:
		if !msg.started {
			m.currentProcess = fmt.Sprintf("Removing ads from: %s", m.source)
			return m, tea.Batch(m.RemoveAdsCmd(true, false, ""))
		}
		if msg.done {
			m.currentProcess = fmt.Sprintf("Copying file from %s to %s", msg.tmpLocation, m.dest)
			return m, tea.Sequence(tea.Printf("%s Removing ads from: %s", checkMark, m.source), m.CopyFileCmd(false, false, msg.tmpLocation))
		}
		return m, tea.Sequence(m.RemoveAdsCmd(true, true, msg.tmpLocation))
	case CopyFileMsg:
		// This one is a good prototypical example
		if !msg.started {
			// Not started yet
			m.progressor.ResetProgress()
			return m, tea.Batch(m.CopyFileCmd(true, false, msg.tmpLocation))
		}
		if msg.done {
			return m, tea.Sequence(tea.Printf("%s Copying file: %s", checkMark, m.source), tea.Quit)
		}
		if msg.err != nil {
			m.SendError(msg.err)
		}
		return m, tea.Sequence(m.CopyFileCmd(true, true, msg.tmpLocation))
	case CleanTmpMsg:
		// All Done
		return m, tea.Batch(tea.Printf("Done processing file: %s", m.source), tea.Quit)
	case messages.ErrMsg:
		// handle errorI guess
		return m, tea.Quit
	case messages.ProgressMsg:
		progressCmd := m.progress.SetPercent(msg.Progress)
		return m, tea.Batch(progressCmd)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd)
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}
