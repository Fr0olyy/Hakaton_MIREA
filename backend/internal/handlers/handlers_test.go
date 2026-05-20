package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hakaton/backend/internal/repositories"
	"hakaton/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestBadUUIDReturnsBadRequest(t *testing.T) {
	router := chi.NewRouter()
	New(services.New(nil, nil, nil)).Register(router)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects/not-a-uuid/dashboard", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestUploadRequiresDatasetFile(t *testing.T) {
	router := chi.NewRouter()
	New(services.New(nil, nil, nil)).Register(router)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("note", "no dataset here"); err != nil {
		t.Fatalf("write multipart field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects/"+uuid.NewString()+"/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestCreateProjectRejectsUnsupportedLevelOneType(t *testing.T) {
	router := chi.NewRouter()
	New(services.New(nil, nil, nil)).Register(router)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"Demo","modality":"text","task_type":"classification"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestWriteResultMapsNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeResult(recorder, nil, repositories.ErrNotFound)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
