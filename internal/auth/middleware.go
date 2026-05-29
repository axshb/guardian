package auth

import (
	"net/http"
)

// RequireAdmin is an HTTP middleware that ensures the request has a valid admin session.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if !session["isAdmin"] {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
