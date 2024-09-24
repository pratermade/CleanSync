package splitter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const chunkSize = 2 * 1024 * 1024 * 1024 // 2GB

func SplitFile(filePath string) ([]string, error) {
	var files []string
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var chunkIndex int
	buffer := make([]byte, chunkSize)
	tmpDir, err := os.MkdirTemp("", "s3sync")
	if err != nil {
		return nil, err
	}
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			chunkFilePath := fmt.Sprintf("%s.part%d", filepath.Base(filePath), chunkIndex)

			chunkFilePath = filepath.Join(tmpDir, chunkFilePath)
			chunkFile, err := os.Create(chunkFilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to create chunk file: %v", err)
			}
			_, err = chunkFile.Write(buffer[:n])
			chunkFile.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to write chunk file: %v", err)
			}
			files = append(files, chunkFilePath)
			chunkIndex++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}
	}

	return files, nil
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
