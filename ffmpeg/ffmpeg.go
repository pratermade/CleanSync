package ffmpeg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Chapter struct {
	start float64
	end   float64
	title string
}

type Video struct {
	videoBaseName string
	videoExt      string
	filePath      string
	TmpFolder     string
	chapters      []Chapter
}

func NewVideo(filePath string, tmpFolder string) (Video, error) {

	vidname := filepath.Base(filePath)
	// if the video name contains a single quote, remove it and rename the file without the quote
	if strings.Contains(vidname, "'") {
		newVidName := strings.Replace(vidname, "'", "", -1)
		newVidPath := filepath.Join(filepath.Dir(filePath), newVidName)
		err := os.Rename(filePath, newVidPath)
		if err != nil {
			return Video{}, fmt.Errorf("error renaming file: %s", err)
		}
		filePath = newVidPath
		vidname = newVidName
	}

	ext := filepath.Ext(vidname)

	video := Video{
		filePath:      filePath,
		TmpFolder:     tmpFolder,
		videoBaseName: strings.Replace(vidname, ext, "", -1),
		videoExt:      ext,
	}

	err := video.getChapterInfo()
	if err != nil {
		return Video{}, err
	}
	return video, nil
}

func (v *Video) getChapterInfo() error {
	var out bytes.Buffer
	args := []string{
		"-i",
		v.filePath,
		"-show_chapters",
		"-loglevel",
		"error",
	}
	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	mwriter := io.MultiWriter(f, &out)
	cmd := exec.Command("ffprobe", args...)
	cmd.Stdout = mwriter
	cmd.Stderr = f
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error running ffprobe with the args: %s, %s", args, err)
	}
	chapterNum := 0
	for {
		line, err := out.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading chapter information: %s", err)
		}
		line = strings.TrimRight(line, "\r\n\t")
		if line == "[CHAPTER]" { // then the next lines are the chapter information
			v.chapters = append(v.chapters, Chapter{})
			for {
				chaptline, err := out.ReadString('\n')
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("error reading chapter line: %s", err)
				}
				chaptline = strings.ToLower(strings.TrimRight(chaptline, "\r\n\t"))
				if strings.HasPrefix(chaptline, "start_time=") {
					parts := strings.Split(chaptline, "=")
					if len(parts) != 2 {
						return fmt.Errorf("malformed start_time chapter information")
					}
					v.chapters[chapterNum].start, err = strconv.ParseFloat(parts[1], 64)
					if err != nil {
						return fmt.Errorf("unable to parse start information")
					}
				}
				if strings.HasPrefix(chaptline, "end_time=") {
					parts := strings.Split(chaptline, "=")
					if len(parts) != 2 {
						return fmt.Errorf("malformed end_time chapter information")
					}
					v.chapters[chapterNum].end, err = strconv.ParseFloat(parts[1], 64)
					if err != nil {
						return fmt.Errorf("unable to parse end information")
					}
				}
				if strings.HasPrefix(chaptline, "tag:title=") {
					parts := strings.Split(chaptline, "=")
					if len(parts) != 2 {
						return fmt.Errorf("malformed title chapter information")
					}
					v.chapters[chapterNum].title = parts[1]
					break
				}
			}
			chapterNum++
		}
	}
	return nil
}

func (v *Video) Recut(ndxs []int) (string, error) {
	// remove the single quotes from the v.videoBaseName
	videoBaseName := strings.Replace(v.videoBaseName, "'", "", -1)
	tempVideo := filepath.Join(v.TmpFolder, fmt.Sprintf("%s%s", videoBaseName, v.videoExt))
	if ndxs == nil {
		// just copy the file to the temp location
		err := os.MkdirAll(v.TmpFolder, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("error creating temp folder: %s", err)
		}

		srcFile, err := os.Open(v.filePath)
		if err != nil {
			return "", fmt.Errorf("error opening source file: %s", err)
		}
		defer srcFile.Close()

		destFile, err := os.Create(tempVideo)
		if err != nil {
			return "", fmt.Errorf("error creating destination file: %s", err)
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return "", fmt.Errorf("error copying file to temp folder: %s", err)
		}

		return tempVideo, nil
	}
	// Create a text document used to reassemble:
	concatString := ""

	for _, ndx := range ndxs {
		concatString = fmt.Sprintf("%sfile '%s'\ninpoint %f\noutpoint %f\n", concatString, v.filePath, v.chapters[ndx].start, v.chapters[ndx].end)
	}
	concatFile := filepath.Join(v.TmpFolder, "concat.txt")
	err := os.WriteFile(concatFile, []byte(concatString), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing concat file: %s", err)
	}

	err = runFFmpegCommand([]string{"-y", "-f", "concat", "-safe", "0", "-i", concatFile, "-c", "copy", "-map", "0", tempVideo})
	if err != nil {
		return "", fmt.Errorf("error concatenating parts: %s", err)
	}

	return tempVideo, nil
}

func (v *Video) GetNonAdIndexes(skipFirst bool) []int {
	if len(v.chapters) == 0 {
		return nil
	}
	var nonads []int
	for i, chap := range v.chapters {
		if strings.ToLower(chap.title) != "advertisement" {
			nonads = append(nonads, i)
		}
	}
	if skipFirst {
		return nonads[1:]
	}

	return nonads
}

func runFFmpegCommand(args []string) error {
	// fmt.Println("running: ffmpeg: ", args)
	cmd := exec.Command("ffmpeg", args...)

	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd.Stdout = f
	cmd.Stderr = f

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
