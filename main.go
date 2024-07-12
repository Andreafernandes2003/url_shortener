package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/glebarez/go-sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

var db *sql.DB

func main() {
	dbFile := "./URL.db" // SQLite database file path
	initDB(dbFile)       // Initialize the database
	defer db.Close()

	router := chi.NewRouter()

	// Basic CORS settings
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Use actual origins in production
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		MaxAge:         300,
	})
	router.Use(cors.Handler)

	// Define routes
	router.Post("/shorten", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			OriginalURL string `json:"original_url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		log.Println("Received URL to shorten:", request.OriginalURL)

		shortURL := generateShortURL(request.OriginalURL)
		shortURLFull := "http://localhost:9091/" + shortURL // Construct the full short URL

		// Save the URL in the database
		newURL := ShortURL{
			ShortURL:    shortURL,
			OriginalURL: request.OriginalURL,
		}
		if err := saveURL(newURL); err != nil {
			http.Error(w, "Could not save URL", http.StatusInternalServerError)
			return
		}

		response := map[string]string{"link": shortURLFull}
		log.Println("Shortened URL:", shortURLFull)
		json.NewEncoder(w).Encode(response)
	})

	router.Get("/{shortURL}", func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "shortURL")

		// Retrieve the original URL from the database
		originalURL, err := findLongURL(shortURL)
		if err != nil {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
	})

	// Start the HTTP server on port 9091
	err := http.ListenAndServe(":9091", router)
	if err != nil {
		panic(err)
	}
}

func initDB(dbFile string) {
	var err error
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		panic("failed to connect database")
	}

	// Create the table if it doesn't exist
	createTable := `
        CREATE TABLE IF NOT EXISTS ShortURL (
            ShortURL TEXT UNIQUE,
            OriginalURL TEXT PRIMARY KEY
        );
    `
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}
}

func saveURL(newURL ShortURL) error {
	_, err := db.Exec("INSERT INTO ShortURL (ShortURL, OriginalURL) VALUES (?, ?) ON CONFLICT(OriginalURL) DO UPDATE SET ShortURL=excluded.ShortURL", newURL.ShortURL, newURL.OriginalURL)
	return err
}

func findLongURL(shortURL string) (string, error) {
	var originalURL string
	err := db.QueryRow("SELECT OriginalURL FROM ShortURL WHERE ShortURL = ?", shortURL).Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func generateShortURL(originalURL string) string {
	hasher := sha1.New()
	hasher.Write([]byte(originalURL))
	return hex.EncodeToString(hasher.Sum(nil))[:10]
}

type ShortURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
