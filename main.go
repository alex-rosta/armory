package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"wowchecker/pkg/handlers"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
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
