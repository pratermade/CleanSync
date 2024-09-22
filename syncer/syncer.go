package syncer

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pterm/pterm"
)

type Syncer struct {
	db         *sql.DB
	FolderPath string
	S3Client   *s3.Client
	Bucket     string
}

const INSERTRECORD = "insert into videos values(?, ?, ?)"
const UPSERTRECORD = "insert into videos values(?, ?, ?) on conflict(filepath) do update set (modified, uploaded) = (?,?)"
const SELECTRECORD = "select filepath from videos where filepath = ? and modified = ?"
const CREATETABLE = "create table videos (filepath text primary key, modified integer, uploaded integer)"
const UPDATEUPLOADSTATUS = "update videos set uploaded = 1 where filepath = ?"
const SELECTUPLOADLIST = "select filepath from videos where uploaded = false"

// InitDb gets the db if it already exists, if not it creates and preps a new one.
func (app *Syncer) InitDb(dbpath string) error {
	db, err := sql.Open("sqlite3", dbpath)

	if err != nil {
		return err
	}
	app.db = db
	if _, err := os.Stat(dbpath); errors.Is(err, os.ErrNotExist) {
		// db does not exist, c reate a new one

		_, err = app.db.Exec(CREATETABLE)
		if err != nil {
			return err
		}
	}
	// db exists, just set it
	return nil
}

func (app *Syncer) GetUploadList() ([]string, error) {
	rows, err := app.db.Query(SELECTUPLOADLIST)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var p string
		rows.Scan(&p)
		res = append(res, p)
	}

	return res, nil
}

func (app *Syncer) updateUploadStatus(p string) error {
	tx, err := app.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(UPDATEUPLOADSTATUS)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(p)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (app *Syncer) UploadDiffs(ctx context.Context, diffs []string, deep bool) error {
	count := len(diffs)
	if count == 0 {
		pterm.Success.Println("No files to update!")
		return nil
	}
	multi := pterm.DefaultMultiPrinter
	pb, err := pterm.DefaultProgressbar.WithTotal(count).WithWriter(multi.NewWriter()).Start("Uploading files")
	if err != nil {
		return err
	}
	multi.Start()
	for _, v := range diffs {
		pb.Increment()
		err := app.PutObject(ctx, v, deep)
		if err != nil {
			return err
		}
		err = app.updateUploadStatus(v)
		if err != nil {
			return err
		}
	}
	pb.Increment()
	multi.Stop()
	return nil
}

func (app *Syncer) UpdateManifest(objs map[string]int64) error {

	for k, v := range objs {
		app.updateRecord(k, v)
	}
	return nil
}

func (app *Syncer) updateRecord(p string, mod int64) error {
	tx, err := app.db.Begin()
	if err != nil {
		return err
	}

	query, err := tx.Prepare(UPSERTRECORD)
	if err != nil {
		return err
	}
	defer query.Close()

	exists, err := app.recordExists(p, mod)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = query.Exec(p, mod, 0, mod, 0)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (app *Syncer) GetBucketKeys(ctx context.Context) ([]string, error) {

	res, err := app.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(app.Bucket),
	})
	if err != nil {
		return nil, err
	}

	keys := []string{}
	for _, obj := range res.Contents {
		keys = append(keys, *obj.Key)
	}
	return keys, nil
}

func (app *Syncer) WalkAndHash(filters []string) (map[string]int64, error) {
	spinnerInfo, err := pterm.DefaultSpinner.Start("Taking inventory of existing files.")
	if err != nil {
		return nil, err
	}
	retMap := make(map[string]int64)
	err = filepath.Walk(app.FolderPath, func(p string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() {
				if !inFilters(info.Name(), filters) {
					return nil
				}
				h, err := getLastModDate(p)
				if err != nil {
					spinnerInfo.Fail(err)
					return err
				}
				p := app.localize(p)
				retMap[p] = h
			}

		}
		return nil
	})
	if err != nil {
		spinnerInfo.Fail(err)
		return nil, err
	}
	spinnerInfo.Success("Taking Inventory of local files.")
	return retMap, nil
}

func inFilters(name string, filters []string) bool {
	for _, filter := range filters {
		if strings.HasSuffix(name, filter) {
			return true
		}
	}
	return false
}
func (app *Syncer) localize(s string) string {
	s = filepath.FromSlash(s)
	return s

}

func (app *Syncer) PutObject(ctx context.Context, obj string, deep bool) error {

	f, err := os.Open(obj)
	if err != nil {
		return err
	}
	defer f.Close()

	storageClass := types.StorageClassStandard
	if deep {
		storageClass = types.StorageClassDeepArchive
	}
	_, err = app.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(app.Bucket),
		Key:          aws.String(app.localize(obj)),
		StorageClass: storageClass,
		Body:         f,
	})
	if err != nil {
		return err
	}
	return nil

}

func getLastModDate(f string) (int64, error) {
	fileinfo, err := os.Stat(f)
	if err != nil {
		return 0, err
	}
	atime := fileinfo.ModTime().Unix()
	return atime, nil
}

func (app *Syncer) recordExists(p string, modtime int64) (bool, error) {
	var res string
	err := app.db.QueryRow(SELECTRECORD, p, modtime).Scan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
