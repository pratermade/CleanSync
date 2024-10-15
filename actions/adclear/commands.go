package adclear

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m *VideoModel) SendError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err}
	}
}

func (m *VideoModel) RemoveAdsCmd(started bool, done bool, tmpLocation string) tea.Cmd {
	return func() tea.Msg {
		msg := RemoveAdsMsg{
			started:     started,
			tmpLocation: tmpLocation,
			done:        done,
		}
		if started {
			nonAdIndexes := m.video.GetNonAdIndexes(m.skipFirst)
			tmpVideo, err := m.video.Recut(nonAdIndexes)
			if err != nil {
				msg.err = err
				return err
			}
			msg.tmpLocation = tmpVideo
			msg.done = true
			return msg
		}
		return msg
	}
}

func (m *VideoModel) CopyFileCmd(started bool, done bool, tmpLocation string) tea.Cmd {
	return func() tea.Msg {
		msg := CopyFileMsg{
			started:     started,
			done:        done,
			tmpLocation: tmpLocation,
		}
		if started {
			err := m.progressor.Copy(tmpLocation, m.dest)
			if err != nil {
				msg.err = err
				return err
			}
			err = os.RemoveAll(m.video.TmpFolder)
			if err != nil {
				msg.err = err
				return msg
			}
			msg.done = true
			return msg
		}
		return msg
	}
}
