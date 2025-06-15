// main.go - Tags API Server
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"josepshbrain-go/handlers"
	"josepshbrain-go/repository"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string from environment variable or default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	tagRepo := repository.NewTagRepository(db)
	tagHandler := handlers.NewTagHandler(tagRepo)

	http.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tagHandler.GetTags(w, r)
		case http.MethodPost:
			tagHandler.CreateTag(w, r)
		case http.MethodPut:
			tagHandler.UpdateTag(w, r)
		case http.MethodDelete:
			tagHandler.DeleteTag(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Tags API server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}