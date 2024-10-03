package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"s3sync/messages"
	"s3sync/splitter"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m UploadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case messages.SplitMsg:

		// Get some initial states for the process
		tmpDir, err := os.MkdirTemp("", "s3sync")
		if err != nil {
			return m, m.SendError(err)
		}
		m.indention = "    "
		partName := fmt.Sprintf("%s.part%d", filepath.Base(msg.OrgFilePath), 0)
		partPath := filepath.Join(tmpDir, partName)
		info := &splitter.SplitInfo{
			OrgFilePath: msg.OrgFilePath,
			TempFolder:  tmpDir,
			PartPath:    partPath,
		}
		name := filepath.Base(msg.OrgFilePath)
		m.currentProcess = fmt.Sprintf("%s  Splitting Part: %s", m.indention, partName)
		return m, tea.Batch(
			tea.Printf("%s  %s is too big, splitting into parts.", flagMark, name),
			m.splitCmd(info),
		)

	case *splitter.SplitInfo:
		partname := filepath.Base(msg.PartPath)

		// This is the exit point once the file is completly split
		if msg.Eof {
			ctx := context.Background()
			test := []string{
				"test1", "test2",
			}
			return m, tea.Batch(
				tea.Printf("%s%s  Splitting part: %s", m.indention, checkMark, partname),
				m.uploadParts(ctx, test),
				// progressCmd,
			)
		}
		// Ths section is what is executed during the split
		partPath := fmt.Sprintf("%s.part%d", filepath.Base(msg.OrgFilePath), msg.Count)
		msg.PartPath = filepath.Join(msg.TempFolder, partPath)
		partname = filepath.Base(msg.PartPath)
		m.currentProcess = fmt.Sprintf("%s  Splitting Part: %s", m.indention, partname)
		return m, tea.Batch(
			tea.Printf("%s%s  Splitting part: %s", m.indention, checkMark, msg.PreviousPart),
			m.splitCmd(msg),
		)

	case messages.UploadPartsMsg:
		m.indention = ""
		return m, tea.Sequence(
			tea.Printf("%s%s  Uploading parts...", flagMark, m.indention),
			tea.Quit,
		)

	case messages.ErrMsg:
		// handle errorI guess
		return m, tea.Quit
	case messages.UploadMsg:
		pkg := m.toUpdate[m.index]
		m.index++
		// progressCmd := m.progress.SetPercent(float64(m.index+1.0) / float64(len(m.toUpdate)))
		if m.index >= len(m.toUpdate)-1 {
			// Everything's been Downloaded. We're done!
			m.done = true
			return m, tea.Sequence(
				// progressCmd,
				tea.Printf("%s %s", checkMark, pkg), // print the last success message
				tea.Quit,                            // exit the program
			)
		}

		ctx := context.Background()
		return m, tea.Batch(
			tea.Printf("%s %s", checkMark, msg.Name),
			m.PutObject(ctx, m.toUpdate[m.index], m.progressor),
		)
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
