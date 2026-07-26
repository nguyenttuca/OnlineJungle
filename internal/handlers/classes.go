package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) ListClassesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	classes, err := env.Queries.ListClasses(ctx)
	if err != nil {
		classes = []sqlcdb.Class{}
	}

	render(w, r, "classes.html", map[string]interface{}{
		"Classes": classes,
	})
}

func (env *Env) ClassDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	class, err := env.Queries.GetClassByID(ctx, id)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}

	members, _ := env.Queries.GetClassMembers(ctx, id)
	problems, _ := env.Queries.ListClassProblems(ctx, id)
	contests, _ := env.Queries.ListClassContests(ctx, id)

	var userRole string
	user := GetUserFromContext(r.Context())
	if user != nil {
		if user.Role == "admin" {
			userRole = "admin" // System admin has full access
		} else {
			role, err := env.Queries.GetUserRoleInClass(ctx, sqlcdb.GetUserRoleInClassParams{
				ClassID: id,
				UserID:  user.ID,
			})
			if err == nil {
				userRole = role
			}
		}
	}

	render(w, r, "class_detail.html", map[string]interface{}{
		"Class":    class,
		"Members":  members,
		"Problems": problems,
		"Contests": contests,
		"UserRole": userRole, // "admin", "teacher", "student", "pending", or ""
	})
}

func (env *Env) ClassJoinHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := env.Queries.AddClassMember(ctx, sqlcdb.AddClassMemberParams{
		ClassID: id,
		UserID:  user.ID,
		Role:    "pending",
	})
	if err != nil {
		log.Println("Error joining class:", err)
	}

	http.Redirect(w, r, "/classes/"+idStr, http.StatusSeeOther)
}

func (env *Env) ClassManageMemberHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	classID, _ := strconv.ParseInt(idStr, 10, 64)

	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is admin or teacher of this class
	if user.Role != "admin" {
		role, err := env.Queries.GetUserRoleInClass(ctx, sqlcdb.GetUserRoleInClassParams{
			ClassID: classID,
			UserID:  user.ID,
		})
		if err != nil || role != "teacher" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	action := r.FormValue("action")
	targetUserIDStr := r.FormValue("user_id")
	targetUserID, _ := strconv.ParseInt(targetUserIDStr, 10, 64)

	switch action {
	case "approve":
		env.Queries.UpdateClassMemberRole(ctx, sqlcdb.UpdateClassMemberRoleParams{
			ClassID: classID,
			UserID:  targetUserID,
			Role:    "student",
		})
	case "reject", "remove":
		env.Queries.RemoveClassMember(ctx, sqlcdb.RemoveClassMemberParams{
			ClassID: classID,
			UserID:  targetUserID,
		})
	case "promote":
		env.Queries.UpdateClassMemberRole(ctx, sqlcdb.UpdateClassMemberRoleParams{
			ClassID: classID,
			UserID:  targetUserID,
			Role:    "teacher",
		})
	case "demote":
		env.Queries.UpdateClassMemberRole(ctx, sqlcdb.UpdateClassMemberRoleParams{
			ClassID: classID,
			UserID:  targetUserID,
			Role:    "student",
		})
	}

	http.Redirect(w, r, "/classes/"+idStr+"?tab=members", http.StatusSeeOther)
}

func (env *Env) ClassUpdateInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	classID, _ := strconv.ParseInt(idStr, 10, 64)

	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" {
		role, err := env.Queries.GetUserRoleInClass(ctx, sqlcdb.GetUserRoleInClassParams{
			ClassID: classID,
			UserID:  user.ID,
		})
		if err != nil || role != "teacher" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	class, err := env.Queries.GetClassByID(ctx, classID)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}

	// We only update schedule or notice based on the form
	name := r.FormValue("name")
	description := r.FormValue("description")
	weeklyScheduleStr := r.FormValue("weekly_schedule")
	noticeMd := r.FormValue("notice_md")

	if name == "" {
		name = class.Name
	}
	if description == "" {
		description = class.Description
	}
	
	var weeklySchedule []byte
	if weeklyScheduleStr == "" && r.Form.Has("weekly_schedule") {
		weeklySchedule = []byte("[]")
	} else if weeklyScheduleStr == "" {
		weeklySchedule = class.WeeklySchedule
	} else {
		weeklySchedule = []byte(weeklyScheduleStr)
	}
	
	if noticeMd == "" && r.Form.Has("notice_md") {
		noticeMd = ""
	} else if noticeMd == "" {
		noticeMd = class.NoticeMd
	}

	env.Queries.UpdateClass(ctx, sqlcdb.UpdateClassParams{
		ID:             classID,
		Name:           name,
		Description:    description,
		WeeklySchedule: weeklySchedule,
		NoticeMd:       noticeMd,
	})

	http.Redirect(w, r, "/classes/"+idStr, http.StatusSeeOther)
}

func (env *Env) AdminCreateClassGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_class_create.html", nil)
}

func (env *Env) AdminCreateClassPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	ctx := r.Context()
	class, err := env.Queries.CreateClass(ctx, sqlcdb.CreateClassParams{
		Name:           name,
		Description:    description,
		WeeklySchedule: []byte("[]"),
		NoticeMd:       "",
	})

	if err != nil {
		http.Error(w, "Failed to create class", http.StatusInternalServerError)
		return
	}

	// Add the current admin as teacher to the class
	user := GetUserFromContext(r.Context())
	if user != nil {
		env.Queries.AddClassMember(ctx, sqlcdb.AddClassMemberParams{
			ClassID: class.ID,
			UserID:  user.ID,
			Role:    "teacher",
		})
	}

	http.Redirect(w, r, "/classes/"+strconv.FormatInt(class.ID, 10), http.StatusSeeOther)
}
