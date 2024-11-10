package menu

import (
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m *MenuModel) SendError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err}
	}

}
