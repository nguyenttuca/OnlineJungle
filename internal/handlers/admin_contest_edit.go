package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) AdminEditContestGetHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	contest, err := env.Queries.GetContestByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	contestProblems, _ := env.Queries.ListContestProblems(r.Context(), id)
	allProblems, _ := env.Queries.ListProblems(r.Context())

	// Need a custom struct to pass the time string formatted correctly for HTML datetime-local
	type ContestView struct {
		Contest         sqlcdb.Contest
		StartAtStr      string
		EndAtStr        string
		ContestProblems []sqlcdb.ListContestProblemsRow
		AllProblems     []sqlcdb.Problem
		Error           string
	}

	cv := ContestView{
		Contest:         contest,
		StartAtStr:      contest.StartAt.Time.Format("2006-01-02T15:04"),
		EndAtStr:        contest.EndAt.Time.Format("2006-01-02T15:04"),
		ContestProblems: contestProblems,
		AllProblems:     allProblems,
	}

	render(w, r, "admin_edit_contest.html", cv)
}

func (env *Env) AdminEditContestPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	title := r.FormValue("name")
	startAtStr := r.FormValue("start_time")
	endAtStr := r.FormValue("end_time")

	if title == "" || startAtStr == "" || endAtStr == "" {
		http.Redirect(w, r, "/admin/contests/"+idStr+"/edit?error=fields_required", http.StatusSeeOther)
		return
	}

	startAt, errStart := time.Parse("2006-01-02T15:04", startAtStr)
	endAt, errEnd := time.Parse("2006-01-02T15:04", endAtStr)

	if errStart != nil || errEnd != nil || endAt.Before(startAt) {
		http.Redirect(w, r, "/admin/contests/"+idStr+"/edit?error=invalid_time", http.StatusSeeOther)
		return
	}

	err = env.Queries.UpdateContest(r.Context(), sqlcdb.UpdateContestParams{
		ID:      id,
		Title:   title,
		StartAt: pgtype.Timestamptz{Time: startAt, Valid: true},
		EndAt:   pgtype.Timestamptz{Time: endAt, Valid: true},
	})

	if err != nil {
		http.Redirect(w, r, "/admin/contests/"+idStr+"/edit?error=update_failed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/contests/"+idStr+"/edit?success=1", http.StatusSeeOther)
}

func (env *Env) AdminAddContestProblemPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	problemIDStr := r.FormValue("problem_id")
	label := r.FormValue("label")
	pointsStr := r.FormValue("points")

	problemID, _ := strconv.ParseInt(problemIDStr, 10, 64)
	points, _ := strconv.Atoi(pointsStr)

	if problemID > 0 && label != "" {
		_ = env.Queries.AddContestProblem(r.Context(), sqlcdb.AddContestProblemParams{
			ContestID: id,
			ProblemID: problemID,
			Label:     label,
			Points:    int32(points),
		})
	}

	http.Redirect(w, r, "/admin/contests/"+idStr+"/edit", http.StatusSeeOther)
}

func (env *Env) AdminRemoveContestProblemPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	problemIDStr := r.FormValue("problem_id")
	problemID, err := strconv.ParseInt(problemIDStr, 10, 64)
	if err == nil {
		_ = env.Queries.RemoveContestProblem(r.Context(), sqlcdb.RemoveContestProblemParams{
			ContestID: id,
			ProblemID: problemID,
		})
	}

	http.Redirect(w, r, "/admin/contests/"+idStr+"/edit", http.StatusSeeOther)
}
