package middleware

import (
	"context"
	"net/http"

	"github.com/stpotter16/gather/internal/sessions"
	"github.com/stpotter16/gather/internal/store"
)

type userContextKey struct{}

func UserFromContext(ctx context.Context) (store.User, bool) {
	u, ok := ctx.Value(userContextKey{}).(store.User)
	return u, ok
}

type userByID interface {
	GetUserByID(ctx context.Context, id int) (store.User, error)
}

func RequireAuth(sm *sessions.Manager, us userByID, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := sm.Get(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		user, err := us.GetUserByID(r.Context(), userID)
		if err != nil {
			sm.Clear(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
