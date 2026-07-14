package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db             *sql.DB
	authServiceURL string
)

type GuestBookEntry struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateEntryRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	dbPath := getEnv("DB_PATH", "/data/guestbook.db")
	authServiceURL = getEnv("AUTH_SERVICE_URL", "http://auth-service:8081")
	port := getEnv("PORT", "8082")

	var err error
	db, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("==========================================")
	fmt.Println("  GuestBook Booking Service")
	fmt.Println("==========================================")
	fmt.Printf("  Database:     %s\n", dbPath)
	fmt.Printf("  Auth Service: %s\n", authServiceURL)
	fmt.Println("==========================================")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/guestbook", corsMiddleware(handleGuestbook))
	mux.HandleFunc("/api/guestbook/", corsMiddleware(handleGuestbookByID))
	mux.HandleFunc("/healthz", handleHealth)

	fmt.Printf("Booking service listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS guestbook (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		address TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err == nil {
		log.Println("Database migration completed")
	}
	return err
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func handleGuestbook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listEntries(w, r)
	case http.MethodPost:
		createEntry(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
	}
}

func handleGuestbookByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	// Extract ID from path: /api/guestbook/{id}
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "missing id"})
		return
	}
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	deleteEntry(w, r, id)
}

func listEntries(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, address, message, created_at FROM guestbook ORDER BY created_at DESC")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to query entries"})
		return
	}
	defer rows.Close()

	entries := []GuestBookEntry{}
	for rows.Next() {
		var e GuestBookEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.Address, &e.Message, &e.CreatedAt); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	writeJSON(w, http.StatusOK, entries)
}

func createEntry(w http.ResponseWriter, r *http.Request) {
	if !verifyAuth(r) {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req CreateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || req.Address == "" || req.Message == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name, address, and message are required"})
		return
	}

	result, err := db.Exec(
		"INSERT INTO guestbook (name, address, message) VALUES (?, ?, ?)",
		req.Name, req.Address, req.Message,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create entry"})
		return
	}

	id, _ := result.LastInsertId()
	log.Printf("Created guestbook entry #%d by %s", id, req.Name)

	entry := GuestBookEntry{
		ID:        int(id),
		Name:      req.Name,
		Address:   req.Address,
		Message:   req.Message,
		CreatedAt: time.Now(),
	}
	writeJSON(w, http.StatusCreated, entry)
}

func deleteEntry(w http.ResponseWriter, r *http.Request, id int) {
	if !verifyAuth(r) {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	result, err := db.Exec("DELETE FROM guestbook WHERE id = ?", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to delete entry"})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "entry not found"})
		return
	}

	log.Printf("Deleted guestbook entry #%d", id)
	writeJSON(w, http.StatusOK, map[string]string{"message": "entry deleted"})
}

func verifyAuth(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", authServiceURL+"/api/auth/me", nil)
	if err != nil {
		log.Printf("Auth check error: %v", err)
		return false
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Auth service unreachable: %v", err)
		return false
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	return resp.StatusCode == http.StatusOK
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
