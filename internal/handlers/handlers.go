package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"URL_Shortener/internal/storage/postgres"
)

type URLData struct {
	ShortURL string
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(originalURL); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	var shortKey string
	if check, err := postgres.GetShortKey(originalURL); err == nil && check != "" {
		shortKey = check
	} else {
		shortKey, err = postgres.SaveURL(originalURL)

		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			log.Println("SaveURL error:", err)
			return
		}
	}

	shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortKey)
	data := URLData{ShortURL: shortURL}
	tmpl, err := template.ParseFiles("templates/short.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		log.Println("Template parse error:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		log.Println("Template execute error:", err)
		return
	}
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "templates/form.html")
		return
	}

	shortKey := r.URL.Path[1:]
	originalURL, err := postgres.GetOriginalURL(shortKey)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Server error", http.StatusInternalServerError)
			log.Println("Database error:", err)
		}
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}
