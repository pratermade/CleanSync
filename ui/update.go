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

	case *splitter.SplitInfo:

		// This is the exit point once the file is completly split. We move on to uploading for the next step
		if msg.Eof {
			ctx := context.Background()
			uploadPartsCmd := m.uploadParts(ctx, msg.OrgFilePath, msg.Parts, 0)
			m.currentProcess = fmt.Sprintf("%s  Uploading part: %s", m.indention, msg.Parts[0])
			storage := "Standard Storage"
			if m.deep {
				storage = "Glacier Deep Archive"
			}
			return m, tea.Batch(
				tea.Printf("%s%s  Uploading %d parts of %s to %s us %s", flagMark, m.indention, len(msg.Parts), filepath.Base(msg.OrgFilePath), m.bucket, storage),
				uploadPartsCmd,
			)
		}
		//

		msg.Parts = append(msg.Parts, fmt.Sprintf("%s.part%d", filepath.Base(msg.OrgFilePath), msg.Index))
		m.indention = "   "
		output := fmt.Sprintf("%s%s  1. Splitting part: %s", m.indention, checkMark, msg.Parts[msg.Index])
		m.currentProcess = fmt.Sprintf("%s  2. Splitting part: %s", m.indention, msg.Parts[msg.Index])
		return m, tea.Sequence(
			m.splitCmd(msg),
			tea.Printf(output),
		)

	case messages.UploadPartsMsg:
		m.indention = "\t"
		pkg := msg.Parts[msg.Index]
		m.progressor.ResetProgress()
		m.currentProcess = fmt.Sprintf("%s  Uploading part: %s", m.indention, pkg)
		if msg.Index >= len(msg.Parts) {
			return m, tea.Batch(tea.Printf("%s%s Parts Uploaded for %s", m.indention, checkMark, msg.OriginalFile), m.uploadCmd(&messages.UploadMsg{Done: false}))
		}
		m.currentProcess = fmt.Sprintf("%s  Uploading part: %s", m.indention, pkg)
		msg.Index++
		ctx := context.Background()
		return m, tea.Batch(
			tea.Printf("%s%s Uploading part: %s", m.indention, checkMark, pkg),
			m.uploadParts(ctx, msg.OriginalFile, msg.Parts, msg.Index),
		)
	case messages.UploadMsg:
		// Were going to check first to see if we needc to split the file up

		pkg := m.toUpdate[m.index]
		info, err := os.Stat(pkg)
		if err != nil {
			return m, m.SendError(err)
		}

		if info.Size() > 4294967296 {
			si := &splitter.SplitInfo{
				OrgFilePath: m.toUpdate[m.index],
				Parts:       []string{},
				OrgFileSize: info.Size(),
			}
			return m, tea.Sequence(
				tea.Printf("%s%s  %s is too big, splitting into parts.", flagMark, m.indention, filepath.Base(pkg)),
				m.splitCmd(si),
			)
		}
		// Do the upload
		ctx := context.Background()
		err = m.doUpload(ctx, m.toUpdate[m.index], m.progressor, info.Size())
		if err != nil {
			return m, m.SendError(err)
		}

		if m.index >= len(m.toUpdate)-1 {
			// Everything's been Downloaded. We're done!
			m.done = true
			return m, tea.Sequence(
				// progressCmd,
				tea.Printf("%s %s", checkMark, m.toUpdate[m.index]), // print the last success message
				tea.Quit, // exit the program
			)
		}
		m.index++
		return m, tea.Batch(
			tea.Printf("%s %s", checkMark, pkg),
			m.PutObjectCmd(ctx),
		)
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
