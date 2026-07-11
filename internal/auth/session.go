package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/tuantu/oj-web/internal/database"
)

var SessionManager *scs.SessionManager

func InitSession(db *database.DB) {
	SessionManager = scs.New()
	SessionManager.Store = postgresstore.New(db.SqlDB)
	SessionManager.Lifetime = 24 * time.Hour
	SessionManager.Cookie.Name = "oj_session"
	SessionManager.Cookie.HttpOnly = true
	SessionManager.Cookie.Secure = false // Set to true if using HTTPS
	SessionManager.Cookie.SameSite = http.SameSiteLaxMode
	SessionManager.Cookie.Path = "/"
	SessionManager.ErrorFunc = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Session error: %v", err)
	}
}
