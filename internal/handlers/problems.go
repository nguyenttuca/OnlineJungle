package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (env *Env) ProblemListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	problems, err := env.Queries.ListProblems(ctx)
	if err != nil {
		log.Printf("Error listing problems: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user := GetUserFromContext(r.Context())

	render(w, r, "problems.html", map[string]interface{}{
		"Problems": problems,
		"User":     user,
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
