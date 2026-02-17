package handlers

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"rag-chat-system/internal/services"
)

type FileHandler struct {
	fileSvc *services.FileService
}

func NewFileHandler(fileSvc *services.FileService) *FileHandler {
	return &FileHandler{fileSvc: fileSvc}
}

func (h *FileHandler) UploadFile(c echo.Context) error {
	projectID := c.Param("id")
	ctx := c.Request().Context()

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open file"})
	}
	defer src.Close()

	f, err := h.fileSvc.UploadFile(ctx, projectID, nil, file.Filename, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, f)
}

func (h *FileHandler) UploadFolder(c echo.Context) error {
	projectID := c.Param("id")
	ctx := c.Request().Context()

	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
	}

	files := form.File["files"]
	paths := form.Value["paths"]

	if len(files) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no files provided"})
	}

	// Track created directories: relPath -> dirID
	dirCache := make(map[string]string)

	for i, fh := range files {
		var relPath string
		if i < len(paths) {
			relPath = paths[i]
		} else {
			relPath = fh.Filename
		}

		// Ensure parent directories exist
		dir := filepath.Dir(relPath)
		var parentID *string

		if dir != "." && dir != "" {
			parentID, err = h.ensureDirChain(ctx, projectID, dir, dirCache)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "create dirs: " + err.Error()})
			}
		}

		src, err := fh.Open()
		if err != nil {
			continue
		}

		_, err = h.fileSvc.UploadFile(ctx, projectID, parentID, filepath.Base(relPath), src)
		src.Close()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "upload: " + err.Error()})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "folder uploaded"})
}

func (h *FileHandler) ensureDirChain(ctx context.Context, projectID, dirPath string, cache map[string]string) (*string, error) {
	if id, ok := cache[dirPath]; ok {
		return &id, nil
	}

	parts := strings.Split(filepath.ToSlash(dirPath), "/")
	var parentID *string
	accumulated := ""

	for _, part := range parts {
		if part == "" {
			continue
		}
		if accumulated == "" {
			accumulated = part
		} else {
			accumulated = accumulated + "/" + part
		}

		if id, ok := cache[accumulated]; ok {
			parentID = &id
			continue
		}

		dir, err := h.fileSvc.EnsureDir(ctx, projectID, parentID, part, accumulated)
		if err != nil {
			return nil, err
		}

		cache[accumulated] = dir.ID
		parentID = &dir.ID
	}

	return parentID, nil
}

func (h *FileHandler) ListFiles(c echo.Context) error {
	projectID := c.Param("id")
	files, err := h.fileSvc.ListByProject(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, files)
}

func (h *FileHandler) DeleteFile(c echo.Context) error {
	fileID := c.Param("id")
	if err := h.fileSvc.DeleteFile(c.Request().Context(), fileID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}
