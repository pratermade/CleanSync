package processVideo

import (
	"cleansync/ffmpeg"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m *VideoModel) SendError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err}
	}
}

func (m *VideoModel) ProcessVideoCmd(s Status, ndx int) tea.Cmd {
	return func() tea.Msg {
		msg := ProcessVideoMessage{
			ndx:    ndx,
			status: s,
		}

		switch s {
		case Starting:
			return ProcessVideoMessage{
				ndx:        ndx,
				lastAction: fmt.Sprintf("%s Started processing.", checkMark),
				nextAction: fmt.Sprintf("Removing Ads file: %s", m.sources[m.ndx]),
				status:     RemovingAds,
			}

		case RemovingAds:
			// Start Removing the ads

			vid, err := ffmpeg.NewVideo(m.sources[ndx], m.tempFolder)
			if err != nil {
				return VideoModel{err: err}

			}
			vid.TmpFolder = m.tempFolder
			nonAdIndexes := vid.GetNonAdIndexes(m.skipFirst)

			tmpVideo, err := vid.Recut(nonAdIndexes)
			if err != nil {
				msg.err = err
				return err
			}
			msg = ProcessVideoMessage{
				ndx:         ndx,
				lastAction:  fmt.Sprintf("%s Removed Ads file: %s", checkMark, m.sources[ndx]),
				nextAction:  fmt.Sprintf("Uploading file: %s", m.sources[m.ndx]),
				status:      Uploading,
				tmpLocation: tmpVideo,
			}
			return msg
		case Uploading:
			// Start uploading
			err := os.MkdirAll(m.dest, os.ModePerm)
			if err != nil {
				msg.err = err
				return msg
			}
			vidName := filepath.Base(m.editedVideo)
			dest := filepath.Join(m.dest, vidName)
			err = m.progressor.Copy(m.editedVideo, dest)
			if err != nil {
				msg.err = err
				return msg
			}
			// Remove the processed file
			err = os.Remove(m.editedVideo)
			if err != nil {
				msg.err = err
				return msg
			}
			nextAction := "completed"
			if ndx <= len(m.sources)-1 {
				nextAction = fmt.Sprintf("Removing Ads file: %s", m.sources[ndx])
			}

			return ProcessVideoMessage{
				ndx:        ndx,
				status:     Idle,
				lastAction: fmt.Sprintf("%s Uploaded file: %s", checkMark, m.sources[ndx]),
				nextAction: nextAction,
			}
		case Idle:
			// Video is both processed and uploaded
			if ndx >= len(m.sources)-1 {
				return ProcessVideoMessage{
					status: Completed,
				}
			}
			msg = ProcessVideoMessage{
				lastAction: fmt.Sprintf("%s Removing Ads file: %s", checkMark, m.sources[ndx+1]),
				ndx:        ndx + 1,
				status:     Idle,
			}
			return msg
		default:
			return ProcessVideoMessage{
				status: Completed,
			}
		}
	}
}
