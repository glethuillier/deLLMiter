package client

import (
	"fmt"
	"strings"
)

// promptConstraints specifies a constant string instruction for response handling without modifications or additions
const (
	systemConstraints = `You will be given a message and you must respond it back. Do not add anything else, do not modify the original text, do not comment it.`
	userConstraints   = `Just repeat back the following message (do not add anything else, do not modify the original text, do not comment it)`
)

type Prompt struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

func getSystemMessage(role string) Message {
	return Message{
		Role:    role,
		Content: systemConstraints,
	}
}

func getUserMessage(role, content string) Message {
	return Message{
		Role:    role,
		Content: fmt.Sprintf("%s: %s", userConstraints, content),
	}
}

func getPrompt(modelName, content string) Prompt {
	modelName = strings.TrimSpace(strings.ToLower(modelName))

	var messages []Message
	switch {
	case strings.Contains(modelName, "mistral"):
		messages = append(messages, getSystemMessage("assistant"))
		messages = append(messages, getUserMessage("user", content))

	case strings.Contains(modelName, "hermes-3"):
		messages = append(messages, getUserMessage("user", content))
		
	default:
		messages = append(messages, getSystemMessage("system"))
		messages = append(messages, getUserMessage("user", content))
	}

	return Prompt{
		Model:       modelName,
		Messages:    messages,
		Temperature: 0.8,
		Stream:      false,
	}
}
