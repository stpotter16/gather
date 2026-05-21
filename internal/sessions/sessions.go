package sessions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const cookieName = "gather_session"

type Manager struct {
	secret []byte
	maxAge time.Duration
}

func New(secret string) (*Manager, error) {
	if secret == "" {
		return nil, errors.New("GATHER_HMAC_SECRET not set")
	}
	return &Manager{
		secret: []byte(secret),
		maxAge: 30 * 24 * time.Hour,
	}, nil
}

func (m *Manager) Set(w http.ResponseWriter, userID int) {
	expires := time.Now().Add(m.maxAge)
	payload := fmt.Sprintf("%d|%d", userID, expires.Unix())
	sig := m.sign(payload)
	value := payload + "." + base64.URLEncoding.EncodeToString(sig)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (m *Manager) Get(r *http.Request) (int, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return 0, err
	}

	dotIdx := strings.LastIndex(cookie.Value, ".")
	if dotIdx < 0 {
		return 0, errors.New("malformed session cookie")
	}

	payload := cookie.Value[:dotIdx]
	sig, err := base64.URLEncoding.DecodeString(cookie.Value[dotIdx+1:])
	if err != nil {
		return 0, errors.New("malformed session cookie")
	}

	if !hmac.Equal(sig, m.sign(payload)) {
		return 0, errors.New("invalid session signature")
	}

	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		return 0, errors.New("malformed session payload")
	}

	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, errors.New("malformed user ID in session")
	}

	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, errors.New("malformed expiry in session")
	}

	if time.Now().Unix() > expires {
		return 0, errors.New("session expired")
	}

	return userID, nil
}

func (m *Manager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	})
}

func (m *Manager) sign(payload string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return mac.Sum(nil)
}
