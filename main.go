package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/glethuillier/deLLMiter/analyzer"
	"github.com/glethuillier/deLLMiter/client"
	"github.com/glethuillier/deLLMiter/generator"
	"github.com/glethuillier/deLLMiter/utils"

	"go.uber.org/zap"
)

const defaultAPIURL = "http://localhost:1234"

func main() {
	modelName := flag.String("model", "", "The name of the model to use (required).")
	apiURL := flag.String("apiURL", defaultAPIURL, "The API URL to use for querying (optional).")
	flag.Parse()

	if *modelName == "" {
		fmt.Println("Error: model name is required.")
		flag.Usage()
		return
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			fmt.Printf("Failed to sync logger: %v\n", syncErr)
		}
	}()

	gen, err := generator.NewGenerator(logger)
	if err != nil {
		logger.Fatal("Failed to create generator", zap.Error(err))
	}

	cl, err := client.NewClient(*apiURL, *modelName)
	if err != nil {
		logger.Fatal("Failed to create client", zap.Error(err))
	}

	analyzer := analyzer.NewAnalyzer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Println("deLLMiter started.")

	go func() {
		<-stop
		logger.Info("deLLMiter shutting down.")
		os.Exit(0)
	}()

	for {
		candidate := gen.GenerateCandidate(2, 4)

		response, queryErr := cl.Query(*modelName, candidate.Message)
		if queryErr != nil {
			logger.Error("Failed to query the model", zap.Error(queryErr))
			continue
		}

		areIdentical, mismatchedDelimiters := analyzer.AreIdentical(candidate, response)
		if !areIdentical {
			fmt.Printf("Send:	 %s\n", candidate.Message)
			fmt.Printf("Received: %s\n\n", response)

			if saveErr := utils.SaveResult(*modelName, candidate, response); saveErr != nil {
				logger.Error("Failed to save the discrepancies", zap.Error(saveErr))
			}

			if len(mismatchedDelimiters) > 0 {
				if saveDelimErr := utils.SaveDelimiters(*modelName, mismatchedDelimiters); saveDelimErr != nil {
					logger.Error("Failed to save LLM delimiters", zap.Error(saveDelimErr))
				}
			}
		}
	}
}
