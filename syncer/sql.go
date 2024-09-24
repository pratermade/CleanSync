package syncer

import (
	"database/sql"
	"errors"
	"os"
)

const CREATEVIDEOSTABLE = "create table videos (id integer primary key not null, filepath text unique, modified integer default (0), uploaded integer default (0), multipart integer default (0))"
const CREATEPARTSTABLE = "create table parts (id INTEGER PRIMARY KEY NOT NULL UNIQUE, video_id INTEGER NOT NULL, filepath TEXT UNIQUE, uploaded INTEGER DEFAULT (0))"

const UPSERTRECORD = "insert into videos (filepath, modified) values(?, ?) on conflict(filepath) do update set (modified, uploaded, multipart) = (?,?,?)"
const SELECTRECORD = "select filepath from videos where filepath = ? and modified = ?"
const SELECTVIDEOIDBBYPATH = "select id from videos where filepath = ?"
const UPDATEUPLOADSTATUS = "update videos set uploaded = 1 where filepath = ?"
const UPDATEUPLOADSTATUSPART = "update PARTS set uploaded = 1 where filepath = ?"
const SELECTUPLOADLIST = "select filepath from videos where uploaded = false"
const SETMULTIPART = "update videos set multipart = 1 where filepath = ?"
const INSERTPART = "insert into parts (video_id, filepath) values(?, ?)"

// InitDb gets the db if it already exists, if not it creates and preps a new one.
func (app *Syncer) InitDb(dbpath string) error {
	db, err := sql.Open("sqlite3", dbpath)

	if err != nil {
		return err
	}
	app.db = db
	if _, err := os.Stat(dbpath); errors.Is(err, os.ErrNotExist) {
		// db does not exist, c reate a new one

		_, err = app.db.Exec(CREATEVIDEOSTABLE)
		if err != nil {
			return err
		}
		_, err = app.db.Exec(CREATEPARTSTABLE)
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
	_, err = query.Exec(p, mod, mod, 0, 0)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// updateUploadStatuspart updates the status for the file specified with p.
func (app *Syncer) updateUploadStatusPart(p string) error {
	tx, err := app.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(UPDATEUPLOADSTATUSPART)
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

// recordParts inserts the split videos parts into the parts table
func (app Syncer) recordParts(videoid int, parts []string) error {
	tx, err := app.db.Begin()
	if err != nil {
		return err
	}
	for _, part := range parts {
		stmt, err := tx.Prepare(INSERTPART)
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(videoid, part)
		if err != nil {
			return err
		}
		stmt.Close()
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// SetMultipart sets the multipart flag in the vidoes table
func (app Syncer) setMultipart(fp string) (int, error) {
	tx, err := app.db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(SETMULTIPART)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(fp)
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	// Lets get the record Id
	var res int
	err = app.db.QueryRow(SELECTVIDEOIDBBYPATH, fp).Scan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return res, nil
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
