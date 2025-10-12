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

		CREATE TABLE IF NOT EXISTS image_bins (
			image_path TEXT NOT NULL,
			bin_id INTEGER NOT NULL,
			PRIMARY KEY (image_path, bin_id),
			FOREIGN KEY (image_path) REFERENCES thumbnails(path) ON DELETE CASCADE
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

func (db *DB) SetThumbnailsBatch(thumbnails map[string]image.Image) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	thumbStmt, err := tx.Prepare("INSERT OR REPLACE INTO thumbnails (path, data) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer thumbStmt.Close()

	binStmt, err := tx.Prepare("INSERT OR IGNORE INTO image_bins (image_path, bin_id) VALUES (?, 0)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer binStmt.Close()

	for path, img := range thumbnails {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, nil); err != nil {
			log.Printf("Error encoding thumbnail for %s: %v", path, err)
			continue
		}

		if _, err := thumbStmt.Exec(path, buf.Bytes()); err != nil {
			log.Printf("Error executing batch statement for %s: %v", path, err)
			continue
		}

		if _, err := binStmt.Exec(path); err != nil {
			log.Printf("Error executing batch statement for bin %s: %v", path, err)
			continue
		}
	}

	return tx.Commit()
}

func (db *DB) GetImagePaths(binID int) ([]string, error) {
	var rows *sql.Rows
	var err error
	if binID == -1 {
		rows, err = db.conn.Query("SELECT path FROM thumbnails")
	} else {
		rows, err = db.conn.Query("SELECT image_path FROM image_bins WHERE bin_id = ?", binID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}

// MoveImage updates the bin for a given image.
func (db *DB) UpdateImage(path string, destID int) error {
	_, err := db.conn.Exec("UPDATE image_bins SET bin_id = ? WHERE image_path  = ?", destID, path)
	return err
}

// CopyImageToBin adds an image to a new bin without removing it from existing ones.
func (db *DB) AddImageToBin(path string, destID int) error {
	_, err := db.conn.Exec("INSERT OR IGNORE INTO image_bins (image_path, bin_id) VALUES (?, ?)", path, destID)
	return err
}

func (db *DB) UpdateImages(paths []string, destID int) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE image_bins SET bin_id = ? WHERE image_path = ?")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, path := range paths {
		if _, err := stmt.Exec(destID, path); err != nil {
			log.Printf("Error executing batch update for %s: %v", path, err)
		}
	}

	return tx.Commit()
}

func (db *DB) AddImagesToBin(paths []string, destID int) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO image_bins (image_path, bin_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, path := range paths {
		if _, err := stmt.Exec(path, destID); err != nil {
			log.Printf("Error executing batch insert for %s: %v", path, err)
		}
	}

	return tx.Commit()
}
