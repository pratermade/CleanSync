package sync

import (
	"bufio"
	"cleansync/filesystem"
	"cleansync/messages"
	"cleansync/splitter"
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m *UploadModel) SendError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err}
	}
}

func (m *UploadModel) splitCmd(info *splitter.SplitInfo) tea.Cmd {
	if len(info.Parts) == 0 {
		return func() tea.Msg {
			return info
		}
	}
	return func() tea.Msg {
		res, err := splitter.SplitFile(info, m.progressor)
		if err != nil {
			return errMsg{err}
		}
		return res

	}
}

func (m *UploadModel) uploadCmd(info *messages.UploadMsg) tea.Cmd {
	return func() tea.Msg {
		return info
	}
}

// putObject actially performs the uploading to the S3 bucket for the file (path) specified by obj.
// if deep is true, will put it in glacier deep storage.
// Here is where the logic will live that will split files if they are too big
func (m *UploadModel) PutObjectCmd(ctx context.Context) tea.Cmd {

	return func() tea.Msg {
		dmsg := messages.UploadMsg{
			Done: false,
		}

		return dmsg
	}
}

func (m *UploadModel) doUpload(ctx context.Context, partFilePath string, pr *filesystem.ProgressReadWriter, fileSize int64) error {
	f, err := os.Open(partFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	pr.Reader = bufio.NewReader(f)

	storageClass := types.StorageClassStandard
	if m.deep {
		storageClass = types.StorageClassDeepArchive
	}
	m.progressor.ResetProgress()
	pr.Size = fileSize

	_, err = m.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(m.bucket),
		Key:           aws.String(filesystem.Localize(partFilePath)),
		StorageClass:  storageClass,
		Body:          pr,
		ContentLength: &pr.Size,
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *UploadModel) uploadParts(ctx context.Context, orgFile string, parts []string, i int) tea.Cmd {
	return func() tea.Msg {
		if i >= len(parts) {
			return errMsg{}
		}
		if len(parts) > 0 {
			f, err := os.Open(parts[i])
			if err != nil {
				return errMsg{err}
			}
			info, err := f.Stat()
			if err != nil {
				return errMsg{err}
			}

			err = m.doUpload(ctx, parts[i], m.progressor, info.Size())
			if err != nil {
				return errMsg{err}
			}
		}

		uploadPartsM := messages.UploadPartsMsg{
			OriginalFile: orgFile,
			Parts:        parts,
			Index:        i,
		}
		return uploadPartsM
	}
}
