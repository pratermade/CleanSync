package adclear

import (
	"cleansync/ffmpeg"
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

func (m *VideoModel) RemoveAdsCmd(started bool, done bool, ndx int) tea.Cmd {
	return func() tea.Msg {
		msg := RemoveAdsMsg{
			started: started,
			done:    done,
		}
		if started {
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
			if ndx >= len(m.sources)-1 {
				msg.done = true
			}
			msg.tmpLocation = tmpVideo
			return msg
		}
		return msg
	}
}

func (m *VideoModel) CopyFileCmd(started bool, done bool, ndx int) tea.Cmd {
	// the dest fiolder if it does not exist

	return func() tea.Msg {
		msg := CopyFileMsg{
			started: started,
			done:    done,
		}

		err := os.MkdirAll(m.dest, os.ModePerm)
		if err != nil {
			msg.err = err
			return msg
		}
		if started {
			vidName := filepath.Base(m.editedVideos[ndx])
			dest := filepath.Join(m.dest, vidName)
			err := m.progressor.Copy(m.editedVideos[ndx], dest)
			if err != nil {
				msg.err = err
				return msg
			}
			if ndx >= len(m.editedVideos)-1 {
				err = os.RemoveAll(m.tempFolder)
				if err != nil {
					msg.err = err
					return msg
				}
				msg.done = true
			}
			return msg
		}
		return msg
	}
}
