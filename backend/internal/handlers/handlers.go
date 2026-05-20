package handlers

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"os"

	"hakaton/backend/internal/repositories"
	"hakaton/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *services.Service
}

func New(service *services.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/health", h.health)

	r.Post("/api/projects", h.createProject)
	r.Get("/api/projects", h.listProjects)
	r.Get("/api/projects/{id}", h.getProject)
	r.Route("/api/projects/{id}", func(r chi.Router) {
		r.Post("/upload", h.upload)
		r.Post("/analyze", h.analyze)
		r.Get("/dashboard", h.dashboard)
		r.Get("/probabilistic-analysis", h.probabilisticAnalysis)
		r.Get("/review-queue", h.reviewQueue)
		r.Get("/objects/{objectId}", h.object)
		r.Get("/objects/{objectId}/file", h.objectFile)
		r.Get("/recommendations", h.recommendations)
		r.Get("/roadmap", h.roadmap)
		r.Post("/export", h.createExport)
		r.Get("/exports/{exportId}/download", h.downloadExport)
		r.Post("/agent/summary", h.agentSummary)
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Health(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "error", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var req services.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	project, err := h.service.CreateProject(r.Context(), req)
	if err != nil {
		if services.IsBadRequest(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	project, err := h.service.GetProject(r.Context(), projectID)
	writeResult(w, project, err)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(256 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "multipart form is required")
		return
	}
	datasetHeader, err := fileHeader(r, "dataset")
	if err != nil {
		datasetHeader, err = fileHeader(r, "dataset.csv")
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "dataset file field is required")
		return
	}
	imagesHeader, _ := fileHeader(r, "images")
	if imagesHeader == nil {
		imagesHeader, _ = fileHeader(r, "images.zip")
	}
	result, err := h.service.Upload(r.Context(), projectID, datasetHeader, imagesHeader)
	writeResult(w, result, err)
}

func (h *Handler) analyze(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	job, err := h.service.Analyze(r.Context(), projectID)
	writeResult(w, job, err)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.Dashboard(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) probabilisticAnalysis(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.ProbabilisticAnalysis(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) reviewQueue(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.ReviewQueue(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) object(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	objectID, ok := parseUUIDParam(w, r, "objectId")
	if !ok {
		return
	}
	result, err := h.service.Object(r.Context(), projectID, objectID)
	writeResult(w, result, err)
}

func (h *Handler) objectFile(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	objectID, ok := parseUUIDParam(w, r, "objectId")
	if !ok {
		return
	}
	result, err := h.service.ObjectFile(r.Context(), projectID, objectID)
	if err != nil {
		writeResult(w, nil, err)
		return
	}
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", `inline; filename="`+result.FileName+`"`)
	http.ServeFile(w, r, result.Path)
}

func (h *Handler) recommendations(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.Recommendations(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) roadmap(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.Roadmap(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) createExport(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.Export(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) downloadExport(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	exportID, ok := parseUUIDParam(w, r, "exportId")
	if !ok {
		return
	}
	export, err := h.service.GetExport(r.Context(), projectID, exportID)
	if err != nil {
		writeResult(w, nil, err)
		return
	}
	if _, err := os.Stat(export.FilePath); err != nil {
		writeError(w, http.StatusNotFound, "export file not found")
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="export.zip"`)
	http.ServeFile(w, r, export.FilePath)
}

func (h *Handler) agentSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func fileHeader(r *http.Request, name string) (*multipart.FileHeader, error) {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, http.ErrMissingFile
	}
	files := r.MultipartForm.File[name]
	if len(files) == 0 {
		return nil, http.ErrMissingFile
	}
	return files[0], nil
}

func parseUUIDParam(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	value := chi.URLParam(r, name)
	id, err := uuid.Parse(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}

func writeResult(w http.ResponseWriter, value any, err error) {
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if services.IsBadRequest(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
