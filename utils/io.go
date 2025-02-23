package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glethuillier/deLLMiter/generator"
)

const resultDir = "./results"

func SaveResult(modelName string, candidate generator.Candidate, response string) error {
	if err := os.MkdirAll(resultDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	fileName := filepath.Join(resultDir, fmt.Sprintf("%s_all.txt", modelName))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("warning: failed to close file: %v\n", cerr)
		}
	}()

	var delimiters, expressions []string
	for _, item := range candidate.Items {
		switch item.Type {
		case "delimiter":
			delimiters = append(delimiters, item.Token)
		case "expression":
			expressions = append(expressions, item.Token)
		}
	}

	logEntry := fmt.Sprintf(
		"Sent	: %s\nReceived: %s\nDelimiters: %v\nExpressions: %v\n\n",
		candidate.Message, response, delimiters, expressions,
	)
	if _, writeErr := file.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write log entry to file: %w", writeErr)
	}

	return nil
}

func SaveDelimiters(modelName string, delimiters []string) error {
	// ensure delimiters are unique and sorted
	uniqueDelimiters := make(map[string]struct{}, len(delimiters))
	for _, d := range delimiters {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			uniqueDelimiters[trimmed] = struct{}{}
		}
	}

	if err := os.MkdirAll(resultDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	fileName := filepath.Join(resultDir, fmt.Sprintf("%s_delimiters.txt", modelName))
	existingDelimiters := make(map[string]struct{})
	if file, err := os.Open(fileName); err == nil {
		defer func() {
			if cerr := file.Close(); cerr != nil {
				fmt.Printf("warning: failed to close file: %v\n", cerr)
			}
		}()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text != "" {
				existingDelimiters[text] = struct{}{}
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading existing delimiters file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to open existing delimiters file: %w", err)
	}

	// merge existing delimiters with new ones
	for d := range existingDelimiters {
		uniqueDelimiters[d] = struct{}{}
	}

	sortedDelimiters := make([]string, 0, len(uniqueDelimiters))
	for d := range uniqueDelimiters {
		sortedDelimiters = append(sortedDelimiters, d)
	}
	sort.Strings(sortedDelimiters)

	file, err := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %w", fileName, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("warning: failed to close file: %v\n", cerr)
		}
	}()

	// write unique and sorted delimiters to the file
	for _, delimiter := range sortedDelimiters {
		if _, err := file.WriteString(delimiter + "\n"); err != nil {
			return fmt.Errorf("failed to write delimiter to file: %w", err)
		}
	}

	return nil
}
