package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) SubmissionListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch recent submissions
	submissions, err := env.Queries.ListAllSubmissions(ctx, sqlcdb.ListAllSubmissionsParams{
		LimitVal:  50,
		OffsetVal: 0,
	})

	if err != nil {
		http.Error(w, "Failed to load submissions", http.StatusInternalServerError)
		return
	}

	render(w, r, "submissions.html", map[string]interface{}{
		"Submissions": submissions,
	})
}

func (env *Env) SubmissionDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	submission, err := env.Queries.GetSubmissionByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	testResults, _ := env.Queries.ListTestResultsBySubmission(ctx, id)

	problem, err := env.Queries.GetProblemByID(ctx, submission.ProblemID)
	if err != nil {
		log.Printf("Error getting problem for submission: %v", err)
	}

	user := GetUserFromContext(ctx)
	canViewSource := false
	if user != nil && (user.ID == submission.UserID || user.Role == "admin") {
		canViewSource = true
	}

	render(w, r, "submission_detail.html", map[string]interface{}{
		"Submission":    submission,
		"Problem":       problem,
		"TestResults":   testResults,
		"CanViewSource": canViewSource,
	})
}
