package services

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"

	"rag-chat-system/internal/models"
	"rag-chat-system/internal/rag"
	"rag-chat-system/internal/repositories"
)

type ChatService struct {
	chatRepo    *repositories.ChatRepo
	messageRepo *repositories.MessageRepo
	ragService  *RAGService
	openaiSvc   *OpenAIService
}

func NewChatService(
	chatRepo *repositories.ChatRepo,
	messageRepo *repositories.MessageRepo,
	ragService *RAGService,
	openaiSvc *OpenAIService,
) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		ragService:  ragService,
		openaiSvc:   openaiSvc,
	}
}

func (s *ChatService) CreateChat(ctx context.Context, title string, projectIDs []string) (*models.Chat, error) {
	if title == "" {
		title = "New Chat"
	}
	if projectIDs == nil {
		projectIDs = []string{}
	}
	chat := &models.Chat{
		ID:         uuid.New().String(),
		Title:      title,
		ProjectIDs: projectIDs,
	}
	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatService) ListChats(ctx context.Context) ([]models.Chat, error) {
	chats, err := s.chatRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	if chats == nil {
		chats = []models.Chat{}
	}
	return chats, nil
}

func (s *ChatService) GetMessages(ctx context.Context, chatID string) ([]models.Message, error) {
	msgs, err := s.messageRepo.ListByChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if msgs == nil {
		msgs = []models.Message{}
	}
	return msgs, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID string) error {
	// Delete messages first (foreign key)
	if err := s.messageRepo.DeleteByChatID(ctx, chatID); err != nil {
		return fmt.Errorf("delete messages: %w", err)
	}
	if err := s.chatRepo.Delete(ctx, chatID); err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}
	return nil
}

func (s *ChatService) UpdateProjectIDs(ctx context.Context, chatID string, projectIDs []string) error {
	return s.chatRepo.UpdateProjectIDs(ctx, chatID, projectIDs)
}

func (s *ChatService) SendMessage(ctx context.Context, chatID, userMessage string, projectIDs []string) (<-chan string, <-chan error) {
	tokenCh := make(chan string, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		// Save user message
		userMsg := &models.Message{
			ID:      uuid.New().String(),
			ChatID:  chatID,
			Role:    "user",
			Content: userMessage,
		}
		if err := s.messageRepo.Create(ctx, userMsg); err != nil {
			errCh <- fmt.Errorf("save user message: %w", err)
			return
		}

		// RAG search
		chunks, err := s.ragService.SearchRelevantChunks(ctx, userMessage, projectIDs)
		if err != nil {
			// Non-fatal: proceed without context if search fails
			chunks = nil
		}

		var systemPrompt string
		if len(chunks) > 0 {
			ragContext := s.ragService.BuildContext(chunks)
			systemPrompt = fmt.Sprintf(rag.SystemPromptWithContext, ragContext)
		} else {
			systemPrompt = rag.SystemPromptNoContext
		}

		// Build messages array: system + history + current user message
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
		}

		// Fetch recent conversation history (up to 20 messages)
		history, err := s.messageRepo.ListByChatID(ctx, chatID)
		if err == nil && len(history) > 0 {
			// Exclude the last message (the one we just saved above)
			if len(history) > 0 && history[len(history)-1].ID == userMsg.ID {
				history = history[:len(history)-1]
			}
			// Keep only the last 20 messages
			if len(history) > 20 {
				history = history[len(history)-20:]
			}
			for _, msg := range history {
				role := openai.ChatMessageRoleUser
				if msg.Role == "assistant" {
					role = openai.ChatMessageRoleAssistant
				}
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    role,
					Content: msg.Content,
				})
			}
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		})

		// Stream from OpenAI
		stream, err := s.openaiSvc.Client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini,
			Messages: messages,
			Stream:   true,
		})
		if err != nil {
			errCh <- fmt.Errorf("openai stream: %w", err)
			return
		}
		defer stream.Close()

		var fullResponse strings.Builder

		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errCh <- fmt.Errorf("stream recv: %w", err)
				return
			}

			token := response.Choices[0].Delta.Content
			if token != "" {
				fullResponse.WriteString(token)
				tokenCh <- token
			}
		}

		// Save assistant message
		assistantMsg := &models.Message{
			ID:      uuid.New().String(),
			ChatID:  chatID,
			Role:    "assistant",
			Content: fullResponse.String(),
		}
		if err := s.messageRepo.Create(ctx, assistantMsg); err != nil {
			errCh <- fmt.Errorf("save assistant message: %w", err)
			return
		}

		// Update chat title from first message (rune-safe truncation)
		runes := []rune(userMessage)
		if len(runes) > 50 {
			_ = s.chatRepo.UpdateTitle(ctx, chatID, string(runes[:50])+"...")
		} else {
			_ = s.chatRepo.UpdateTitle(ctx, chatID, userMessage)
		}
	}()

	return tokenCh, errCh
}
