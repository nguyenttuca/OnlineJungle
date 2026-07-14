package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
	"github.com/tuantu/oj-web/internal/dispatcher"
)

func (env *Env) SubmitGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	problem, err := env.Queries.GetProblemBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	render(w, r, "submit.html", map[string]interface{}{
		"Problem": problem,
	})
}

func (env *Env) SubmitPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	slug := chi.URLParam(r, "slug")
	problem, err := env.Queries.GetProblemBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	language := r.FormValue("language")
	sourceCode := r.FormValue("source_code")

	if language == "" || sourceCode == "" {
		http.Error(w, "Language and source code are required", http.StatusBadRequest)
		return
	}

	// Create submission in DB (status: queued)
	submission, err := env.Queries.CreateSubmission(ctx, sqlcdb.CreateSubmissionParams{
		UserID:      user.ID,
		ProblemID:   problem.ID,
		Language:    language,
		SourceCode:  sourceCode,
		CodeSize:    int32(len(sourceCode)),
		RunAllTests: true,
	})

	if err != nil {
		log.Printf("Failed to create submission: %v", err)
		http.Error(w, "Failed to submit", http.StatusInternalServerError)
		return
	}

	select {
	case dispatcher.WakeupDispatcher <- struct{}{}:
	default:
	}

	// Redirect to submission detail page (or back to problem with success message)
	http.Redirect(w, r, "/submissions/"+strconv.FormatInt(submission.ID, 10), http.StatusSeeOther)
}
