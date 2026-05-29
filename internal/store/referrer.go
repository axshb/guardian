package store

import (
	"database/sql"
	"strings"
	"sync"
)

// ReferrerManager handles domain banning, unbanning, and redirect images.
// It caches domain statuses and settings to eliminate database queries on the hot path.
type ReferrerManager struct {
	db *sql.DB

	mu          sync.RWMutex
	bannedCache map[string]bool
	redirectUrl string
}

// NewReferrerManager creates a manager backed by the same SQLite database.
func NewReferrerManager(db *sql.DB) *ReferrerManager {
	rm := &ReferrerManager{
		db:          db,
		bannedCache: make(map[string]bool),
	}
	rm.loadCache()
	return rm
}

func (r *ReferrerManager) loadCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load bans
	rows, err := r.db.Query(`SELECT domain, banned FROM domains`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var domain string
			var banned int
			if err := rows.Scan(&domain, &banned); err == nil {
				if banned == 1 {
					r.bannedCache[normalizeDomain(domain)] = true
				}
			}
		}
	}

	// Load redirect image
	var url string
	err = r.db.QueryRow(`SELECT value FROM settings WHERE key = 'redirect_image'`).Scan(&url)
	if err == nil {
		r.redirectUrl = url
	}
}

// IsBanned checks whether a domain (or its www-prefixed variant) is banned.
// This uses an in-memory cache and is safe for high-concurrency access.
func (r *ReferrerManager) IsBanned(domain string) bool {
	norm := normalizeDomain(domain)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.bannedCache[norm]
}

