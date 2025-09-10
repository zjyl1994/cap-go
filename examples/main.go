package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zjyl1994/cap-go"
)

// Simple in-memory storage implementation for demo
type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (m *MemoryStorage) Get(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	if !exists {
		return ""
	}
	return value
}

func (m *MemoryStorage) Set(key, data string, expire time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = data
	// Note: This simple implementation doesn't handle expiration
	// In a real implementation, you would set up a cleanup routine
}

func (m *MemoryStorage) Del(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

var capInstance = cap.NewCap(NewMemoryStorage())

//go:embed index.html
var indexHTML []byte

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Message: errorMsg,
	})
}

// homeHandler serves the index.html page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(indexHTML)
}

// challengeHandler creates a new challenge
func challengeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	challenge := capInstance.CreateChallenge(nil)

	token := challenge.Token
	log.Printf("Generated challenge token: %s\n", token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(challenge)
}

// redeemHandler validates a solution
func redeemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var sol cap.Solution
	if err := json.NewDecoder(r.Body).Decode(&sol); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	result := capInstance.RedeemChallenge(&sol)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// verifyTokenHandler validates a token
func verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	token := r.URL.Query().Get("token")
	result := capInstance.ValidateToken(token, false)
	log.Println("ValidateToken", token, "result:", result)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"token": token, "result": result})
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/cap/challenge", challengeHandler)
	http.HandleFunc("/cap/redeem", redeemHandler)
	http.HandleFunc("/verify-token", verifyTokenHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
