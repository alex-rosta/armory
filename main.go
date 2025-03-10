package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"wowchecker/pkg/handlers"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file, continuing without it. Ignore this if running as container.")
	}

	fileServer := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) == ".css" {
			w.Header().Set("Content-Type", "text/css")
		}
		fileServer.ServeHTTP(w, r)
	})))

	http.HandleFunc("/", handlers.LookupCharacter)
	fmt.Println("Listening on :http://localhost:3000")
	http.ListenAndServe(":3000", nil)
}
