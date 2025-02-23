package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		baseURL        string
		requestedModel string
		apiResponse    string
		statusCode     int
		expectedError  string
	}{
		{
			name:           "ValidModel",
			baseURL:        "http://mockserver",
			requestedModel: "model_1",
			apiResponse:    `{"object":"list","data":[{"id":"model_1","object":"model","type":"type1"}]}`,
			statusCode:     http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "InvalidModel",
			baseURL:        "http://mockserver",
			requestedModel: "unknown_model",
			apiResponse:    `{"object":"list","data":[{"id":"model_1","object":"model","type":"type1"}]}`,
			statusCode:     http.StatusOK,
			expectedError:  "model unknown_model not supported/loaded by the API server",
		},
		{
			name:           "ServerError",
			baseURL:        "http://mockserver",
			requestedModel: "model_1",
			apiResponse:    ``,
			statusCode:     http.StatusInternalServerError,
			expectedError:  "models API returned non-200 status: 500",
		},
		{
			name:           "MalformedJSON",
			baseURL:        "http://mockserver",
			requestedModel: "model_1",
			apiResponse:    `invalid_json`,
			statusCode:     http.StatusOK,
			expectedError:  "failed to unmarshal response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.apiResponse))
			}))
			defer server.Close()

			client, err := NewClient(server.URL, tc.requestedModel)

			if tc.expectedError == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if client == nil {
					t.Fatal("expected a valid client instance, got nil")
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("expected error containing: %v, got: %v", tc.expectedError, err)
				}
			}
		})
	}
}

func TestClient_Query(t *testing.T) {
	tests := []struct {
		name             string
		setupServer      func() *httptest.Server
		modelName        string
		messageContent   string
		expectedResponse string
		expectedError    string
	}{
		{
			name: "SuccessfulQuery",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					var reqBody Prompt
					if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
						http.Error(w, "invalid request", http.StatusBadRequest)
						return
					}

					if reqBody.Model != "model_1" ||
						(!strings.Contains(reqBody.Messages[0].Content, "hello") && !strings.Contains(reqBody.Messages[1].Content, "hello")) {
						http.Error(w, "invalid request payload", http.StatusBadRequest)
						return
					}
					response := Response{
						Choices: []Choice{
							{Message: Message{Content: "response"}},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(response)
				}))
			},
			modelName:        "model_1",
			messageContent:   "hello",
			expectedResponse: "response",
			expectedError:    "",
		},
		{
			name: "ServerError",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}))
			},
			modelName:        "model_1",
			messageContent:   "hello",
			expectedResponse: "",
			expectedError:    "non-200 response received: 500",
		},
		{
			name: "InvalidJSONResponse",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`invalid_json`))
				}))
			},
			modelName:        "model_1",
			messageContent:   "hello",
			expectedResponse: "",
			expectedError:    "failed to unmarshal response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()
			defer server.Close()

			client := &Client{
				queryURL:   server.URL,
				httpClient: &http.Client{},
			}
			response, err := client.Query(tc.modelName, tc.messageContent)

			if tc.expectedError == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if response != tc.expectedResponse {
					t.Fatalf("expected response: %s, got: %s", tc.expectedResponse, response)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("expected error containing: %v, got: %v", tc.expectedError, err)
				}
			}
		})
	}
}
