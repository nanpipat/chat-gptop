package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"rag-chat-system/internal/models"
	"rag-chat-system/internal/repositories"
)

type ProjectHandler struct {
	repo *repositories.ProjectRepo
}

func NewProjectHandler(repo *repositories.ProjectRepo) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

func (h *ProjectHandler) Create(c echo.Context) error {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	p := &models.Project{
		ID:   uuid.New().String(),
		Name: req.Name,
	}
	if err := h.repo.Create(c.Request().Context(), p); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, p)
}

func (h *ProjectHandler) List(c echo.Context) error {
	projects, err := h.repo.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if projects == nil {
		projects = []models.Project{}
	}
	return c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}