// ListReferrers returns every known domain and its ban status.
func (r *ReferrerManager) ListReferrers() ([]ReferrerInfo, error) {
	rows, err := r.db.Query(`SELECT domain, banned FROM domains ORDER BY domain ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ReferrerInfo{}
	for rows.Next() {
		var d ReferrerInfo
		if err := rows.Scan(&d.Domain, &d.Banned); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListReferrersDetailed returns domains enriched with view stats and album access.
func (r *ReferrerManager) ListReferrersDetailed() ([]ReferrerDetail, error) {
	// 1. Load all domains
	rows, err := r.db.Query(`SELECT domain, banned FROM domains`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]*mutableReferrer)
	for rows.Next() {
		var domain string
		var banned int
		if err := rows.Scan(&domain, &banned); err != nil {
			return nil, err
		}
		norm := normalizeDomain(domain)
		ref := m[norm]
		if ref == nil {
			ref = &mutableReferrer{domain: norm, views: make(map[string]*mutableView)}
			m[norm] = ref
		}
		ref.banned = ref.banned || banned == 1
	}

	// 2. Load all asset views and merge by normalized domain
	viewRows, err := r.db.Query(`SELECT domain, asset_id, count, last_seen FROM asset_views`)
	if err != nil {
		return nil, err
	}
	defer viewRows.Close()

	for viewRows.Next() {
		var domain, assetID string
		var count int
		var lastSeen int64
		if err := viewRows.Scan(&domain, &assetID, &count, &lastSeen); err != nil {
			return nil, err
		}
		norm := normalizeDomain(domain)
		ref := m[norm]
		if ref == nil {
			ref = &mutableReferrer{domain: norm, views: make(map[string]*mutableView)}
			m[norm] = ref
		}
		ref.totalViews += count
		if lastSeen > ref.lastSeen {
			ref.lastSeen = lastSeen
		}
		mv := ref.views[assetID]
		if mv == nil {
			mv = &mutableView{}
			ref.views[assetID] = mv
		}
		mv.count += count
		if lastSeen > mv.lastSeen {
			mv.lastSeen = lastSeen
		}
	}
	if err := viewRows.Err(); err != nil {
		return nil, err
	}

	// 3. Resolve asset -> album mapping
	assetRows, err := r.db.Query(`SELECT id, album FROM art_pieces`)
	if err != nil {
		return nil, err
	}
	defer assetRows.Close()

	assetAlbums := make(map[string]string)
	for assetRows.Next() {
		var id, album string
		if err := assetRows.Scan(&id, &album); err != nil {
			return nil, err
		}
		assetAlbums[id] = album
	}
	if err := assetRows.Err(); err != nil {
		return nil, err
	}

	// 4. Build response sorted by lastSeen desc
	result := []ReferrerDetail{}
	for _, ref := range m {
		albumMap := make(map[string][]ImageAccess)
		for assetID, mv := range ref.views {
			album := assetAlbums[assetID]
			if album == "" {
				album = "Unknown"
			}
			albumMap[album] = append(albumMap[album], ImageAccess{
				ID:       assetID,
				Views:    mv.count,
				LastSeen: mv.lastSeen,
			})
		}

		albums := []AlbumAccess{}
		for album, images := range albumMap {
			albums = append(albums, AlbumAccess{
				Album:  album,
				Images: images,
			})
		}
		result = append(result, ReferrerDetail{
			Domain:     ref.domain,
			Banned:     ref.banned,
			LastSeen:   ref.lastSeen,
			TotalViews: ref.totalViews,
			Albums:     albums,
		})
	}

	return result, nil
}

// SetBanned inserts or updates a domain's ban status.
func (r *ReferrerManager) SetBanned(domain string, banned bool) error {
	norm := normalizeDomain(domain)

	// Check if legacy www-prefixed entry exists
	var legacyDomain string
	_ = r.db.QueryRow(`SELECT domain FROM domains WHERE domain = ?`, "www."+norm).Scan(&legacyDomain)

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if legacyDomain != "" {
		_, err = tx.Exec(`DELETE FROM domains WHERE domain = ?`, legacyDomain)
		if err != nil {
			return err
		}
	}

	bannedInt := 0
	if banned {
		bannedInt = 1
	}

	_, err = tx.Exec(
		`INSERT INTO domains (domain, banned) VALUES (?, ?)
		 ON CONFLICT(domain) DO UPDATE SET banned = excluded.banned`,
		norm, bannedInt)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err == nil {
		r.mu.Lock()
		if banned {
			r.bannedCache[norm] = true
		} else {
			delete(r.bannedCache, norm)
		}
		r.mu.Unlock()
	}
	return err
}

// ForgetDomain removes a domain and its legacy www variant.
func (r *ReferrerManager) ForgetDomain(domain string) error {
	norm := normalizeDomain(domain)
	_, err := r.db.Exec(`DELETE FROM domains WHERE domain = ? OR domain = ?`, norm, "www."+norm)
	if err == nil {
		r.mu.Lock()
		delete(r.bannedCache, norm)
		r.mu.Unlock()
	}
	return err
}

// GetRedirectImageUrl returns the configured redirect image URL for banned domains.
func (r *ReferrerManager) GetRedirectImageUrl() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.redirectUrl
}

// SetRedirectImage stores the redirect image URL.
func (r *ReferrerManager) SetRedirectImage(url string) error {
	_, err := r.db.Exec(
		`INSERT INTO settings (key, value) VALUES ('redirect_image', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`, url)
	if err == nil {
		r.mu.Lock()
		r.redirectUrl = url
		r.mu.Unlock()
	}
	return err
}

func normalizeDomain(domain string) string {
	if domain == "" {
		return ""
	}
	return strings.TrimPrefix(domain, "www.")
}

// ---- Response types ----

// ReferrerInfo is a lightweight domain listing.
type ReferrerInfo struct {
	Domain string `json:"domain"`
	Banned bool   `json:"banned"`
}

// ReferrerDetail is an enriched domain view with stats.
type ReferrerDetail struct {
	Domain     string        `json:"domain"`
	Banned     bool          `json:"banned"`
	LastSeen   int64         `json:"lastSeen"`
	TotalViews int           `json:"totalViews"`
	Albums     []AlbumAccess `json:"albums"`
}

// AlbumAccess groups image accesses by album.
type AlbumAccess struct {
	Album  string        `json:"album"`
	Images []ImageAccess `json:"images"`
}

// ImageAccess tracks views per image.
type ImageAccess struct {
	ID       string `json:"id"`
	Views    int    `json:"views"`
	LastSeen int64  `json:"lastSeen"`
}

type mutableReferrer struct {
	domain     string
	banned     bool
	lastSeen   int64
	totalViews int
	views      map[string]*mutableView
}

type mutableView struct {
	count    int
	lastSeen int64
}

// extractHost pulls the hostname from a URL string.
func ExtractHost(url string) string {
	if url == "" {
		return ""
	}
	url = strings.TrimSpace(url)
	// Simple extraction; chi's URL parsing or net/url could also be used
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "//")
	url = strings.TrimPrefix(url, "www.")
	idx := strings.IndexAny(url, "/:?#")
	if idx >= 0 {
		url = url[:idx]
	}
	return url
}
