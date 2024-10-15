package adclear

import (
	"cleansync/ffmpeg"
	"cleansync/filesystem"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type VideoModel struct {
	width          int
	height         int
	video          ffmpeg.Video
	spinner        spinner.Model
	progress       progress.Model
	done           bool
	currentProcess string
	progressor     *filesystem.ProgressReadWriter
	err            error
	skipFirst      bool
	dest           string
	source         string
}

var (
	doneStyle = lipgloss.NewStyle().Margin(1, 2)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

// NewModel initializes and returns a new model
func NewVideo(source string, dest string, skipFirst bool, progressor *filesystem.ProgressReadWriter) VideoModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	vid, err := ffmpeg.NewVideo(source)
	if err != nil {
		return VideoModel{err: err}

	}

	vm := VideoModel{
		spinner:    s,
		progress:   p,
		video:      vid,
		progressor: progressor,
		skipFirst:  skipFirst,
		dest:       dest,
		source:     source,
	}

	return vm
}

// Init is the entry point of the ui/program
func (m VideoModel) Init() tea.Cmd {
	return tea.Batch(m.RemoveAdsCmd(false, false, ""), m.spinner.Tick)
}

// View is the initial state of the ui
func (m VideoModel) View() string {

	// w := lipgloss.Width(fmt.Sprintf("%d", 1))

	if m.done {
		return doneStyle.Render("Done! Message")
	}

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog))
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(m.currentProcess)
	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog
}

// max does what it implies and returns the mbigger of 2 ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
