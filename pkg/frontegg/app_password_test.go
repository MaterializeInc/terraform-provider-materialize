package frontegg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

func setupMockServer() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(ApiTokenPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := `[{"clientId":"test-client-id","description":"Test description","owner":"test-owner","created_at":"2020-01-01T00:00:00Z","secret":"test-secret"}]`
		w.Write([]byte(response))
	})

	return httptest.NewServer(handler)
}

func TestListAppPasswords(t *testing.T) {
	mockServer := setupMockServer()
	defer mockServer.Close()

	client := &clients.FronteggClient{
		HTTPClient: &http.Client{},
		Endpoint:   mockServer.URL,
		Token:      "mock-token",
	}

	passwords, err := ListAppPasswords(context.Background(), client)
	if err != nil {
		t.Fatalf("ListAppPasswords returned an error: %v", err)
	}

	if len(passwords) != 1 {
		t.Fatalf("Expected 1 password, got %d", len(passwords))
	}

	expectedClientID := "test-client-id"
	if passwords[0].ClientID != expectedClientID {
		t.Errorf("Expected clientId %s, got %s", expectedClientID, passwords[0].ClientID)
	}
}
