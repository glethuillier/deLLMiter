package client

import (
	"testing"
)

func TestGetSystemMessage(t *testing.T) {
	tests := []struct {
		name string
		role string
		want Message
	}{
		{
			name: "basic case",
			role: "assistant",
			want: Message{
				Role:    "assistant",
				Content: systemConstraints,
			},
		},
		{
			name: "empty role",
			role: "",
			want: Message{
				Role:    "",
				Content: systemConstraints,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSystemMessage(tt.role)
			if got != tt.want {
				t.Errorf("getSystemMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserMessage(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		content string
		want    Message
	}{
		{
			name:    "basic case",
			role:    "user",
			content: "Hello",
			want: Message{
				Role:    "user",
				Content: userConstraints + ": Hello",
			},
		},
		{
			name:    "empty content",
			role:    "user",
			content: "",
			want: Message{
				Role:    "user",
				Content: userConstraints + ": ",
			},
		},
		{
			name:    "empty role",
			role:    "",
			content: "Some Content",
			want: Message{
				Role:    "",
				Content: userConstraints + ": Some Content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUserMessage(tt.role, tt.content)
			if got != tt.want {
				t.Errorf("getUserMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPrompt(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		content   string
		want      Prompt
	}{
		{
			name:      "mistral model",
			modelName: "Mistral 7B",
			content:   "Hello",
			want: Prompt{
				Model:       "mistral 7b",
				Messages:    []Message{getSystemMessage("assistant"), getUserMessage("user", "Hello")},
				Temperature: 0.8,
				Stream:      false,
			},
		},
		{
			name:      "hermes model",
			modelName: "Hermes-3",
			content:   "How are you?",
			want: Prompt{
				Model:       "hermes-3",
				Messages:    []Message{getUserMessage("user", "How are you?")},
				Temperature: 0.8,
				Stream:      false,
			},
		},
		{
			name:      "default model",
			modelName: "gpt-4",
			content:   "Tell me something",
			want: Prompt{
				Model:       "gpt-4",
				Messages:    []Message{getSystemMessage("system"), getUserMessage("user", "Tell me something")},
				Temperature: 0.8,
				Stream:      false,
			},
		},
		{
			name:      "trims whitespace in modelName",
			modelName: "  GPT-4 ",
			content:   "Some text",
			want: Prompt{
				Model:       "gpt-4",
				Messages:    []Message{getSystemMessage("system"), getUserMessage("user", "Some text")},
				Temperature: 0.8,
				Stream:      false,
			},
		},
		{
			name:      "empty modelName",
			modelName: "",
			content:   "Empty model",
			want: Prompt{
				Model:       "",
				Messages:    []Message{getSystemMessage("system"), getUserMessage("user", "Empty model")},
				Temperature: 0.8,
				Stream:      false,
			},
		},
		{
			name:      "empty content",
			modelName: "hermes-3",
			content:   "",
			want: Prompt{
				Model:       "hermes-3",
				Messages:    []Message{getUserMessage("user", "")},
				Temperature: 0.8,
				Stream:      false,
			},
		},
		{
			name:      "case insensitive modelName",
			modelName: "MiStRaL",
			content:   "Check sensitivity",
			want: Prompt{
				Model:       "mistral",
				Messages:    []Message{getSystemMessage("assistant"), getUserMessage("user", "Check sensitivity")},
				Temperature: 0.8,
				Stream:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPrompt(tt.modelName, tt.content)

			if got.Model != tt.want.Model {
				t.Errorf("getPrompt().Model = %v, want %v", got.Model, tt.want.Model)
			}
			if got.Temperature != tt.want.Temperature {
				t.Errorf("getPrompt().Temperature = %v, want %v", got.Temperature, tt.want.Temperature)
			}
			if got.Stream != tt.want.Stream {
				t.Errorf("getPrompt().Stream = %v, want %v", got.Stream, tt.want.Stream)
			}
			if len(got.Messages) != len(tt.want.Messages) {
				t.Fatalf("getPrompt().Messages length = %v, want %v", len(got.Messages), len(tt.want.Messages))
			}
			for i, msg := range got.Messages {
				if msg != tt.want.Messages[i] {
					t.Errorf("getPrompt().Messages[%d] = %v, want %v", i, msg, tt.want.Messages[i])
				}
			}
		})
	}
}
