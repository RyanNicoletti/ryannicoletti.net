package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"ryannicoletti.net/config"
	"ryannicoletti.net/internal/models"
	"ryannicoletti.net/ui"
)

type templateData struct {
	CurrentYear int
	CurrentTime int64
	Post        models.Post
	Posts       []models.Post
	Comments    []models.Comment
	Form        any
	PageName    string
}

func newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
		CurrentTime: time.Now().Unix(),
	}
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		patterns := []string{"html/base.tmpl", "html/partials/*.tmpl", page}
		ts, err := template.New(name).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	return cache, nil
}

func render(w http.ResponseWriter, r *http.Request, app *config.Application, status int, page string, data templateData) {
	name := strings.TrimSuffix(filepath.Base(page), filepath.Ext(page))
	data.PageName = name
	var ts *template.Template
	var err error
	if app.Environment == "production" {
		var ok bool
		ts, ok = app.TemplateCache[page]
		if !ok {
			err := fmt.Errorf("the template %s does not exist", page)
			serverError(app, w, r, err)
			return
		}
	} else {
		ts, err = template.New(name).ParseFS(ui.Files, "html/base.tmpl", "html/partials/*.tmpl", "html/pages/"+page)
		if err != nil {
			serverError(app, w, r, err)
			return
		}
	}

	buf := new(bytes.Buffer)
	err = ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		serverError(app, w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}
