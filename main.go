package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var g_db *sql.DB

type URLShortener struct {
	Short_url string `json:"short_url"`
	Long_url  string `json:"long_url"`
}

func main() {

	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/test")
	if err != nil {
		panic(err.Error())
	}

	g_db = db

	defer db.Close()

	http.HandleFunc("/shorten", shorten)

	fmt.Println("Server starting at port: 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	original_url := r.FormValue("url")
	if original_url == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	short_url := generateShortURL()

	url_shortener := &URLShortener{
		Short_url: short_url,
		Long_url:  original_url,
	}

	_, err := addURL(url_shortener)
	if err != nil {
		http.Error(w, "Failed to save urls", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(url_shortener)

}

func generateShortURL() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err.Error())
	}

	return base64.StdEncoding.EncodeToString(b)
}

func addURL(urlshortener *URLShortener) (int64, error) {
	result, err := g_db.Exec("INSERT INTO urls (short_code, original_url) VALUES (?, ?)", urlshortener.Short_url, urlshortener.Long_url)
	if err != nil {
		return 0, fmt.Errorf("addURL: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addURL: %v", err)
	}

	return id, nil
}
