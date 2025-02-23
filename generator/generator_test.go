package generator

import (
	"errors"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name               string
		setupDelimiters    func() (string, func())
		expectError        bool
		expectedDelimiters []string
	}{
		{
			name: "valid delimiters file",
			setupDelimiters: func() (string, func()) {
				dir := t.TempDir()
				filePath := filepath.Join(dir, knownDelimitersFilePath)
				os.WriteFile(filePath, []byte("<|begin_of_text|>\n<|end_of_text|>\n"), 0644)
				return dir, func() { os.Remove(filePath) }
			},
			expectError:        false,
			expectedDelimiters: []string{"<|begin_of_text|>", "<|end_of_text|>"},
		},
		{
			name: "empty delimiters file",
			setupDelimiters: func() (string, func()) {
				dir := t.TempDir()
				filePath := filepath.Join(dir, knownDelimitersFilePath)
				os.WriteFile(filePath, []byte(""), 0644)
				return dir, func() { os.Remove(filePath) }
			},
			expectError: true,
		},
		{
			name: "no delimiters file",
			setupDelimiters: func() (string, func()) {
				dir := t.TempDir()
				return dir, func() {}
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseDir, tearDown := tc.setupDelimiters()
			defer tearDown()

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(baseDir)

			gen, err := NewGenerator(logger)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				if gen != nil {
					t.Errorf("expected nil generator but got one")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if gen == nil {
					t.Errorf("expected non-nil generator but got nil")
				}
				if len(gen.knownDelimiters) != len(tc.expectedDelimiters) {
					t.Errorf("expected %v delimiters, got %v", tc.expectedDelimiters, gen.knownDelimiters)
				}
			}
		})
	}
}

func TestGetKnownDelimiters(t *testing.T) {
	logger := zap.NewNop()
	g := &Generator{
		knownDelimiters: []string{"<|begin_of_text|>", "<|end_header_id|>", "<|end_of_text|>"},
		logger:          logger,
	}

	tests := []struct {
		name       string
		generator  *Generator
		wantOutput []string
	}{
		{
			name:       "multiple delimiters",
			generator:  g,
			wantOutput: []string{"<|begin_of_text|>", "<|end_header_id|>", "<|end_of_text|>"},
		},
		{
			name:       "no delimiters",
			generator:  &Generator{logger: logger},
			wantOutput: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.generator.GetKnownDelimiters()
			if len(got) != len(tc.wantOutput) {
				t.Errorf("expected length %d, got %d", len(tc.wantOutput), len(got))
				return
			}
			for i, gotValue := range got {
				if gotValue != tc.wantOutput[i] {
					t.Errorf("expected value %s at index %d, got %s", tc.wantOutput[i], i, gotValue)
				}
			}
		})
	}
}

func TestGenerateCandidate(t *testing.T) {
	logger := zap.NewNop()
	tests := []struct {
		name        string
		generator   *Generator
		minItems    int
		maxItems    int
		checkOutput func(candidate Candidate) error
	}{
		{
			name: "valid candidate with delimiters",
			generator: &Generator{
				knownDelimiters: []string{"<|begin_of_text|>", "<|end_of_text|>"},
				logger:          logger,
			},
			minItems: 2,
			maxItems: 5,
			checkOutput: func(candidate Candidate) error {
				if len(candidate.Items) < 2 {
					return errors.New("candidate has fewer items than minimum")
				}
				if len(candidate.Message) == 0 {
					return errors.New("candidate Message is empty")
				}
				foundExpression := false
				for _, item := range candidate.Items {
					if item.Type == "expression" {
						foundExpression = true
					}
				}
				if !foundExpression {
					return errors.New("candidate contains no expression")
				}
				return nil
			},
		},
		{
			name: "no delimiters available",
			generator: &Generator{
				logger: logger,
			},
			minItems: 2,
			maxItems: 5,
			checkOutput: func(candidate Candidate) error {
				if len(candidate.Items) > 0 {
					return errors.New("expected no candidate items but got some")
				}
				if len(candidate.Message) > 0 {
					return errors.New("expected empty candidate Message but got some")
				}
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.generator.GenerateCandidate(tc.minItems, tc.maxItems)
			if err := tc.checkOutput(got); err != nil {
				t.Errorf("Test failed: %v", err)
			}
		})
	}
}
