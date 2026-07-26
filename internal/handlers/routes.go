package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

type Env struct {
	Pool    *pgxpool.Pool
	Queries *sqlcdb.Queries
}

func SetupRoutes(r chi.Router, env *Env) {
	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/", env.HomeHandler)
		r.Get("/problems", env.ProblemListHandler)
		r.Get("/problems/{slug}", env.ProblemDetailHandler)
		r.Get("/problems/{slug}/editorial", env.ProblemEditorialHandler)
		r.Get("/problems/{slug}/submit", env.SubmitGetHandler)
		r.With(SubmitLimiter.Middleware).Post("/problems/{slug}/submit", env.SubmitPostHandler)
		r.Get("/contests", env.ContestsHandler)
		r.Get("/contests/{id}", env.ContestDetailHandler)
		r.Get("/contests/{id}/problems/{slug}", env.ContestProblemDetailHandler)
		r.Get("/contests/{id}/problems/{slug}/submit", env.ContestSubmitGetHandler)
		r.With(SubmitLimiter.Middleware).Post("/contests/{id}/problems/{slug}/submit", env.ContestSubmitPostHandler)
		r.Get("/users", env.UsersHandler)
		r.Get("/users/{username}", env.UserProfileHandler)
		r.Get("/login", env.LoginGetHandler)
		r.With(LoginLimiter.Middleware).Post("/login", env.LoginPostHandler)
		r.Get("/logout", env.LogoutHandler)

		r.Get("/submissions", env.SubmissionListHandler)
		r.Get("/submissions/{id}", env.SubmissionDetailHandler)

		// Public Blog Routes
		r.Get("/blogs", env.BlogListHandler)
		r.Get("/blogs/{slug}", env.BlogDetailHandler)

		// Classroom Routes
		r.Get("/classes", env.ListClassesHandler)
		r.Get("/classes/{id}", env.ClassDetailHandler)
	})

	// Protected routes (requires auth)
	r.Group(func(r chi.Router) {
		r.Use(RequireAuth)
		r.Post("/problems/{slug}/submit", env.SubmitPostHandler)
		r.Post("/classes/{id}/join", env.ClassJoinHandler)
		r.Post("/classes/{id}/members", env.ClassManageMemberHandler)
		r.Post("/classes/{id}/update", env.ClassUpdateInfoHandler)
	})

	// Admin and Teacher routes (Teachers can manage problems and contests)
	r.Group(func(r chi.Router) {
		r.Use(RequireAuth)
		r.Use(env.RequireTeacherOrAdmin)
		r.Get("/admin", env.AdminDashboardHandler)
		
		r.Get("/admin/problems/create", env.AdminCreateProblemGetHandler)
		r.Post("/admin/problems/create", env.AdminCreateProblemPostHandler)
		r.Get("/admin/problems/{slug}/edit", env.AdminEditProblemGetHandler)
		r.Post("/admin/problems/{slug}/edit", env.AdminEditProblemPostHandler)
		r.Get("/admin/problems/{slug}/tests", env.AdminEditTestGetHandler)
		r.Post("/admin/problems/{slug}/tests", env.AdminEditTestPostHandler)

		r.Get("/admin/contests/create", env.AdminCreateContestGetHandler)
		r.Post("/admin/contests/create", env.AdminCreateContestPostHandler)
		r.Get("/admin/contests/{id}/edit", env.AdminEditContestGetHandler)
		r.Post("/admin/contests/{id}/edit", env.AdminEditContestPostHandler)
		r.Post("/admin/contests/{id}/problems/add", env.AdminAddContestProblemPostHandler)
		r.Post("/admin/contests/{id}/problems/remove", env.AdminRemoveContestProblemPostHandler)
	})

	// System Admin routes (only true admins)
	r.Group(func(r chi.Router) {
		r.Use(RequireAuth)
		r.Use(RequireAdmin)
		
		r.Get("/admin/users/create", env.AdminCreateUserGetHandler)
		r.Post("/admin/users/create", env.AdminCreateUserPostHandler)

		// Admin Blog routes
		r.Get("/admin/blogs", env.AdminBlogsHandler)
		r.Get("/admin/blogs/new", env.AdminBlogNewGetHandler)
		r.Post("/admin/blogs/new", env.AdminBlogNewPostHandler)
		r.Get("/admin/blogs/{id}/edit", env.AdminBlogEditGetHandler)
		r.Post("/admin/blogs/{id}/edit", env.AdminBlogEditPostHandler)
		r.Post("/admin/blogs/{id}/delete", env.AdminBlogDeleteHandler)

		r.Get("/admin/groups/create", env.AdminCreateGroupGetHandler)
		r.Get("/admin/announcements/create", env.AdminCreateAnnouncementGetHandler)
		
		// Admin Class Routes
		r.Get("/admin/classes/create", env.AdminCreateClassGetHandler)
		r.Post("/admin/classes/create", env.AdminCreateClassPostHandler)

		// Judge Nodes
		r.Get("/admin/judges", env.AdminJudgesListHandler)
		r.Get("/admin/judges/create", env.AdminJudgeCreateGetHandler)
		r.Post("/admin/judges/create", env.AdminJudgeCreatePostHandler)
		r.Get("/admin/judges/{id}/edit", env.AdminJudgeEditGetHandler)
		r.Post("/admin/judges/{id}/edit", env.AdminJudgeEditPostHandler)
		r.Post("/admin/judges/{id}/delete", env.AdminJudgeDeletePostHandler)
	})
}
