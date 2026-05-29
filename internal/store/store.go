package store

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// viewKey identifies a unique asset/domain pair for view counting.
type viewKey struct {
	assetID string
	domain  string
}

// Store wraps the SQLite database and provides DAL methods for assets.
type Store struct {
	db *sql.DB

	viewMu       sync.Mutex
	pendingViews map[viewKey]int
}

// Image holds a row from the art_pieces table.
type Image struct {
	ID          string
	ImgchestURL string
	Album       string
	CreatedAt   string
	ContentType string
}

// AlbumInfo is returned by ListAlbums.
type AlbumInfo struct {
	Album            string `json:"album"`
	Count            int    `json:"count"`
	Cover            string `json:"cover,omitempty"`
	CoverContentType string `json:"coverContentType,omitempty"`
	Position         int    `json:"position"`
}

// ImageInfo is returned by ListImages.
type ImageInfo struct {
	ID          string `json:"id"`
	Views       int    `json:"views"`
	Position    int    `json:"position"`
	ContentType string `json:"contentType,omitempty"`
}

// ReferrerView is a single referrer breakdown for an asset.
type ReferrerView struct {
	Domain    string `json:"domain"`
	Count     int    `json:"count"`
	FirstSeen int64  `json:"first_seen"`
	LastSeen  int64  `json:"last_seen"`
}

// AssetStats aggregates views for a single asset.
type AssetStats struct {
	Total     int            `json:"total"`
	Referrers []ReferrerView `json:"referrers"`
}

// New opens (or creates) the SQLite database and migrates the schema.
func New(dbPath string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_txlock=immediate", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	s := &Store{
		db:           db,
		pendingViews: make(map[viewKey]int),
	}
	go s.viewFlusher()

	return s, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB exposes the underlying *sql.DB for other managers that share it.
func (s *Store) DB() *sql.DB {
	return s.db
}

// FindImage looks up an asset by its short ID.
func (s *Store) FindImage(id string) (*Image, error) {
	row := s.db.QueryRow(
		`SELECT id, imgchest_url, album, created_at, content_type FROM art_pieces WHERE id = ?`, id)
	var img Image
	var ct sql.NullString
	err := row.Scan(&img.ID, &img.ImgchestURL, &img.Album, &img.CreatedAt, &ct)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if ct.Valid {
		img.ContentType = ct.String
	}
	return &img, nil
}

// SetImageContentType updates the content type for an asset.
func (s *Store) SetImageContentType(id string, ct string) error {
	_, err := s.db.Exec(`UPDATE art_pieces SET content_type = ? WHERE id = ?`, ct, id)
	return err
}

