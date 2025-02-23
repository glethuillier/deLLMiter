package generator

import (
	"bufio"
	"errors"
	"github.com/brianvoe/gofakeit/v7"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

const knownDelimitersFilePath = "known_delimiters.txt"

type Generator struct {
	knownDelimiters []string
	logger          *zap.Logger
}

func NewGenerator(logger *zap.Logger) (*Generator, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, errors.New("failed to get the current working directory: " + err.Error())
	}
	filePath := filepath.Join(baseDir, knownDelimitersFilePath)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("failed to open the known delimiters file: " + err.Error())
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			logger.Error("failed to close the known delimiters file", zap.Error(cerr))
		}
	}()

	var delimiters []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			delimiters = append(delimiters, text)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("error reading delimiters file: " + err.Error())
	}

	if len(delimiters) == 0 {
		return nil, errors.New("no delimiters found in the file")
	}

	return &Generator{knownDelimiters: delimiters}, nil
}

func (g *Generator) GetKnownDelimiters() []string {
	return append([]string{}, g.knownDelimiters...)
}

type Item struct {
	Type  ItemType
	Token string
}

type ItemType string

const (
	Delimiter  ItemType = "delimiter"
	Expression ItemType = "expression"
)

// TODO: find an alternative name to `Items`
type Candidate struct {
	Message string
	Items   []Item
}

func (g *Generator) GenerateCandidate(minItemsCount, maxItemsCount int) Candidate {
	if len(g.knownDelimiters) == 0 {
		g.logger.Error("no known delimiters available")
		return Candidate{}
	}

	var items []Item

	totalItems := rand.Intn(maxItemsCount) + minItemsCount
	hasExpression := false

	for i := 0; i < totalItems; i++ {
		if rand.Intn(5) < 1 && len(g.knownDelimiters) > 0 {
			delimiter := g.knownDelimiters[rand.Intn(len(g.knownDelimiters))]
			items = append(items, Item{Type: "delimiter", Token: delimiter})
		} else {
			word := gofakeit.Word()
			items = append(items, Item{Type: "expression", Token: word})
			hasExpression = true
		}
	}

	// ensure at least one expression exists
	if !hasExpression {
		word := gofakeit.Word()
		items = append(items, Item{Type: "expression", Token: word})
	}

	return Candidate{
		Message: strings.Join(func() []string {
			tokens := make([]string, 0, len(items))
			for _, item := range items {
				tokens = append(tokens, item.Token)
			}
			return tokens
		}(), " "),
		Items: items,
	}
}
