package handlers

import (
	"net/http"

	"github.com/tuantu/oj-web/internal/auth"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
	"golang.org/x/crypto/bcrypt"
)

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

func (env *Env) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "login.html", nil)
}

func (env *Env) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := env.Queries.GetUserByUsername(r.Context(), username)
	if err != nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		render(w, r, "login.html", map[string]string{"Error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		render(w, r, "login.html", map[string]string{"Error": "Invalid credentials"})
		return
	}

	auth.SessionManager.Put(r.Context(), "userID", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (env *Env) AdminCreateUserGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "register.html", nil)
}

func (env *Env) AdminCreateUserPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	displayName := r.FormValue("display_name")

	if username == "" || password == "" {
		render(w, r, "register.html", map[string]string{"Error": "Username and password are required"})
		return
	}

	if len(username) < 3 || len(username) > 30 {
		render(w, r, "register.html", map[string]string{"Error": "Username must be between 3 and 30 characters"})
		return
	}

	if len(password) < 8 {
		render(w, r, "register.html", map[string]string{"Error": "Password must be at least 8 characters"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = env.Queries.CreateUser(r.Context(), sqlcdb.CreateUserParams{
		Username:     username,
		PasswordHash: string(hashedPassword),
		DisplayName:  displayName,
		Role:         "user",
	})

	if err != nil {
		render(w, r, "register.html", map[string]string{"Error": "Username already taken"})
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (env *Env) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.SessionManager.Destroy(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
