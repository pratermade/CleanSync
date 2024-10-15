package messages

type UploadMsg struct {
	Done bool
}

type UploadPartsMsg struct {
	OriginalFile string
	Parts        []string
	Index        int
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
