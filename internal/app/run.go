package app

import (
	"log"
	"net/http"

	"URL_Shortener/internal/handlers"
	"URL_Shortener/internal/storage/postgres"
)

func Run() {
	err := postgres.Init()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handlers.RedirectHandler)
	http.HandleFunc("/create", handlers.CreateHandler)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
