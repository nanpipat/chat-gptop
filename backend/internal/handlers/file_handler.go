package handlers

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
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

	// Get multipart boundary from Content-Type
	contentType := c.Request().Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "expected multipart form"})
	}
	boundary := params["boundary"]
	if boundary == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no boundary in content-type"})
	}

	// Stream parts one at a time — avoids Go 1.24 multipart size limits
	mr := multipart.NewReader(c.Request().Body, boundary)

	type fileEntry struct {
		filename string
		data     []byte
		relPath  string
	}
	var entries []fileEntry
	var paths []string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "read part: " + err.Error()})
		}

		fieldName := part.FormName()
		switch fieldName {
		case "paths":
			val, err := io.ReadAll(part)
			if err != nil {
				part.Close()
				continue
			}
			paths = append(paths, string(val))
		case "files":
			data, err := io.ReadAll(part)
			if err != nil {
				part.Close()
				continue
			}
			entries = append(entries, fileEntry{
				filename: part.FileName(),
				data:     data,
			})
		}
		part.Close()
	}

	if len(entries) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no files provided"})
	}

	// Assign relative paths to entries
	for i := range entries {
		if i < len(paths) {
			entries[i].relPath = paths[i]
		} else {
			entries[i].relPath = entries[i].filename
		}
	}

	// Track created directories: relPath -> dirID
	dirCache := make(map[string]string)

	for _, entry := range entries {
		dir := filepath.Dir(entry.relPath)
		var parentID *string

		if dir != "." && dir != "" {
			parentID, err = h.ensureDirChain(ctx, projectID, dir, dirCache)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "create dirs: " + err.Error()})
			}
		}

		reader := bytes.NewReader(entry.data)
		_, err = h.fileSvc.UploadFile(ctx, projectID, parentID, filepath.Base(entry.relPath), reader)
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
