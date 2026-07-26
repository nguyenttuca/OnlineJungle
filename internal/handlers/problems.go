package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) ProblemListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")

	user := GetUserFromContext(ctx)
	var userID *int64
	if user != nil {
		userID = &user.ID
	}

	rawProblems, err := env.Queries.SearchProblems(ctx, sqlcdb.SearchProblemsParams{
		SearchQuery:   q,
		Category:      category,
		UserID:        userID,
		IncludeHidden: false,
	})
	if err != nil {
		log.Printf("Error searching problems: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// In-memory filter by status
	var filteredProblems []sqlcdb.SearchProblemsRow
	for _, p := range rawProblems {
		include := true
		// p.UserStatus will be "" if no submission was found, because we used COALESCE in SQL
		if status == "solved" {
			if p.UserStatus != "AC" {
				include = false
			}
		} else if status == "attempted" {
			if p.UserStatus == "AC" || p.UserStatus == "" {
				include = false
			}
		} else if status == "unsolved" {
			if p.UserStatus != "" {
				include = false
			}
		}
		if include {
			filteredProblems = append(filteredProblems, p)
		}
	}

	categories, _ := env.Queries.GetProblemCategories(ctx)

	render(w, r, "problems.html", map[string]interface{}{
		"Problems":   filteredProblems,
		"Categories": categories,
		"Q":          q,
		"Category":   category,
		"Status":     status,
		"User":       user,
	})
}

func (env *Env) ProblemDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	problem, err := env.Queries.GetProblemBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	render(w, r, "problem_detail.html", map[string]interface{}{
		"Problem": problem,
	})
}

func (env *Env) ProblemEditorialHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	problem, err := env.Queries.GetProblemBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if problem.EditorialContent == "" {
		http.NotFound(w, r)
		return
	}

	render(w, r, "editorial.html", map[string]interface{}{
		"Problem": problem,
	})
}
