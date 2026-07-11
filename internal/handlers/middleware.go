package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/tuantu/oj-web/internal/auth"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

type contextKey string

const userContextKey = contextKey("user")

// LoadAndSave is a middleware that loads and saves session data
func LoadAndSave(next http.Handler) http.Handler {
	return auth.SessionManager.LoadAndSave(next)
}

// Authenticate is a middleware that checks if the user is logged in
// and sets the user object in the request context.
func (env *Env) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.SessionManager.GetInt64(r.Context(), "userID")
		log.Printf("Authenticate middleware: Session userID = %d", userID)
		if userID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		user, err := env.Queries.GetUserByID(r.Context(), userID)
		if err != nil {
			log.Printf("Authenticate middleware: Error fetching user %d: %v", userID, err)
			auth.SessionManager.Remove(r.Context(), "userID")
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("Authenticate middleware: User found: %s", user.Username)
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth is a middleware that ensures the user is authenticated.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userContextKey)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext gets the authenticated user from the context.
func GetUserFromContext(ctx context.Context) *sqlcdb.User {
	user, ok := ctx.Value(userContextKey).(sqlcdb.User)
	if ok {
		return &user
	}
	return nil
}

// RequireAdmin is a middleware that ensures the user is an admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
