package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ryannicoletti.net/config"
	"ryannicoletti.net/internal/models"
	"ryannicoletti.net/internal/validator"
	"ryannicoletti.net/ui"
)

type commentFormData struct {
	Name    string
	Website string
	Comment string
	validator.Validator
}

func home(app *config.Application) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts, err := app.Posts.GetAll()
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				http.NotFound(w, r)
			} else {
				serverError(app, w, r, err)
			}
			return
		}
		data := newTemplateData(r)
		data.Posts = posts
		render(w, r, app, http.StatusOK, "home.tmpl", data)
	})
}

func postView(app *config.Application) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			http.NotFound(w, r)
			return
		}
		post, err := app.Posts.GetBySlug(slug)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				http.NotFound(w, r)
			} else {
				serverError(app, w, r, err)
			}
			return
		}

		comments, err := app.Comments.GetByPostSlug(post.Slug)
		if err != nil {
			app.Logger.Warn("Error loading comments", "post", post.Slug, "error", err)
			comments = []models.Comment{}
		}

		data := newTemplateData(r)
		data.Post = post
		data.Comments = comments
		data.Form = commentFormData{}
		render(w, r, app, http.StatusOK, "post.tmpl", data)
	})
}

func createComment(app *config.Application) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			clientError(w, http.StatusBadRequest)
			return
		}
		// reject comment if it was filled out instantly (stupid bots)
		formTime := r.PostForm.Get("form_time")
		if formTime != "" {
			if t, err := strconv.ParseInt(formTime, 10, 64); err == nil {
				if time.Now().Unix()-t < 3 {
					http.Redirect(w, r, fmt.Sprintf("/post/%s", r.PostForm.Get("post_slug")), http.StatusSeeOther)
					return
				}
			}
		}
		// reject comment if this field is filled, honey pot for dumb bots
		if r.PostForm.Get("website_url") != "" {
			http.Redirect(w, r, fmt.Sprintf("/post/%s", r.PostForm.Get("post_slug")), http.StatusSeeOther)
			return
		}
		formData := commentFormData{
			Name:    r.PostForm.Get("name"),
			Comment: r.PostForm.Get("comment"),
			Website: r.PostForm.Get("website"),
		}

		formData.ValidateNotBlank(formData.Name, "name", "This field cannot be blank")
		formData.ValidateNotBlank(formData.Comment, "comment", "This field cannot be blank")
		formData.ValidateLength(formData.Name, 100, "name", "This field cannot be more than 100 characters long")
		url := formData.ValidateUrl(formData.Website, "website")

		postSlug := r.PostForm.Get("post_slug")

		if !formData.IsValid() {
			p, err := app.Posts.GetBySlug(postSlug)
			if err != nil {
				serverError(app, w, r, err)
				return
			}

			c, err := app.Comments.GetByPostSlug(p.Slug)
			if err != nil {
				app.Logger.Warn("Error loading comments for form", "post", p.Slug, "error", err)
				c = []models.Comment{}
			}

			data := newTemplateData(r)
			data.Post = p
			data.Comments = c
			data.Form = formData
			render(w, r, app, http.StatusUnprocessableEntity, "post.tmpl", data)
			return
		}

		p, err := app.Posts.GetBySlug(postSlug)
		if err != nil {
			serverError(app, w, r, err)
			return
		}
		id, err := app.Comments.Insert(formData.Name, &url, formData.Comment, postSlug)
		if err != nil {
			serverError(app, w, r, err)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/post/%s#comment-%s", p.Slug, id), http.StatusSeeOther)
	})
}

func about(app *config.Application) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := newTemplateData(r)
		render(w, r, app, http.StatusOK, "about.tmpl", data)
	})
}

func links(app *config.Application) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := newTemplateData(r)
		render(w, r, app, http.StatusOK, "links.tmpl", data)
	})
}

func robots() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		data, err := ui.Files.ReadFile("static/robots.txt")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	})
}
