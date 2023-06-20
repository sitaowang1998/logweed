package metadata

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type MetaMySQL struct {
	conn *sql.DB
}

func (db *MetaMySQL) Connect(usr string, pwd string, addr string) error {
	var err error
	db.conn, err = sql.Open("mysql",
		fmt.Sprintf("%v:%v@tcp(%v)/clp_meta", usr, pwd, addr))
	if err != nil {
		log.Print(err)
		return err
	}
	err = db.conn.Ping()
	if err != nil {
		log.Print(err)
	}
	return err
}

func (db *MetaMySQL) Close() error {
	if db.conn == nil {
		err := errors.New("no valid connection")
		log.Print(err)
		return err
	}
	err := db.conn.Close()
	if err != nil {
		log.Print(err)
	}
	return err
}

var qCreateArchives = `CREATE TABLE IF NOT EXISTS archives (
    archive_id BIGINT NOT NULL AUTO_INCREMENT,
    uncompressed_size BIGINT UNSIGNED,
    size BIGINT UNSIGNED,
    fid VARCHAR(33),
    PRIMARY KEY (archive_id)
);`

var qCreateFiles = `CREATE TABLE IF NOT EXISTS files (
    id BIGINT NOT NULL AUTO_INCREMENT,
    file_path VARCHAR(255) NOT NULL UNIQUE,
    tag VARCHAR(255),
    begin_timestamp BIGINT UNSIGNED,
    end_timestamp BIGINT UNSIGNED,
    archive_id BIGINT,
    uncompressed_bytes BIGINT UNSIGNED,
    num_messages BIGINT UNSIGNED,
    PRIMARY KEY (id),
    INDEX (tag),
    FOREIGN KEY (archive_id) REFERENCES archives(archive_id)
);`

func (db *MetaMySQL) InitService() error {
	if db.conn == nil {
		err := errors.New("no valid connection")
		log.Print(err)
		return err
	}
	_, err := db.conn.Exec(qCreateArchives)
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = db.conn.Exec(qCreateFiles)
	if err != nil {
		log.Print(err)
	}
	return err
}

var qListTags = `SELECT DISTINCT tag FROM files;`

func (db *MetaMySQL) ListTags() ([]string, error) {
	if db.conn == nil {
		err := errors.New("no valid connection")
		log.Print(err)
		return nil, err
	}
	rows, err := db.conn.Query(qListTags)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()
	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

var qGetFiles = `SELECT file_path FROM files WHERE tag = ?;`

func (db *MetaMySQL) GetFiles(tag string) ([]string, error) {
	if db.conn == nil {
		err := errors.New("no valid connection")
		log.Print(err)
		return nil, err
	}
	rows, err := db.conn.Query(qGetFiles, tag)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()
	files := make([]string, 0)
	for rows.Next() {
		var file string
		err := rows.Scan(&file)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

var qAddArchive = `INSERT INTO archives(uncompressed_size, size, fid) VALUES (?, ?, ?)`
var qAddFile = `INSERT INTO files(file_path, tag, begin_timestamp, end_timestamp, archive_id, uncompressed_bytes, num_messages) VALUES (?, ?, ?, ?, ?, ?, ?)`

func (db *MetaMySQL) AddMetadata(archives []ArchiveMetadata, files []FileMetadata) error {
	if db.conn == nil {
		err := errors.New("no valid connection")
		log.Print(err)
		return err
	}
	archiveIDs := make([]int64, 0, len(archives))
	tx, err := db.conn.Begin()
	if err != nil {
		log.Print(err)
		return err
	}
	archiveStmt, err := tx.Prepare(qAddArchive)
	if err != nil {
		log.Print(err)
		return err
	}
	// defering the close of prepared statement only works after go 1.4
	defer archiveStmt.Close()
	for _, archive := range archives {
		res, err := archiveStmt.Exec(archive.UncompressedSize, archive.Size, archive.Fid)
		if err != nil {
			log.Print(err)
			tx.Rollback()
			return err
		}
		archiveID, err := res.LastInsertId()
		if err != nil {
			log.Print(err)
			tx.Rollback()
			return err
		}
		archiveIDs = append(archiveIDs, archiveID)
	}
	fileStmt, err := tx.Prepare(qAddFile)
	if err != nil {
		log.Print(err)
		return err
	}
	defer fileStmt.Close()
	for _, file := range files {
		archiveID := archiveIDs[file.ArchiveID]
		_, err := fileStmt.Exec(file.FilePath, file.Tag, file.BeginTimestamp, file.EndTimestamp, archiveID, file.UncompressedBytes, file.NumMessages)
		if err != nil {
			log.Print(err)
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

var qSearch = `SELECT archives.fid FROM archives INNER JOIN (SELECT DISTINCT archive_id FROM files WHERE tag = ? AND ? <= end_timestamp AND ? >= begin_timestamp) AS t ON archives.archive_id = t.archive_id`

func (db *MetaMySQL) Search(tag string, beginTimestamp uint64, endTimestamp uint64) ([]string, error) {
	if db.conn == nil {
		err := errors.New("no valid connection")
		log.Print(err)
		return nil, err
	}
	fids := make([]string, 0)
	rows, err := db.conn.Query(qSearch, tag, beginTimestamp, endTimestamp)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var fid string
		err := rows.Scan(&fid)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		fids = append(fids, fid)
	}
	return fids, nil
}
