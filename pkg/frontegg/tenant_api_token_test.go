package frontegg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

func setupTenantApiTokenMockServer() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(TenantApiTokenPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := `[{"clientId":"test-client-id","description":"Test description","secret":"test-secret","createdByUserId":"test-user-id","createdAt":"2020-01-01T00:00:00Z","roleIds":[]}]`
		w.Write([]byte(response))
	})

	return httptest.NewServer(handler)
}

func TestListTenantApiTokens(t *testing.T) {
	mockServer := setupTenantApiTokenMockServer()
	defer mockServer.Close()

	client := &clients.FronteggClient{
		HTTPClient: &http.Client{},
		Endpoint:   mockServer.URL,
		Token:      "mock-token",
	}

	tokens, err := ListTenantApiTokens(context.Background(), client)
	if err != nil {
		t.Fatalf("ListTenantApiTokens returned an error: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("Expected 1 token, got %d", len(tokens))
	}

	expectedClientID := "test-client-id"
	if tokens[0].ClientID != expectedClientID {
		t.Errorf("Expected clientId %s, got %s", expectedClientID, tokens[0].ClientID)
	}
}
