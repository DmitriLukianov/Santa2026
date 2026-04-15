package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"secret-santa-backend/internal/controller/http/v1/response"
	"secret-santa-backend/internal/definitions"

	"github.com/google/uuid"
)

const maxUploadSize = 5 << 20 // 5 MB

type UploadHandler struct {
	baseURL   string
	uploadDir string
}

func NewUploadHandler(baseURL, uploadDir string) *UploadHandler {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create upload directory %q: %v", uploadDir, err))
	}
	return &UploadHandler{baseURL: baseURL, uploadDir: uploadDir}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}
	defer file.Close()

	// Определяем MIME-тип по содержимому файла
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}
	contentType := http.DetectContentType(buf)
	if !strings.HasPrefix(contentType, "image/") {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	// Возвращаемся в начало файла после чтения для определения типа
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	dst := filepath.Join(h.uploadDir, filename)

	out, err := os.Create(dst)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	url := fmt.Sprintf("/static/uploads/%s", filename)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
