package splitter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const chunkSize = int64(2 * 1024 * 1024 * 1024) // 2GB
// const chunkSize = int64(20 * 1024 * 1024) // 2GB

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

	info.OrgFileSize = fileInfo.Size()

	buffer := make([]byte, chunkSize)

	partFile, err := os.Create(info.PartPath)
	if err != nil {
		return nil, err
	}
	defer partFile.Close()

	info.PreviousPart = filepath.Base(info.PartPath)
	n, err := file.ReadAt(buffer, info.Offset)
	if err != nil {
		if err == io.EOF {

			info.Eof = true
			info.Offset = 0
			_, err = partFile.Write(buffer[:n])
			if err != nil {
				return nil, err
			}
			return info, nil
		}
		return nil, err
	}
	_, err = partFile.Write(buffer[:n])
	if err != nil {
		return nil, err
	}
	info.Count++

	info.Offset = info.Offset + chunkSize
	info.PercentComplete = (float64(info.Offset) + float64(chunkSize)) / float64(info.OrgFileSize)

	return info, nil
}

// for {
// 	n, err := file.Read(buffer)
// 	if n > 0 {
// 		chunkFilePath := fmt.Sprintf("%s.part%d", filepath.Base(filePath), chunkIndex)
// 		chunkFilePath = filepath.Join(tmpDir, chunkFilePath)
// 		chunkFile, err := os.Create(chunkFilePath)
// 		if err != nil {
// 			retErr <- fmt.Errorf("failed to create chunk file: %v", err)
// 		}
// 		_, err = chunkFile.Write(buffer[:n])
// 		chunkFile.Close()
// 		if err != nil {
// 			retErr <- fmt.Errorf("failed to write chunk file: %v", err)
// 		}

// 		progress <- chunkFilePath
// 		chunkIndex++
// 	}
// 	if err == io.EOF {
// 		break
// 	}
// 	if err != nil {
// 		retErr <- fmt.Errorf("failed to read file: %v", err)
// 	}
// }
// retErr <- nil
// }

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
