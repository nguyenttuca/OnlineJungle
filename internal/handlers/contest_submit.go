package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) ContestSubmitPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

	language := r.FormValue("language")
	sourceCode := r.FormValue("source_code")

	if language == "" || sourceCode == "" {
		http.Error(w, "Language and source code are required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	if now.Before(contest.StartAt.Time) {
		http.Error(w, "Contest chưa bắt đầu", http.StatusForbidden)
		return
	}

	// Determine if it's counted for contest
	var dbContestID *int64
	if now.After(contest.StartAt.Time) && now.Before(contest.EndAt.Time) {
		dbContestID = &contestID
	}

	submission, err := env.Queries.CreateSubmission(ctx, sqlcdb.CreateSubmissionParams{
		UserID:      user.ID,
		ProblemID:   problem.ID,
		ContestID:   dbContestID,
		Language:    language,
		SourceCode:  sourceCode,
		CodeSize:    int32(len(sourceCode)),
		RunAllTests: true,
		Score:       0,
	})

	if err != nil {
		http.Error(w, "Failed to submit", http.StatusInternalServerError)
		return
	}

	// Submission được tạo với status "queued". Dispatcher
	// (internal/dispatcher/dispatcher.go) tự poll mỗi 500ms và sẽ tự động
	// dequeue + gửi bài này cho worker chấm. KHÔNG gọi DequeueSubmission ở
	// đây — Dispatcher là nơi DUY NHẤT được phép đổi trạng thái
	// queued -> dispatched, tránh race condition khiến bài bị đánh dấu
	// dispatched nhưng không có worker nào thực sự nhận và xử lý.

	http.Redirect(w, r, "/submissions/"+strconv.FormatInt(submission.ID, 10), http.StatusSeeOther)
}
