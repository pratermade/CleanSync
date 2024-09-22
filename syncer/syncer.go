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

// GetUploadList queries the db and returns a slice of files that need updated.
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

// UploadDiffs uploads the files(paths) in the diffs slice, will commit to glacier deep archive if deep is set to true
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
		err := app.putObject(ctx, v, deep)
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

// UpdateManifest Updates the database for all the files (paths) specified in objs slice
func (app *Syncer) UpdateManifest(objs map[string]int64) error {

	for k, v := range objs {
		app.updateRecord(k, v)
	}
	return nil
}

// WalkAndHash walks the directory structure that is specifed in the Syncer.Folderpath.
// Will filter for filetypes listed in the filters slice.
// Returns a map of filepath[lastModDate]
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

// updateRecord updates or inserts an individual record with the p path and the last mod date specified by mod
// checks to see if the record needs updating first, only will update if the modified date has changed
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

// updateUploadStatus updates the status for the file specified with p.
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

// inFilters checks to see if the name of the file has one of the extensions listed in the filters slice, it returns true.
func inFilters(name string, filters []string) bool {
	for _, filter := range filters {
		if strings.HasSuffix(name, filter) {
			return true
		}
	}
	return false
}

// localize converts paths to windows paths if needed, has its own function for future needs.
func (app *Syncer) localize(s string) string {
	s = filepath.FromSlash(s)
	return s

}

// putObject actially performs the uploading to the S3 bucket for the file (path) specified by obj.
// if deep is true, will put it in glacier deep storage.
func (app *Syncer) putObject(ctx context.Context, obj string, deep bool) error {

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

// recordExists checks to see if there is a matching record for the provided p (file path) and modified time modtime.
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
