package splitter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"s3sync/messages"
	"time"
)

const chunkSize = int64(16 * 1024 * 1024) // 16mb chunks x 128 iterations = 2GB pieces
// const chunkSize = int64(20 * 1024 * 1024) // 2GB

func (pr *ProgressWriter) GetProgress(ch chan messages.ProgressMsg) {
	for {
		time.Sleep(time.Second)
		if pr.writen != 0 {
			progress := float64(pr.writen) / float64(pr.total)
			ch <- messages.ProgressMsg{
				Progress: progress,
			}
			continue
		}
		ch <- messages.ProgressMsg{Progress: 0.0}
	}
}

type ProgressWriter struct {
	writer io.Writer
	total  int64
	writen int64
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.writen += int64(n)
	return n, err
}

type SplitInfo struct {
	OrgFilePath     string
	PartPath        string
	PreviousPart    string
	Offset          int64
	Eof             bool
	Count           int
	TempFolder      string
	PercentComplete float64
	OrgFileSize     int64
}

func SplitFile(info *SplitInfo) (*SplitInfo, error) {

	file, err := os.Open(info.OrgFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	pw := &ProgressWriter{
		writer: bufio.NewWriter(file),
		total:  fileInfo.Size(),
	}

	info.OrgFileSize = fileInfo.Size()

	partFile, err := os.Create(info.PartPath)
	if err != nil {
		return nil, err
	}
	defer partFile.Close()

	info.PreviousPart = filepath.Base(info.PartPath)
	for i := 0; i < 120; i++ { // 16 MB chunks * 128 iterations = 2GB files
		n, err := writeChunk(file, pw, info.Offset, info.PartPath)
		if err != nil {
			if err == io.EOF {
				// done, reached the end of the file
				info.Eof = true
				info.Offset = 0
				break
			}
			return nil, err
		}
		info.Offset = info.Offset + int64(n)
	}
	info.Count++
	return info, nil
}

func writeChunk(reader *os.File, writer *ProgressWriter, offset int64, outfile string) (int, error) {
	buffer := make([]byte, chunkSize)
	n, err := reader.ReadAt(buffer, offset)
	if err != nil {
		if err == io.EOF {
			n, err = writer.Write(buffer[:n])
			if err != nil {
				return 0, err
			}
			return n, nil
		}
		return 0, err
	}
	n, err = writer.Write(buffer[:n])
	if err != nil {
		return 0, err
	}
	return n, nil

}

func RecombineFile(partPrefix string) (string, error) {

	combinedFile, err := os.Create(partPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create combined file: %v", err)
	}
	defer combinedFile.Close()

	var partIndex int
	for {
		partFile := fmt.Sprintf("%s.part%d", partPrefix, partIndex)
		file, err := os.Open(partFile)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to open part file: %v", err)
		}

		_, err = io.Copy(combinedFile, file)
		file.Close()
		if err != nil {
			return "", fmt.Errorf("failed to copy part file: %v", err)
		}
		partIndex++
	}

	return partPrefix, nil
}

func CleanUp(objs []string) error {
	if len(objs) < 1 {
		return fmt.Errorf("nothing to clean up")
	}
	folder := filepath.Base(objs[0])
	err := os.RemoveAll(folder)
	if err != nil {
		return err
	}
	return nil
}
