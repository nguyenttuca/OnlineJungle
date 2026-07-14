package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/justinas/nosurf"
)

var templates = make(map[string]*template.Template)

func InitTemplates(templatesDir string) error {
	layoutFiles, err := filepath.Glob(filepath.Join(templatesDir, "layouts", "*.html"))
	if err != nil {
		return err
	}
	pageFiles, err := filepath.Glob(filepath.Join(templatesDir, "pages", "*.html"))
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"truncatePreview": func(s string) string {
			if len(s) > 300 {
				return s[:300] + "\n... [Data too large, editing disabled]"
			}
			return s
		},
	}

	for _, pageFile := range pageFiles {
		name := filepath.Base(pageFile)
		files := append(layoutFiles, pageFile)

		t := template.New(name).Funcs(funcMap)
		t, err := t.ParseFiles(files...)
		if err != nil {
			return err
		}
		templates[name] = t
	}
	return nil
}

func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	t, ok := templates[tmpl]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if data == nil {
		data = map[string]interface{}{}
	}

	user := GetUserFromContext(r.Context())

	err := t.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"Page":      tmpl,
		"Data":      data,
		"User":      user,
		"CsrfToken": nosurf.Token(r),
	})
	if err != nil {
		log.Printf("Template execution error for %s: %v", tmpl, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
