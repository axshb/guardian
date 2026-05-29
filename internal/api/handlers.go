package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/axshb/guardian/internal/store"
	"github.com/go-chi/chi/v5"
)

// Handler holds service dependencies.
type Handler struct {
	store     *store.Store
	referrers *store.ReferrerManager
	client    *http.Client
}

// NewHandler creates a Handler wired to the store and an HTTP client.
func NewHandler(s *store.Store, r *store.ReferrerManager) *Handler {
	return &Handler{
		store:     s,
		referrers: r,
		client:    &http.Client{},
	}
}

// RegisterRoutes mounts all assets endpoints on the router.
// adminMdw is the auth middleware for admin-only endpoints.
func (h *Handler) RegisterRoutes(r chi.Router, adminMdw func(http.Handler) http.Handler) {
	r.Get("/health", h.health)
	r.Get("/assets/{id}.{ext}", h.resolve)
	r.Get("/art/{id}", h.resolve)
	r.Get("/assets/{id}", h.resolve)

	// Admin-only API routes
	r.Group(func(r chi.Router) {
		r.Use(adminMdw)
		r.Get("/api/assets/{id}/stats", h.assetStats)
		r.Get("/api/albums", h.getAlbums)
		r.Post("/api/albums", h.importImages)
		r.Delete("/api/albums/{album}", h.deleteAlbum)
		r.Get("/api/albums/{album}/images", h.getAlbumImages)
		r.Delete("/api/albums/{album}/images/{id}", h.deleteImage)
		r.Put("/api/albums/{album}/images/{id}", h.updateImageSource)
		r.Get("/api/referrers", h.getReferrers)
		r.Post("/api/referrers/set-redirect", h.setRedirectImage)
		r.Post("/api/referrers/ban", h.banDomain)
		r.Post("/api/referrers/unban", h.unbanDomain)
		r.Post("/api/referrers/forget", h.forgetDomain)
		r.Put("/api/albums/{album}/reorder", h.reorderImages)
		r.Put("/api/albums/reorder", h.reorderAlbums)
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// resolve proxies an image by ID, tracking referrers and enforcing bans.
func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if ext := path.Ext(id); ext != "" {
		id = strings.TrimSuffix(id, ext)
	}

	img, err := h.store.FindImage(id)
	if err != nil {
		log.Printf("find image %s: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if img == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	domain := store.ExtractHost(r.Header.Get("Referer"))
	if domain != "" && h.referrers.IsBanned(domain) {
		redirectURL := h.referrers.GetRedirectImageUrl()
		if redirectURL != "" {
			resp, err := h.client.Get(redirectURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				ct := resp.Header.Get("Content-Type")
				if ct == "" {
					ct = "image/jpeg"
				}
				setNoCacheHeaders(w)
				w.Header().Set("Content-Type", ct)
				_, _ = io.Copy(w, resp.Body)
				return
			}
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	resp, err := h.client.Get(img.ImgchestURL)
	if err != nil {
		log.Printf("proxy image %s: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}

	// Store detected content type if not already known
	if img.ContentType == "" && ct != "" {
		go func() {
			if err := h.store.SetImageContentType(id, ct); err != nil {
				log.Printf("store content_type %s: %v", id, err)
			}
		}()
	}

	setNoCacheHeaders(w)
	w.Header().Set("Content-Type", ct)
	_, _ = io.Copy(w, resp.Body)

	// Record view in background; failures must not break image serving
	go func() {
		if err := h.store.RecordView(id, domain); err != nil {
			log.Printf("record view %s from %s: %v", id, domain, err)
		}
	}()
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0, s-maxage=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
}

// assetStats returns view-count breakdown for a single asset.
func (h *Handler) assetStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stats, err := h.store.GetAssetStats(id)
	if err != nil {
		log.Printf("asset stats %s: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

// getAlbums lists all albums.
func (h *Handler) getAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := h.store.ListAlbums()
	if err != nil {
		log.Printf("list albums: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(albums)
}

// importImages creates a batch of assets in an album.
func (h *Handler) importImages(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Album string   `json:"album"`
		URLs  []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Album == "" || len(body.URLs) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.store.ImportImages(body.Album, body.URLs); err != nil {
		log.Printf("import images: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// deleteAlbum removes an entire album.
func (h *Handler) deleteAlbum(w http.ResponseWriter, r *http.Request) {
	album := chi.URLParam(r, "album")
	if err := h.store.DeleteAlbum(album); err != nil {
		log.Printf("delete album %s: %v", album, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getAlbumImages lists images in an album.
func (h *Handler) getAlbumImages(w http.ResponseWriter, r *http.Request) {
	album := chi.URLParam(r, "album")
	images, err := h.store.ListImages(album)
	if err != nil {
		log.Printf("list images %s: %v", album, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(images)
}

// deleteImage removes a single asset.
func (h *Handler) deleteImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.store.DeleteImage(id); err != nil {
		log.Printf("delete image %s: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updateImageSource changes the imgchest_url for an asset.
func (h *Handler) updateImageSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.URL == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.store.UpdateImageSource(id, body.URL); err != nil {
		log.Printf("update source %s: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getReferrers returns detailed domain list with view stats and album access.
func (h *Handler) getReferrers(w http.ResponseWriter, r *http.Request) {
	domains, err := h.referrers.ListReferrersDetailed()
	if err != nil {
		log.Printf("list referrers: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	redirect := h.referrers.GetRedirectImageUrl()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"domains":       domains,
		"redirectImage": redirect,
	})
}

// setRedirectImage stores the redirect image URL for banned domains.
func (h *Handler) setRedirectImage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.referrers.SetRedirectImage(body.URL); err != nil {
		log.Printf("set redirect image: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// banDomain marks a domain as banned.
func (h *Handler) banDomain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Domain == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.referrers.SetBanned(body.Domain, true); err != nil {
		log.Printf("ban domain %s: %v", body.Domain, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// unbanDomain removes the ban on a domain.
func (h *Handler) unbanDomain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Domain == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.referrers.SetBanned(body.Domain, false); err != nil {
		log.Printf("unban domain %s: %v", body.Domain, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// forgetDomain removes a domain from the database entirely.
func (h *Handler) forgetDomain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Domain == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.referrers.ForgetDomain(body.Domain); err != nil {
		log.Printf("forget domain %s: %v", body.Domain, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// reorderImages updates the position of images in an album.
func (h *Handler) reorderImages(w http.ResponseWriter, r *http.Request) {
	album := chi.URLParam(r, "album")
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(body.IDs) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.store.ReorderImages(album, body.IDs); err != nil {
		log.Printf("reorder images %s: %v", album, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// reorderAlbums updates the position of albums globally.
func (h *Handler) reorderAlbums(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(body.IDs) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.store.ReorderAlbums(body.IDs); err != nil {
		log.Printf("reorder albums: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
