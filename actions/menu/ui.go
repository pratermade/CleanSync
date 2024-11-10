package menu

import (
	"cleansync/filesystem"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuModel struct {
	width          int
	height         int
	progress       progress.Model
	spinner        spinner.Model
	done           bool
	progressor     *filesystem.ProgressReadWriter
	index          int
	currentProcess string
	err            error
}

var (
	doneStyle = lipgloss.NewStyle().Margin(1, 2)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	flagMark  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2C900")).SetString("⚑")
)

func NewMenu() MenuModel {
	return MenuModel{}
}

// Init is the entry point of the ui/program
func (m MenuModel) Init() tea.Cmd {
	// ctx := context.Background()
	return tea.Batch(m.spinner.Tick)
}

// View is the initial state of the ui
func (m MenuModel) View() string {
	//n := len(m.toUpdate)
	n := 1
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done! Processed %d packages.\n", n))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(m.currentProcess)
	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + pkgCount
}

// max does what it implies and returns the mbigger of 2 ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
