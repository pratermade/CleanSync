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
	OrgFilePath string
	Name        string
	Offset      int64
	Eof         bool
	Count       int
	TempFolder  string
}

func SplitFile(info *SplitInfo) (*SplitInfo, error) {
	splitInfoRes := &SplitInfo{
		Eof:         false,
		Name:        "",
		Count:       info.Count,
		TempFolder:  info.TempFolder,
		Offset:      info.Offset,
		OrgFilePath: info.OrgFilePath,
	}
	if info.TempFolder == "" {
		tmpDir, err := os.MkdirTemp("", "s3sync")
		if err != nil {
			return nil, err
		}
		splitInfoRes.TempFolder = tmpDir
	}
	file, err := os.Open(splitInfoRes.OrgFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, chunkSize)

	partPath := fmt.Sprintf("%s.part%d", filepath.Base(splitInfoRes.OrgFilePath), info.Count)
	partPath = filepath.Join(splitInfoRes.TempFolder, partPath)
	partFile, err := os.Create(partPath)
	if err != nil {
		return nil, err
	}
	defer partFile.Close()
	n, err := file.ReadAt(buffer, info.Offset)
	if err != nil {
		if err == io.EOF {

			splitInfoRes.Eof = true
			splitInfoRes.Offset = 0
			splitInfoRes.Name = partPath
			_, err = partFile.Write(buffer[:n])
			if err != nil {
				return nil, err
			}
			return splitInfoRes, nil
		}
		return nil, err
	}
	_, err = partFile.Write(buffer[:n])
	if err != nil {
		return nil, err
	}
	splitInfoRes.Name = partPath
	splitInfoRes.Offset = splitInfoRes.Offset + chunkSize
	splitInfoRes.Count++
	return splitInfoRes, nil
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
