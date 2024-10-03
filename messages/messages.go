package messages

import (
	"io"
	"time"
)

type UploadMsg struct {
	Name string
	Done bool
}

type UploadPartsMsg struct {
	Parts []string
	Done  bool
}

type SplitMsg struct {
	OrgFilePath string
	OrgFileSize int64
}

type ErrMsg struct {
	err string
}

type SplittingFileMsg struct {
	name        string
	offset      int64
	eof         bool
	count       int
	tempFolder  string
	orgFilePath string
	orgFileSize int64
}

type ProgressMsg struct {
	Progress float64
	// Action   int
}

const (
	None = iota
	WriteAction
	UploadAction
)

type ProgressReadWriter struct {
	Writer    io.Writer
	Reader    io.Reader
	Size      int64
	Completed int64
}

func (pw *ProgressReadWriter) GetProgress(ch chan ProgressMsg) {
	for {
		time.Sleep(250 * time.Millisecond)
		if pw.Completed != 0 {
			progress := float64(pw.Completed) / float64(pw.Size)
			ch <- ProgressMsg{
				Progress: progress,
				// Action:   action,
			}
			continue
		}
		ch <- ProgressMsg{
			Progress: 0.0,
			// Action:   None,
		}
	}
}

func (pw *ProgressReadWriter) ResetProgress() {
	pw.Size = 0
	pw.Completed = 0
}

func (pw *ProgressReadWriter) Write(p []byte) (n int, err error) {

	n, err = pw.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	pw.Completed += int64(n)
	return n, err
}

func (pr *ProgressReadWriter) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Completed += int64(n)
	return n, err
}
