package ollama

// import (
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestClient_Generate(t *testing.T) {
// 	// 1. Create a mock Ollama API server using httptest
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Check request method and path if desired
// 		if r.Method != http.MethodPost {
// 			t.Errorf("Expected POST, got %s", r.Method)
// 		}

// 		// Read request body (optional check)
// 		var req GenerateRequest
// 		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 			http.Error(w, "cannot decode request", http.StatusBadRequest)
// 			return
// 		}

// 		// Mock a response from the Ollama API
// 		resp := GenerateResponse{
// 			Response: "This is a mocked completion",
// 		}
// 		w.Header().Set("Content-Type", "application/json")
// 		if err := json.NewEncoder(w).Encode(resp); err != nil {
// 			http.Error(w, "cannot encode response", http.StatusInternalServerError)
// 		}
// 	})
// 	server := httptest.NewServer(handler)
// 	defer server.Close()

// 	// 2. Create a new Ollama client with the mock server's URL
// 	c := NewClient(
// 		WithBaseURL(server.URL),
// 	)

// 	temperature := 0.7

// 	// 3. Prepare a request
// 	req := GenerateRequest{
// 		Prompt:      "Hello, world!",
// 		Model:       "gemma2",
// 		Temperature: &temperature,
// 	}

// 	// 4. Call the client to generate a response
// 	resp, err := c.Generate(context.Background(), req)
// 	if err != nil {
// 		t.Fatalf("Generate call failed: %v", err)
// 	}

// 	// 5. Verify results
// 	expected := "This is a mocked completion"
// 	if resp.Response != expected {
// 		t.Errorf("Expected completion %q, got %q", expected, resp.Response)
// 	}
// }
