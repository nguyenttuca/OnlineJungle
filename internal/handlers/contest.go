package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) getContestProblemOrNotFound(w http.ResponseWriter, r *http.Request, contestID int64, slug string) (*sqlcdb.Problem, bool) {
	problem, err := env.Queries.GetContestProblem(r.Context(), sqlcdb.GetContestProblemParams{
		ContestID: contestID,
		Slug:      slug,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return nil, false
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return nil, false
	}
	return &problem, true
}

func (env *Env) ContestProblemDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	contestID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	slug := chi.URLParam(r, "slug")

	contest, err := env.Queries.GetContestByID(ctx, contestID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	problem, ok := env.getContestProblemOrNotFound(w, r, contestID, slug)
	if !ok {
		return
	}

	now := time.Now()
	canSubmit := now.After(contest.StartAt.Time) && now.Before(contest.EndAt.Time)

	render(w, r, "problem_detail.html", map[string]interface{}{
		"Problem":   problem,
		"ContestID": &contestID,
		"CanSubmit": canSubmit,
	})
}

func (env *Env) ContestSubmitGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	contestID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	slug := chi.URLParam(r, "slug")

	contest, err := env.Queries.GetContestByID(ctx, contestID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	problem, ok := env.getContestProblemOrNotFound(w, r, contestID, slug)
	if !ok {
		return
	}

	now := time.Now()
	canSubmit := now.After(contest.StartAt.Time) && now.Before(contest.EndAt.Time)

	render(w, r, "submit.html", map[string]interface{}{
		"Problem":   problem,
		"ContestID": &contestID,
		"CanSubmit": canSubmit,
	})
}
