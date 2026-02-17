package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"rag-chat-system/internal/services"
)

type ChatHandler struct {
	chatSvc *services.ChatService
}

func NewChatHandler(chatSvc *services.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

func (h *ChatHandler) CreateChat(c echo.Context) error {
	var req struct {
		Title string `json:"title"`
	}
	_ = c.Bind(&req)

	chat, err := h.chatSvc.CreateChat(c.Request().Context(), req.Title)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, chat)
}

func (h *ChatHandler) ListChats(c echo.Context) error {
	chats, err := h.chatSvc.ListChats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, chats)
}

func (h *ChatHandler) GetMessages(c echo.Context) error {
	chatID := c.Param("id")
	msgs, err := h.chatSvc.GetMessages(c.Request().Context(), chatID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, msgs)
}

func (h *ChatHandler) SendMessage(c echo.Context) error {
	chatID := c.Param("id")

	var req struct {
		Message    string   `json:"message"`
		ProjectIDs []string `json:"project_ids"`
	}
	if err := c.Bind(&req); err != nil || req.Message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "message is required"})
	}

	ctx := c.Request().Context()

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	tokenCh, errCh := h.chatSvc.SendMessage(ctx, chatID, req.Message, req.ProjectIDs)

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
	}

	for {
		select {
		case token, ok := <-tokenCh:
			if !ok {
				// Check for errors
				select {
				case err := <-errCh:
					if err != nil {
						fmt.Fprintf(c.Response().Writer, "data: [ERROR] %s\n\n", err.Error())
						flusher.Flush()
					}
				default:
				}
				fmt.Fprintf(c.Response().Writer, "data: [DONE]\n\n")
				flusher.Flush()
				return nil
			}
			fmt.Fprintf(c.Response().Writer, "data: %s\n\n", token)
			flusher.Flush()

		case err := <-errCh:
			if err != nil {
				fmt.Fprintf(c.Response().Writer, "data: [ERROR] %s\n\n", err.Error())
				flusher.Flush()
				return nil
			}
		}
	}
}
