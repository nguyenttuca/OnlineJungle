package handlers

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func (env *Env) BlogListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	blogs, err := env.Queries.GetBlogs(ctx)
	if err != nil {
		http.Error(w, "Failed to load blogs", http.StatusInternalServerError)
		return
	}

	render(w, r, "blogs.html", map[string]interface{}{
		"Blogs": blogs,
	})
}

func (env *Env) BlogDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	blog, err := env.Queries.GetBlogBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	// Setup markdown parser and HTML renderer
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(blog.ContentMd))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlContent := string(markdown.Render(doc, renderer))

	render(w, r, "blog_detail.html", map[string]interface{}{
		"Blog":        blog,
		"HTMLContent": template.HTML(htmlContent),
	})
}
