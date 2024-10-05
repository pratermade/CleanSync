package splitter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"s3sync/messages"
)

const chunkSize = int64(16 * 1024 * 1024) // 16mb chunks x 128 iterations = 2GB pieces
const partSize = 2 * 1024 * 1024 * 1024   // 2GB

// func (pw *ProgressWriter) GetProgress(ch chan messages.ProgressMsg) {
// 	for {
// 		time.Sleep(250 * time.Millisecond)
// 		if pw.writen != 0 {
// 			progress := float64(pw.writen) / float64(pw.toWrite)
// 			ch <- messages.ProgressMsg{
// 				Progress: progress,
// 				Action:   messages.WriteAction,
// 			}
// 			continue
// 		}
// 		ch <- messages.ProgressMsg{
// 			Progress: 0.0,
// 			Action:   messages.None,
// 		}
// 	}
// }

// func (pw *ProgressWriter) ResetProgress() {
// 	pw.toWrite = 0
// 	pw.writen = 0
// }

// type ProgressWriter struct {
// 	writer  io.Writer
// 	toWrite int64
// 	writen  int64
// }

// func (pw *ProgressWriter) Write(p []byte) (n int, err error) {

// 	n, err = pw.writer.Write(p)
// 	if err != nil {
// 		return 0, err
// 	}
// 	pw.writen += int64(n)
// 	return n, err
// }

type SplitInfo struct {
	OrgFilePath     string
	Parts           []string
	Offset          int64
	Eof             bool
	TempFolder      string
	PercentComplete float64
	OrgFileSize     int64
	Index           int
}

func SplitFile(info *SplitInfo, pw *messages.ProgressReadWriter) (*SplitInfo, error) {

	file, err := os.Open(info.OrgFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	info.OrgFileSize = fileInfo.Size()

	partFile, err := os.Create(info.Parts[info.Index])
	if err != nil {
		return nil, err
	}
	defer partFile.Close()

	pw.Writer = bufio.NewWriter(partFile)
	pw.ResetProgress()

	for i := 0; i < 128; i++ { // 16 MB chunks * 128 iterations = 2GB files
		n, err := writeChunk(file, pw, info.Offset, info.OrgFileSize)
		if err != nil {
			if err == io.EOF {
				// done, reached the end of the file
				info.Eof = true
				info.Offset = 0
				return info, nil
			}
			return nil, err
		}
		info.Offset = info.Offset + int64(n)
	}
	info.Index++
	return info, nil
}

func writeChunk(reader *os.File, writer *messages.ProgressReadWriter, offset int64, fullSize int64) (int, error) {
	sizeLeft := fullSize - offset
	writer.Size = sizeLeft % partSize
	if sizeLeft >= partSize {
		writer.Size = partSize
	}
	buffer := make([]byte, chunkSize)
	n, err := reader.ReadAt(buffer, offset)
	if err != nil {
		if err == io.EOF {
			eoferr := err
			n, err = writer.Write(buffer[:n])
			if err != nil {
				return 0, err
			}
			return n, eoferr
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
