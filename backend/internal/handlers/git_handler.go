package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"rag-chat-system/internal/services"
)

type GitHandler struct {
	gitSvc *services.GitService
}

func NewGitHandler(gitSvc *services.GitService) *GitHandler {
	return &GitHandler{gitSvc: gitSvc}
}

// SaveGitConfig saves or updates the git configuration for a project.
// PUT /projects/:id/git
func (h *GitHandler) SaveGitConfig(c echo.Context) error {
	projectID := c.Param("id")

	var req struct {
		GitURL    string `json:"git_url"`
		GitBranch string `json:"git_branch"`
		Token     string `json:"token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.GitURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "git_url is required"})
	}

	if err := h.gitSvc.ConfigureGit(c.Request().Context(), projectID, req.GitURL, req.GitBranch, req.Token); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "saved"})
}

// GetGitConfig returns the git configuration for a project.
// GET /projects/:id/git
func (h *GitHandler) GetGitConfig(c echo.Context) error {
	projectID := c.Param("id")

	config, err := h.gitSvc.GetConfig(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if config == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{})
	}

	return c.JSON(http.StatusOK, config)
}

// SyncGit triggers a git sync for a project (runs in background).
// POST /projects/:id/git/sync
func (h *GitHandler) SyncGit(c echo.Context) error {
	projectID := c.Param("id")

	if err := h.gitSvc.SyncAsync(projectID); err != nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"status": "syncing"})
}

// RemoveGitConfig removes the git configuration from a project.
// DELETE /projects/:id/git
func (h *GitHandler) RemoveGitConfig(c echo.Context) error {
	projectID := c.Param("id")

	if err := h.gitSvc.RemoveConfig(c.Request().Context(), projectID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "removed"})
}
