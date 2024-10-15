package filesystem

import (
	"bufio"
	"cleansync/messages"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const chunkSize = int64(16 * 1024 * 1024)

// WalkAndHash walks the directory structure that is specifed in the Syncer.Folderpath.
// Will filter for filetypes listed in the filters slice.
// Returns a map of filepath[lastModDate]
func WalkAndHash(filters []string, folderPath string) (map[string]int64, error) {

	retMap := make(map[string]int64)
	err := filepath.Walk(folderPath, func(p string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() {
				if !inFilters(info.Name(), filters) {
					return nil
				}
				h, err := getLastModDate(p)
				if err != nil {
					return err
				}
				p := Localize(p)
				retMap[p] = h
			}

		}
		return nil
	})
	if err != nil {
		// spinnerInfo.Fail(err)
		return nil, err
	}
	// spinnerInfo.Success("Taking Inventory of local files.")
	return retMap, nil
}

// localize converts paths to windows paths if needed, has its own function for future needs.
func Localize(s string) string {
	s = filepath.FromSlash(s)
	return s

}

// inFilters checks to see if the name of the file has one of the extensions listed in the filters slice, it returns true.
func inFilters(name string, filters []string) bool {
	for _, filter := range filters {
		if strings.HasSuffix(name, filter) {
			return true
		}
	}
	return false
}

// get lastModDate returns the last moidified date for the file specified by f (file path).
// Returns unix time
func getLastModDate(f string) (int64, error) {
	fileinfo, err := os.Stat(f)
	if err != nil {
		return 0, err
	}
	atime := fileinfo.ModTime().Unix()
	return atime, nil
}

type ProgressReadWriter struct {
	Writer    io.Writer
	Reader    io.Reader
	Size      int64
	Completed int64
}

func (pw *ProgressReadWriter) GetProgress(ch chan messages.ProgressMsg) {
	for {
		time.Sleep(250 * time.Millisecond)
		if pw.Completed != 0 {
			progress := float64(pw.Completed) / float64(pw.Size)
			ch <- messages.ProgressMsg{
				Progress: progress,
			}
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
	if err != nil {
		return 0, err
	}
	pr.Completed += int64(n)
	return n, err
}

func (pr *ProgressReadWriter) Copy(source string, dest string) error {

	buffer := make([]byte, chunkSize)

	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	pr.Size = fileSize
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	pr.Writer = bufio.NewWriter(destFile)

	var offset int64
	for {
		n, err := sourceFile.ReadAt(buffer, offset)
		if err != nil {
			if err == io.EOF {
				_, err = pr.Writer.Write(buffer[:n])
				if err != nil {
					return err
				}
				break
			}
			return err
		}
		n, err = pr.Writer.Write(buffer[:n])
		if err != nil {
			return err
		}
		offset = offset + int64(n)
		pr.Completed = offset
	}
	return nil
}
