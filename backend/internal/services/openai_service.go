package services

import (
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	Client *openai.Client
}

func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		Client: openai.NewClient(apiKey),
	}
}
