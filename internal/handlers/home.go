package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) HomeHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "home.html", nil)
}

func (env *Env) ContestsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contests, _ := env.Queries.ListContests(ctx)

	render(w, r, "contests.html", map[string]interface{}{
		"Contests": contests,
	})
}

func (env *Env) ContestDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	contest, err := env.Queries.GetContestByID(ctx, id)
	if err != nil {
		http.Error(w, "Contest not found", http.StatusNotFound)
		return
	}

	problems, _ := env.Queries.ListContestProblems(ctx, id)

	var standings interface{}
	if contest.RankingType == "ICPC" {
		standings, _ = env.Queries.CalculateContestStandingsICPC(ctx, id)
	} else {
		standings, _ = env.Queries.CalculateContestStandingsIOI(ctx, id)
	}

	render(w, r, "contest_detail.html", map[string]interface{}{
		"Contest":   contest,
		"Problems":  problems,
		"Standings": standings,
	})
}

func (env *Env) UsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch top 100 users for leaderboard
	users, err := env.Queries.GetUsersLeaderboard(ctx, sqlcdb.GetUsersLeaderboardParams{
		LimitVal:  100,
		OffsetVal: 0,
	})

	if err != nil {
		http.Error(w, "Failed to load leaderboard", http.StatusInternalServerError)
		return
	}

	render(w, r, "users.html", map[string]interface{}{
		"Users": users,
	})
}

func (env *Env) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")

	user, err := env.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	solved, _ := env.Queries.GetUserSolvedProblems(ctx, user.ID)

	recentSubmissions, _ := env.Queries.ListSubmissionsByUser(ctx, sqlcdb.ListSubmissionsByUserParams{
		UserID:    user.ID,
		LimitVal:  10,
		OffsetVal: 0,
	})

	render(w, r, "user_profile.html", map[string]interface{}{
		"ProfileUser":       user,
		"SolvedCount":       len(solved),
		"RecentSubmissions": recentSubmissions,
	})
}
