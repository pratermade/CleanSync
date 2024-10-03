package ui

import (
	"bufio"
	"context"
	"os"
	"s3sync/filesystem"
	"s3sync/messages"
	"s3sync/splitter"

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
	return func() tea.Msg {

		res, err := splitter.SplitFile(info, m.progressor)
		if err != nil {
			return errMsg{err}
		}
		return res

	}
}

// putObject actially performs the uploading to the S3 bucket for the file (path) specified by obj.
// if deep is true, will put it in glacier deep storage.
// Here is where the logic will live that will split files if they are too big
func (m *UploadModel) PutObject(ctx context.Context, obj string, pr *messages.ProgressReadWriter) tea.Cmd {

	return func() tea.Msg {

		// First check to see if it needs split up.
		info, err := os.Stat(obj)
		if err != nil {
			return errMsg{err}
		}

		if info.Size() > 4294967296 {
			return messages.SplitMsg{
				OrgFilePath: obj,
				OrgFileSize: info.Size(),
			}
		}

		f, err := os.Open(obj)
		if err != nil {
			return errMsg{err}
		}

		defer f.Close()

		pr.Reader = bufio.NewReader(f)
		pr.Size = info.Size()

		storageClass := types.StorageClassStandard
		if m.deep {
			storageClass = types.StorageClassDeepArchive
		}

		_, err = m.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(m.bucket),
			Key:           aws.String(filesystem.Localize(obj)),
			StorageClass:  storageClass,
			Body:          pr,
			ContentLength: &pr.Size,
		})
		if err != nil {
			return errMsg{err}
		}

		dmsg := messages.UploadMsg{
			Name: obj,
			Done: false,
		}

		return dmsg
	}
}

func (m *UploadModel) uploadParts(ctx context.Context, parts []string) tea.Cmd {
	return func() tea.Msg {
		uploadPartsM := messages.UploadPartsMsg{
			Parts: parts,
		}
		return uploadPartsM
	}
}
