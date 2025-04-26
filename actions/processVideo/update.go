package processVideo

import (
	"cleansync/messages"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

const logo = `
			   ________               _____                 
			  / ____/ /__  ____ _____/ ___/ __  ______  _____
			 / /   / / _ \/ __ /  __ \\__ \/ / / / __ \/ ___/
			/ /___/ /  __/ /_/ / / / /__/ / /_/ / / / / /__  
			\____/_/\___/\__,_/_/ /_/____/\__, /_/ /_/\___/  
			-----------------------------/____/---------------      
`

func (m VideoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case ProcessVideoMessage:
		// Process the video
		if msg.err != nil {
			m.SendError(msg.err)
			return m, tea.Quit
		}
		if msg.status == Completed {
			m.done = true
			return m, tea.Batch(tea.Printf("Done processing file: %s", m.sources[m.ndx]), tea.Quit)
		}
		// Main Loop
		m.ndx = msg.ndx
		m.editedVideo = msg.tmpLocation

		m.currentProcess = msg.nextAction
		return m, tea.Sequence(tea.Printf(msg.lastAction), m.ProcessVideoCmd(msg.status, m.ndx))
	case CleanTmpMsg:
		// All Donewinter
		return m, tea.Batch(tea.Printf("Done processing file: %s", m.sources[m.ndx]), tea.Quit)
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
