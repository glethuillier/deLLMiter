package analyzer

import (
	"strings"

	"github.com/glethuillier/deLLMiter/generator"
)

// threshold defines the minimum number of mismatched delimiter occurrences required to mark a delimiter as missing
// TODO: replace with a dynamic heuristic
const threshold = 10

// Analyzer is responsible for comparing the text sent to the model with its response
// to detect and categorize delimiters used by the model
// TODO: the Analyzer must be refactored to identify delimiters and higher-order expressions with more granularity
type Analyzer struct {
	MissingDelimiterCounts map[string]int
}

// NewAnalyzer creates and initializes a new Analyzer instance.
func NewAnalyzer() *Analyzer {
	return &Analyzer{MissingDelimiterCounts: make(map[string]int)}
}

// AreIdentical compares the generated candidate message with the model's response to check for equality,
// while identifying mismatched delimiters. It returns a boolean indicating equality, and a slice of delimiters
// that are present in the original but mismatched in the response.
// TODO: refactor to more robustly identify delimiters
func (a *Analyzer) AreIdentical(original generator.Candidate, response string) (bool, []string) {
	// TODO: implement a more flexible comparison
	if strings.EqualFold(original.Message, response) {
		return true, nil
	}

	uniqueDelimiters := make(map[string]int)

	for _, item := range original.Items {
		if item.Type == "delimiter" {
			uniqueDelimiters[item.Token]++
		}
	}

	var mismatchedDelimiters []string

	for delimiter, originalCount := range uniqueDelimiters {
		responseCount := strings.Count(response, delimiter)

		if originalCount != responseCount {
			mismatchedDelimiters = append(mismatchedDelimiters, delimiter)
			a.MissingDelimiterCounts[delimiter]++
		} else {
			a.MissingDelimiterCounts[delimiter] = 0
		}
	}

	var missingDelimiters []string

	for delimiter, count := range a.MissingDelimiterCounts {
		if count >= threshold {
			missingDelimiters = append(missingDelimiters, delimiter)
		}
	}

	return false, missingDelimiters
}
