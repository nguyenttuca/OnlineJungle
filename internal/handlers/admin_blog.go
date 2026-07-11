package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

func (env *Env) AdminBlogsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	blogs, err := env.Queries.GetAllBlogsAdmin(ctx)
	if err != nil {
		http.Error(w, "Failed to load blogs", http.StatusInternalServerError)
		return
	}

	render(w, r, "admin_blogs.html", map[string]interface{}{
		"Blogs": blogs,
	})
}

func (env *Env) AdminBlogNewGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_blog_form.html", nil)
}

func (env *Env) AdminBlogNewPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := r.Context().Value(userContextKey).(sqlcdb.User)
	userID := user.ID

	title := r.FormValue("title")
	slug := r.FormValue("slug")
	contentMd := r.FormValue("content_md")
	isPublished := r.FormValue("is_published") == "on"

	if title == "" || slug == "" || contentMd == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	_, err := env.Queries.CreateBlog(ctx, sqlcdb.CreateBlogParams{
		Title:       title,
		Slug:        slug,
		ContentMd:   contentMd,
		AuthorID:    userID,
		IsPublished: isPublished,
	})

	if err != nil {
		http.Error(w, "Failed to create blog", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/blogs", http.StatusSeeOther)
}

func (env *Env) AdminBlogEditGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	blog, err := env.Queries.GetBlogByIDAdmin(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Blog not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to load blog", http.StatusInternalServerError)
		}
		return
	}

	render(w, r, "admin_blog_form.html", map[string]interface{}{
		"Blog": blog,
	})
}

func (env *Env) AdminBlogEditPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	title := r.FormValue("title")
	slug := r.FormValue("slug")
	contentMd := r.FormValue("content_md")
	isPublished := r.FormValue("is_published") == "on"

	if title == "" || slug == "" || contentMd == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	err := env.Queries.UpdateBlog(ctx, sqlcdb.UpdateBlogParams{
		ID:          id,
		Title:       title,
		Slug:        slug,
		ContentMd:   contentMd,
		IsPublished: isPublished,
	})

	if err != nil {
		http.Error(w, "Failed to update blog", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/blogs", http.StatusSeeOther)
}

func (env *Env) AdminBlogDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	err := env.Queries.DeleteBlog(ctx, id)
	if err != nil {
		http.Error(w, "Failed to delete blog", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/blogs", http.StatusSeeOther)
}
