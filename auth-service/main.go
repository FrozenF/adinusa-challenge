package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	sessionDir   string
	adminUser    string
	adminPass    string
)

type Session struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MeResponse struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	sessionDir = getEnv("SESSION_DIR", "/tmp/sessions")
	adminUser = getEnv("ADMIN_USER", "admin")
	adminPass = getEnv("ADMIN_PASS", "admin123")
	port := getEnv("PORT", "8081")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		log.Fatalf("Failed to create session directory: %v", err)
	}

	fmt.Println("==========================================")
	fmt.Println("  GuestBook Auth Service")
	fmt.Println("==========================================")
	fmt.Printf("  Default Admin User: %s\n", adminUser)
	fmt.Printf("  Default Admin Pass: %s\n", adminPass)
	fmt.Printf("  Session Directory:  %s\n", sessionDir)
	fmt.Println("==========================================")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", corsMiddleware(handleLogin))
	mux.HandleFunc("/api/auth/logout", corsMiddleware(handleLogout))
	mux.HandleFunc("/api/auth/me", corsMiddleware(handleMe))
	mux.HandleFunc("/healthz", handleHealth)

	fmt.Printf("Auth service listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Username != adminUser || req.Password != adminPass {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		return
	}

	token := uuid.New().String()
	session := Session{
		Token:     token,
		Username:  req.Username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	data, _ := json.Marshal(session)
	sessionPath := filepath.Join(sessionDir, token+".json")
	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create session"})
		return
	}

	log.Printf("Login successful for user: %s", req.Username)
	writeJSON(w, http.StatusOK, LoginResponse{
		Token:    token,
		Username: req.Username,
		Message:  "login successful",
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	token := extractToken(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "no token provided"})
		return
	}

	sessionPath := filepath.Join(sessionDir, token+".json")
	if err := os.Remove(sessionPath); err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "session not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to remove session"})
		return
	}

	log.Printf("Logout successful for token: %s...", token[:8])
	writeJSON(w, http.StatusOK, map[string]string{"message": "logout successful"})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	token := extractToken(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "no token provided"})
		return
	}

	sessionPath := filepath.Join(sessionDir, token+".json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "session not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to read session"})
		return
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "invalid session data"})
		return
	}

	if time.Now().After(session.ExpiresAt) {
		os.Remove(sessionPath)
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "session expired"})
		return
	}

	writeJSON(w, http.StatusOK, MeResponse{
		Username: session.Username,
		Message:  "authenticated",
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
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
