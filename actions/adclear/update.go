package adclear

import (
	"cleansync/messages"
	"fmt"
	"path/filepath"

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
	case RemoveAdsMsg:
		if !msg.started {
			m.currentProcess = fmt.Sprintf("Removing ads from: %s", m.sources[m.ndx])
			return m, tea.Sequence(tea.Printf(logo), m.RemoveAdsCmd(true, false, m.ndx))
		}
		if msg.done {
			// return m, tea.Sequence(tea.Printf("%s Removing ads from: %s", checkMark, m.sources[m.ndx]), m.CopyFileCmd(false, false, m.ndx))
			m.ndx = 0
			m.editedVideos = append(m.editedVideos, msg.tmpLocation)
			return m, tea.Sequence(tea.Printf("%s Removing ads from: %s", checkMark, m.sources[m.ndx]), m.CopyFileCmd(false, false, m.ndx))
		}

		// Main Loop
		m.editedVideos = append(m.editedVideos, msg.tmpLocation)
		m.ndx++
		m.currentProcess = fmt.Sprintf("Removing ads from: %s", m.sources[m.ndx])
		return m, tea.Sequence(tea.Printf("%s Removing ads from: %s", checkMark, m.sources[m.ndx-1]), m.RemoveAdsCmd(true, false, m.ndx))
	case CopyFileMsg:
		m.progressor.ResetProgress()
		vidname := filepath.Base(m.editedVideos[m.ndx])
		if !msg.started {
			// Not started yet
			m.currentProcess = fmt.Sprintf("Copying file from %s to %s", vidname, m.dest)

			return m, tea.Batch(m.CopyFileCmd(true, false, m.ndx))
		}
		if msg.done {
			return m, tea.Sequence(tea.Printf("%s Copying file: %s to %s", checkMark, vidname, m.dest), tea.Quit)
		}
		if msg.err != nil {
			m.SendError(msg.err)
		}
		m.ndx++
		nextVid := filepath.Base(m.editedVideos[m.ndx])
		m.currentProcess = fmt.Sprintf(" Copying file: %s to %s", nextVid, m.dest)
		return m, tea.Sequence(tea.Printf("%s Copying file: %s to %s", checkMark, vidname, m.dest), m.CopyFileCmd(true, false, m.ndx))
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
