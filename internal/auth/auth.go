package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	secret     []byte
	secretOnce sync.Once
	secretPath string
)

// SetSecretPath sets the file path where the HMAC secret is persisted.
func SetSecretPath(p string) {
	secretPath = p
}

func getSecret() []byte {
	secretOnce.Do(func() {
		if secretPath != "" {
			if data, err := os.ReadFile(secretPath); err == nil && len(data) >= 32 {
				secret = data
				return
			}
		}
		s := make([]byte, 32)
		if _, err := rand.Read(s); err != nil {
			s = []byte("guardian-session-secret-change-me")
		}
		secret = s
		if secretPath != "" {
			_ = os.WriteFile(secretPath, s, 0600)
		}
	})
	return secret
}

// SessionData holds the claims stored in a session token.
type SessionData struct {
	IsAdmin bool  `json:"isAdmin"`
	Expiry  int64 `json:"expiry"`
}

// CreateSession generates a signed session token.
func CreateSession(data map[string]bool) (string, error) {
	payload := SessionData{
		IsAdmin: data["isAdmin"],
		Expiry:  time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, getSecret())
	mac.Write([]byte(encoded))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encoded + "." + sig, nil
}

// DecodeSession verifies and decodes a session token.
func DecodeSession(token string) map[string]bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return map[string]bool{"isAdmin": false}
	}

	// Verify signature
	mac := hmac.New(sha256.New, getSecret())
	mac.Write([]byte(parts[0]))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return map[string]bool{"isAdmin": false}
	}

	b, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return map[string]bool{"isAdmin": false}
	}

	var data SessionData
	if err := json.Unmarshal(b, &data); err != nil {
		return map[string]bool{"isAdmin": false}
	}

	if time.Now().Unix() > data.Expiry {
		return map[string]bool{"isAdmin": false}
	}

	return map[string]bool{"isAdmin": data.IsAdmin}
}

// HashPassword creates a bcrypt hash of the given password.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

// VerifyPassword checks a password against a bcrypt hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// -- Rate limiter --

type rateEntry struct {
	count   int
	resetAt int64
}

var (
	rateMu     sync.Mutex
	rateLimits = make(map[string]*rateEntry)
)

const maxAttempts = 5
const windowSecs = 300

// CheckRateLimit checks if a key (IP) has exceeded the rate limit.
func CheckRateLimit(key string) (allowed bool, remaining int) {
	now := time.Now().Unix()
	rateMu.Lock()
	defer rateMu.Unlock()

	entry, exists := rateLimits[key]
	if !exists || now > entry.resetAt {
		rateLimits[key] = &rateEntry{count: 1, resetAt: now + windowSecs}
		return true, maxAttempts - 1
	}

	if entry.count >= maxAttempts {
		return false, 0
	}

	entry.count++
	return true, maxAttempts - entry.count
}