// ImportImages creates new art_piece rows for a batch of URLs in an album.
func (s *Store) ImportImages(album string, urls []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	prefix := strings.ToLower(strings.ReplaceAll(album, " ", "-"))
	now := time.Now().Format("2006-01-02 15:04:05")

	var nextPos int
	_ = tx.QueryRow(`SELECT COALESCE(MAX(position), 0) + 1 FROM art_pieces WHERE album = ?`, album).Scan(&nextPos)

	stmt, err := tx.Prepare(
		`INSERT INTO art_pieces (id, imgchest_url, album, created_at, position) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, url := range urls {
		id := fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), i)
		if _, err := stmt.Exec(id, url, album, now, nextPos+i); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ListAlbums returns every album with its image count, cover thumbnail, and position.
func (s *Store) ListAlbums() ([]AlbumInfo, error) {
	rows, err := s.db.Query(
		`SELECT a.album, COUNT(*),
			(SELECT id FROM art_pieces WHERE album = a.album ORDER BY position ASC, created_at DESC LIMIT 1),
			(SELECT content_type FROM art_pieces WHERE album = a.album ORDER BY position ASC, created_at DESC LIMIT 1),
			COALESCE(m.position, 0)
		 FROM art_pieces a
		 LEFT JOIN albums_meta m ON a.album = m.album
		 GROUP BY a.album
		 ORDER BY COALESCE(m.position, 0) ASC, a.album ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	albums := []AlbumInfo{}
	for rows.Next() {
		var a AlbumInfo
		var ct sql.NullString
		if err := rows.Scan(&a.Album, &a.Count, &a.Cover, &ct, &a.Position); err != nil {
			return nil, err
		}
		if ct.Valid {
			a.CoverContentType = ct.String
		}
		albums = append(albums, a)
	}
	return albums, rows.Err()
}

// ListImages returns all images in an album with their total view counts, ordered by position.
func (s *Store) ListImages(album string) ([]ImageInfo, error) {
	rows, err := s.db.Query(
		`SELECT a.id, COALESCE(SUM(v.count), 0), a.position, a.content_type FROM art_pieces a
		 LEFT JOIN asset_views v ON a.id = v.asset_id
		 WHERE a.album = ? GROUP BY a.id ORDER BY a.position ASC, a.created_at DESC`, album)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := []ImageInfo{}
	for rows.Next() {
		var img ImageInfo
		var ct sql.NullString
		if err := rows.Scan(&img.ID, &img.Views, &img.Position, &ct); err != nil {
			return nil, err
		}
		if ct.Valid {
			img.ContentType = ct.String
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

// DeleteAlbum removes every asset in an album.
func (s *Store) DeleteAlbum(album string) error {
	_, err := s.db.Exec(`DELETE FROM art_pieces WHERE album = ?`, album)
	return err
}

// DeleteImage removes a single asset.
func (s *Store) DeleteImage(id string) error {
	_, err := s.db.Exec(`DELETE FROM art_pieces WHERE id = ?`, id)
	return err
}

// UpdateImageSource updates the imgchest_url for a single asset.
func (s *Store) UpdateImageSource(id string, url string) error {
	_, err := s.db.Exec(`UPDATE art_pieces SET imgchest_url = ? WHERE id = ?`, url, id)
	return err
}

// ReorderImages updates the position field for a batch of image IDs in an album.
func (s *Store) ReorderImages(album string, ids []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE art_pieces SET position = ? WHERE id = ? AND album = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, id := range ids {
		if _, err := stmt.Exec(i, id, album); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ReorderAlbums updates the position field in albums_meta for a batch of album names.
func (s *Store) ReorderAlbums(ids []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, album := range ids {
		_, err := tx.Exec(
			`INSERT INTO albums_meta (album, position) VALUES (?, ?)
			 ON CONFLICT(album) DO UPDATE SET position = excluded.position`,
			album, i)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RecordView increments the view counter for an asset/domain pair in memory.
// It is completely non-blocking and thread-safe. Skips empty domains.
func (s *Store) RecordView(assetID, domain string) error {
	if domain == "" {
		return nil
	}
	s.viewMu.Lock()
	s.pendingViews[viewKey{assetID: assetID, domain: domain}]++
	s.viewMu.Unlock()
	return nil
}

// viewFlusher periodically writes all pending view counts to the database in a single transaction.
func (s *Store) viewFlusher() {
	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		s.viewMu.Lock()
		batch := s.pendingViews
		if len(batch) > 0 {
			s.pendingViews = make(map[viewKey]int)
		}
		s.viewMu.Unlock()

		if len(batch) == 0 {
			continue
		}

		s.flushViews(batch)
	}
}

func (s *Store) flushViews(batch map[viewKey]int) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("flush views begin tx: %v", err)
		return
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	stmtUpdate, err := tx.Prepare(`UPDATE asset_views SET count = count + ?, last_seen = ? WHERE asset_id = ? AND domain = ?`)
	if err != nil {
		return
	}
	defer stmtUpdate.Close()

	stmtInsert, err := tx.Prepare(`INSERT INTO asset_views (asset_id, domain, count, first_seen, last_seen) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	defer stmtInsert.Close()

	for k, count := range batch {
		res, err := stmtUpdate.Exec(count, now, k.assetID, k.domain)
		if err == nil {
			affected, _ := res.RowsAffected()
			if affected == 0 {
				_, _ = stmtInsert.Exec(k.assetID, k.domain, count, now, now)
			}
		}
	}
	_ = tx.Commit()
}

// GetAssetStats returns total views and a per-referrer breakdown for an asset.
func (s *Store) GetAssetStats(assetID string) (*AssetStats, error) {
	var total int
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(count), 0) FROM asset_views WHERE asset_id = ?`, assetID).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		`SELECT domain, count, first_seen, last_seen FROM asset_views
		 WHERE asset_id = ? ORDER BY count DESC`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := []ReferrerView{}
	for rows.Next() {
		var r ReferrerView
		if err := rows.Scan(&r.Domain, &r.Count, &r.FirstSeen, &r.LastSeen); err != nil {
			return nil, err
		}
		refs = append(refs, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AssetStats{Total: total, Referrers: refs}, nil
}

// migrate creates tables and indexes if they do not exist.
func migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS art_pieces (
	id TEXT PRIMARY KEY,
	imgchest_url TEXT NOT NULL,
	album TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_art_pieces_album ON art_pieces(album);

CREATE TABLE IF NOT EXISTS asset_views (
	asset_id TEXT NOT NULL,
	domain TEXT NOT NULL,
	count INTEGER NOT NULL DEFAULT 0,
	first_seen INTEGER NOT NULL,
	last_seen INTEGER NOT NULL,
	PRIMARY KEY (asset_id, domain)
);

CREATE INDEX IF NOT EXISTS idx_views_asset ON asset_views(asset_id);

CREATE TABLE IF NOT EXISTS domains (
	domain TEXT PRIMARY KEY,
	banned INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT
);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Add position column to art_pieces if missing
	_, _ = db.Exec(`ALTER TABLE art_pieces ADD COLUMN position INTEGER DEFAULT 0`)

	// Add content_type column to art_pieces if missing
	_, _ = db.Exec(`ALTER TABLE art_pieces ADD COLUMN content_type TEXT`)

	// Album ordering table
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS albums_meta (
		album TEXT PRIMARY KEY,
		position INTEGER NOT NULL DEFAULT 0
	)`)

	// Clean up legacy blank-domain view rows
	_, _ = db.Exec(`DELETE FROM asset_views WHERE domain = '' OR domain IS NULL`)

	log.Println("database migrated")
	return nil
}
