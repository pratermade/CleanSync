package processVideo

import (
	"cleansync/filesystem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type VideoModel struct {
	width          int
	height         int
	spinner        spinner.Model
	progress       progress.Model
	done           bool
	currentProcess string
	progressor     *filesystem.ProgressReadWriter
	err            error
	skipFirst      bool
	dest           string
	sources        []string
	ndx            int
	tempFolder     string
	editedVideo    string
}

const defaultWidth = 40

var supportedExtensions = []string{".mp4", ".mkv"}
var (
	doneStyle = lipgloss.NewStyle().Margin(1, 2)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

// NewModel initializes and returns a new model
func NewVideo(source string, dest string, skipFirst bool, progressor *filesystem.ProgressReadWriter) (*VideoModel, error) {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(defaultWidth),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	sources, err := getSources(source, dest)
	if err != nil {
		return nil, err
	}

	tmpFolder, err := os.MkdirTemp("", "cleansync")
	if err != nil {
		return nil, err
	}

	vm := VideoModel{
		spinner:    s,
		progress:   p,
		progressor: progressor,
		skipFirst:  skipFirst,
		dest:       dest,
		sources:    sources,
		tempFolder: tmpFolder,
		ndx:        0,
	}

	return &vm, nil
}

func hasSupportedExtension(fileName string, extensions []string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}

func getSources(source string, dest string) ([]string, error) {
	var sources []string

	info, err := os.Stat(source)
	if err != nil {
		return sources, err
	}

	if info.IsDir() {
		err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if the file has an extension of .mp4 or .mkv
			// and is not a directory
			// We can add more extensions here if needed
			if !info.IsDir() && hasSupportedExtension(path, supportedExtensions) {
				sources = append(sources, path)
			}
			return nil

		})
		if err != nil {
			return sources, err
		}

		// Since this is a folder specified, we need to make sure the destiation is also a folder and not a file
		if !info.IsDir() {
			return sources, fmt.Errorf("destination/source mismatch. Please provide a folder for both source and destination")
		}

		return sources, err
	}

	// If it is not a directory supplied, assume it is a file
	sources = append(sources, source)
	return sources, nil
}

// Init is the entry point of the ui/program
func (m VideoModel) Init() tea.Cmd {
	return tea.Batch(m.ProcessVideoCmd(Starting, 0), m.spinner.Tick)
}

// View is the initial state of the ui
func (m VideoModel) View() string {

	// w := lipgloss.Width(fmt.Sprintf("%d", 1))

	if m.done {
		return doneStyle.Render("Done!")
	}

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog))
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(m.currentProcess)
	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
