package vercelapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serve(t *testing.T, handler http.HandlerFunc) (cleanup func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	baseURL = srv.URL
	return func() {
		srv.Close()
		baseURL = "https://api.vercel.com"
	}
}

// GetUser

func TestGetUser_Success(t *testing.T) {
	defer serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok_test" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(GetUserResponse{
			User: User{ID: "u_1", Email: "alice@example.com", Name: "Alice", Username: "alice"},
		})
	})()

	user, err := GetUser("tok_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u_1" || user.Email != "alice@example.com" {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestGetUser_EmptyToken(t *testing.T) {
	_, err := GetUser("")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestGetUser_APIError(t *testing.T) {
	defer serve(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(GetUserResponse{
			Error: Error{Code: "forbidden", Message: "invalid token"},
		})
	})()

	_, err := GetUser("bad_token")
	if err == nil {
		t.Error("expected error from API error response")
	}
	if err.Error() != "invalid token" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetUser_InvalidJSON(t *testing.T) {
	defer serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})()

	_, err := GetUser("tok_test")
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

