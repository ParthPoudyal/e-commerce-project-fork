package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/auction-studio/backend/internal/config"
	"github.com/google/uuid"
)

type UploadHandler struct {
	cfg *config.Config
}

func NewUploadHandler(cfg *config.Config) *UploadHandler {
	return &UploadHandler{cfg: cfg}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// 10MB default limit
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadSize)
	if err := r.ParseMultipartForm(h.cfg.MaxUploadSize); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "file too large (max 10MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("image")
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "no file provided in form field 'file' or 'image'")
			return
		}
	}
	defer file.Close()

	// Validate content type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "could not read file header")
		return
	}
	contentType := http.DetectContentType(buff)
	validTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}

	ext, ok := validTypes[contentType]
	if !ok {
		// Fallback to checking extension from filename
		fExt := strings.ToLower(filepath.Ext(header.Filename))
		if fExt == ".jpg" || fExt == ".jpeg" || fExt == ".png" || fExt == ".webp" || fExt == ".gif" {
			ext = fExt
		} else {
			ErrorJSON(w, http.StatusBadRequest, "unsupported file format. Only JPEG, PNG, WEBP, and GIF are allowed")
			return
		}
	}

	// Rewind file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0755); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dstPath := filepath.Join(h.cfg.UploadDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to write file")
		return
	}

	url := fmt.Sprintf("/uploads/%s", filename)
	JSON(w, http.StatusCreated, map[string]string{"url": url})
}
