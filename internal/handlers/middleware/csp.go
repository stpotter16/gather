package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
)

type contextKey struct{ name string }

var nonceKey = contextKey{"csp-nonce"}

func NonceFromContext(ctx context.Context) (string, bool) {
	nonce, ok := ctx.Value(nonceKey).(string)
	return nonce, ok
}

func CspMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		nonce := base64.StdEncoding.EncodeToString(b)

		ctx := context.WithValue(r.Context(), nonceKey, nonce)

		csp := fmt.Sprintf(
			"default-src 'self'; script-src 'self' 'nonce-%s'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'",
			nonce,
		)
		w.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
