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

type CachedImage struct {
	Thumbnail image.Image
	Preview   image.Image
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
			thumbnail BLOB,
			preview BLOB
		);

		CREATE TABLE IF NOT EXISTS image_bins (
			image_path TEXT NOT NULL,
			bin_id INTEGER NOT NULL,
			PRIMARY KEY (image_path, bin_id),
			FOREIGN KEY (image_path) REFERENCES thumbnails(path) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_iamge_bins_bin_id ON image_bins(bin_id);
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
	err := db.conn.QueryRow("SELECT thumbnail FROM thumbnails WHERE path = ?", path).Scan(&data)
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

func (db *DB) GetPreview(path string) (image.Image, bool) {
	var data []byte
	err := db.conn.QueryRow("SELECT preview FROM thumbnails WHERE path = ?", path).Scan(&data)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("error getting preview from DB for %s: %v", path, err)
		}
		return nil, false
	}

	img, err := jpeg.Decode(bytes.NewBuffer(data))
	if err != nil {
		log.Printf("error decoding preview for %s: %v", path, err)
	}

	return img, true
}

func (db *DB) SetImage(path string, imgData CachedImage) error {
	return db.SetImages(map[string]CachedImage{path: imgData})
}

func (db *DB) SetImages(images map[string]CachedImage) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	imgStmt, err := tx.Prepare("INSERT OR REPLACE INTO thumbnails (path, thumbnail, preview) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer imgStmt.Close()

	binStmt, err := tx.Prepare(`
		INSERT INTO image_bins (image_path, bin_id)
		SELECT ?, 0
		WHERE NOT EXISTS (SELECT 1 FROM image_bins WHERE image_path = ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer binStmt.Close()

	for path, imgData := range images {
		var thumBuf, previewBuf bytes.Buffer
		if err := jpeg.Encode(&thumBuf, imgData.Thumbnail, nil); err != nil {
			log.Printf("Error encoding thumbnail for %s: %v", path, err)
			continue
		}

		if err := jpeg.Encode(&previewBuf, imgData.Preview, nil); err != nil {
			log.Printf("Error encoding  preview for %s: %v", path, err)
			continue
		}

		if _, err := imgStmt.Exec(path, thumBuf.Bytes(), previewBuf.Bytes()); err != nil {
			log.Printf("Error executing img batch statement for %s: %v", path, err)
			continue
		}

		if _, err := binStmt.Exec(path, path); err != nil {
			log.Printf("Error executing bin batch statement for bin %s: %v", path, err)
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

// CopyImageToBin adds an image to a new bin without removing it from existing ones.
func (db *DB) AddImageToBin(path string, destID int) error {
	_, err := db.conn.Exec("INSERT OR IGNORE INTO image_bins (image_path, bin_id) VALUES (?, ?)", path, destID)
	return err
}

func (db *DB) UpdateImages(paths []string, sourceID, destID int) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE image_bins SET bin_id = ? WHERE image_path = ? AND bin_id = ?")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, path := range paths {
		if _, err := stmt.Exec(destID, path, sourceID); err != nil {
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
