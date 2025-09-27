package database

import (
	"bytes"
	"database/sql"
	"image"
	"image/jpeg"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	currentSchemaVersion = 1
	dbFileName           = ".picsort.db"
)

type DB struct {
	conn *sql.DB
}

func New(datasetPath string) (*DB, error) {
	dbPath := filepath.Join(datasetPath, dbFileName)
	conn, err := sql.Open("sqlite3", dbPath+"?_journal=WAL")
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() {
	db.conn.Close()
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS metadata (
			key TEXT PRIMARY KEY,
			value TEXT
		);
		CREATE TABLE IF NOT EXISTS thumbnails (
			path TEXT PRIMARY KEY,
			data BLOB
		);
		CREATE TABLE IF NOT EXISTS images (
			path TEXT PRIMERY KEY,
			bin INTEGER
		);
	`)
	if err != nil {
		return err
	}

	var version int
	err = db.conn.QueryRow("SELECT value FROM metadata WHERE key = 'schema_version'").Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if version < currentSchemaVersion {
		log.Printf("initializing schema version %d", currentSchemaVersion)
		_, err = db.conn.Exec("INSERT OR REPLACE INTO metadata (key, value) VALUES ('schema_version', ?)", currentSchemaVersion)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetThumbnail(path string) (image.Image, bool) {
	var data []byte
	err := db.conn.QueryRow("SELECT data FROM thumbnails WHERE path = ?", path).Scan(&data)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("error getting thumbnail from DB for %s: %v", path, err)
		}
		return nil, false
	}

	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("error decoding thumbnail for %s: %v", path, err)
	}

	return img, true
}

func (db *DB) SetThumbnail(path string, img image.Image) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		log.Printf("error encoding thumbnail for %s: %v", path, err)
		return
	}

	_, err := db.conn.Exec("INSERT OR REPLACE INTO thumbnails (path, data) VALUES (?, ?)", path, buf.Bytes())
	if err != nil {
		log.Printf("error setting thumbnail in DB for %s: %v", path, err)
	}
}

func (db *DB) SetThumbnailsBatch(thumbnails map[string]image.Image) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO thumbnails (path, data) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for path, img := range thumbnails {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, nil); err != nil {
			log.Printf("Error encoding thumbnail for %s: %v", path, err)
			continue
		}

		_, err := stmt.Exec(path, buf.Bytes())
		if err != nil {
			log.Printf("Error executing batch statement for %s: %v", path, err)
			continue
		}
	}

	return tx.Commit()
}
