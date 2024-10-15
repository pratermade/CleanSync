package sync

import (
	"cleansync/filesystem"
	"cleansync/localsql"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UploadModel struct {
	toUpdate       []string
	folderPath     string
	s3Client       *s3.Client
	db             *localsql.Sqldb
	bucket         string
	deep           bool
	index          int
	width          int
	height         int
	spinner        spinner.Model
	progress       progress.Model
	done           bool
	filters        []string
	currentProcess string
	indention      string
	progressor     *filesystem.ProgressReadWriter
}

var (
	doneStyle = lipgloss.NewStyle().Margin(1, 2)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	flagMark  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2C900")).SetString("⚑")
)

// NewModel initializes and returns a new model
func NewModel(folderPath string, client *s3.Client, fileList []string, bucket string, db *localsql.Sqldb, filters []string, progressor *filesystem.ProgressReadWriter, deep bool) UploadModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return UploadModel{
		spinner:    s,
		progress:   p,
		s3Client:   client,
		bucket:     bucket,
		deep:       deep,
		db:         db,
		index:      0,
		folderPath: folderPath,
		filters:    filters,
		toUpdate:   fileList,
		progressor: progressor,
	}
}

// Init is the entry point of the ui/program
func (m UploadModel) Init() tea.Cmd {
	ctx := context.Background()
	return tea.Batch(m.PutObjectCmd(ctx), m.spinner.Tick)
}

// View is the initial state of the ui
func (m UploadModel) View() string {
	n := len(m.toUpdate)
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
