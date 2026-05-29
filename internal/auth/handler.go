package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/axshb/guardian/internal/config"
	"github.com/go-chi/chi/v5"
)

// Handler holds dependencies for auth-related routes.
type Handler struct {
	dataPath string
}

// NewHandler creates an auth handler.
func NewHandler(dataPath string) *Handler {
	return &Handler{dataPath: dataPath}
}

// RegisterRoutes mounts auth endpoints.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/api/config", h.getConfig)
	r.Patch("/api/config", h.patchConfig)
	r.Post("/api/setup", h.setup)
	r.Post("/api/auth/login", h.login)
	r.Post("/api/auth/logout", h.logout)
	r.Get("/api/auth/me", h.me)
}

func (h *Handler) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadAppConfig(h.dataPath)
	if err != nil {
		log.Printf("load config: %v", err)
		jsonError(w, http.StatusInternalServerError, "Failed to load config")
		return
	}
	// Don't expose the admin hash
	resp := map[string]any{
		"setupComplete": cfg.SetupComplete,
		"theme":         cfg.Theme,
	}
	writeJSON(w, resp)
}

func (h *Handler) patchConfig(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	if !session["isAdmin"] {
		jsonError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "Bad request")
		return
	}

	cfg, err := config.LoadAppConfig(h.dataPath)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to load config")
		return
	}

	if v, ok := body["theme"].(string); ok {
		cfg.Theme = v
	}

	if err := config.SaveAppConfig(h.dataPath, cfg); err != nil {
		log.Printf("save config: %v", err)
		jsonError(w, http.StatusInternalServerError, "Failed to save config")
		return
	}

	writeJSON(w, cfg)
}

func (h *Handler) setup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
		Theme    string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "Bad request")
		return
	}

	if len(body.Password) < 8 {
		jsonError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	hash, err := HashPassword(body.Password)
	if err != nil {
		log.Printf("hash password: %v", err)
		jsonError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	theme := body.Theme
	if theme == "" {
		theme = "gruvbox"
	}

	cfg := &config.AppConfig{
		SetupComplete: true,
		Theme:         theme,
		AdminHash:     hash,
	}

	if err := config.SaveAppConfig(h.dataPath, cfg); err != nil {
		log.Printf("save config: %v", err)
		jsonError(w, http.StatusInternalServerError, "Failed to save config")
		return
	}

	token, err := CreateSession(map[string]bool{"isAdmin": true})
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "guardian_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 3600,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	allowed, _ := CheckRateLimit(clientIP)
	if !allowed {
		jsonError(w, http.StatusTooManyRequests, "Too many attempts")
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "Bad request")
		return
	}

	cfg, err := config.LoadAppConfig(h.dataPath)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to load config")
		return
	}

	if cfg.AdminHash == "" {
		jsonError(w, http.StatusBadRequest, "Not configured")
		return
	}

	if !VerifyPassword(body.Password, cfg.AdminHash) {
		jsonError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	token, err := CreateSession(map[string]bool{"isAdmin": true})
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "guardian_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 3600,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "guardian_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	writeJSON(w, session)
}

// getSession extracts session data from the cookie.
func getSession(r *http.Request) map[string]bool {
	cookie, err := r.Cookie("guardian_session")
	if err != nil {
		return map[string]bool{"isAdmin": false}
	}
	return DecodeSession(cookie.Value)
}

// jsonError sends a structured error response.
func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// writeJSON encodes v as JSON.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode json: %v", err)
	}
}
