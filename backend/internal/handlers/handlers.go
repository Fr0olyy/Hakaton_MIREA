package handlers

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"hakaton/backend/internal/auth"
	"hakaton/backend/internal/models"
	"hakaton/backend/internal/repositories"
	"hakaton/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service   *services.Service
	jwtSecret string
}

func New(service *services.Service, jwtSecret string) *Handler {
	return &Handler{service: service, jwtSecret: jwtSecret}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/health", h.health)

	// Public auth endpoints
	r.Post("/api/auth/register", h.register)
	r.Post("/api/auth/login", h.login)

	// Public demo/read-only endpoints.
	// Нужно для frontend demo-flow: список проектов и получение проекта
	// не должны падать до загрузки датасета/анализа.
	r.Get("/api/demo-datasets", h.demoDatasets)
	r.Post("/api/demo-datasets/{modality}/project", h.createDemoProject)
	r.Get("/api/projects", h.listProjects)
	r.Get("/api/projects/{id}", h.getProject)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(h.jwtSecret))

		r.Get("/api/me", h.me)

		// Teams
		r.Post("/api/teams", h.createTeam)
		r.Get("/api/teams", h.listTeams)
		r.Get("/api/teams/{teamId}", h.getTeam)
		r.Post("/api/teams/{teamId}/members", h.addTeamMember)
		r.Delete("/api/teams/{teamId}/members/{userId}", h.removeTeamMember)

		// Projects
		r.Post("/api/projects", h.createProject)
		r.Get("/api/jobs/{jobId}", h.getJob)

		r.Route("/api/projects/{id}", func(r chi.Router) {
			// Root project route with trailing slash support
			r.Get("/", h.getProject)

			// Project member management
			r.Get("/members", h.listProjectMembers)
			r.Post("/members", h.addProjectMember)
			r.Delete("/members/{userId}", h.removeProjectMember)

			// Workflow
			r.Post("/upload", h.upload)
			r.Post("/analyze", h.analyze)
			r.Get("/dashboard", h.dashboard)
			r.Get("/probabilistic-analysis", h.probabilisticAnalysis)
			r.Get("/review-queue", h.reviewQueue)
			r.Get("/jobs", h.listJobs)
			r.Get("/objects", h.listObjects)
			r.Get("/objects/{objectId}", h.object)
			r.Get("/objects/{objectId}/file", h.objectFile)
			r.Post("/objects/{objectId}/action", h.objectAction)
			r.Post("/objects/{objectId}/comments", h.addObjectComment)
			r.Get("/objects/{objectId}/comments", h.listObjectComments)
			r.Post("/review-queue/assign", h.assignReview)
			r.Get("/review-queue/assigned-to-me", h.assignedToMe)

			// Tasks
			r.Get("/collection-tasks", h.listCollectionTasks)
			r.Post("/collection-tasks", h.createCollectionTask)
			r.Patch("/collection-tasks/{taskId}", h.updateCollectionTask)
			r.Get("/synthetic-tasks", h.listSyntheticTasks)
			r.Post("/synthetic-tasks", h.createSyntheticTask)
			r.Patch("/synthetic-tasks/{taskId}", h.updateSyntheticTask)
			r.Get("/class-action-plan", h.classActionPlan)

			// Outputs
			r.Get("/recommendations", h.recommendations)
			r.Get("/roadmap", h.roadmap)
			r.Post("/export", h.createExport)
			r.Get("/exports/{exportId}/download", h.downloadExport)

			// Agent Level 1
			r.Post("/agent/summary", h.agentSummary)
			r.Post("/agent/dataset-summary", h.agentDatasetSummary)
			r.Post("/agent/recommendations-summary", h.agentRecommendationsSummary)
			r.Post("/agent/roadmap-summary", h.agentRoadmapSummary)

			// Agent Level 2: Role-aware
			r.Post("/agent/admin-summary", h.agentAdminSummary)
			r.Post("/agent/ml-engineer-summary", h.agentMLEngineerSummary)
			r.Post("/agent/annotator-summary", h.agentAnnotatorSummary)
			r.Post("/agent/expert-summary", h.agentExpertSummary)
			r.Post("/agent/analyst-summary", h.agentAnalystSummary)

			// Agent Level 2: Object Explanation
			r.Post("/objects/{objectId}/agent/explain", h.agentObjectExplanation)

			// Agent Level 2: Collection / Synthetic Plans
			r.Post("/agent/collection-plan", h.agentCollectionPlan)
			r.Post("/agent/synthetic-plan", h.agentSyntheticPlan)

			// Agent Level 2: Task Generation
			r.Post("/agent/generate-collection-tasks", h.agentGenerateCollectionTasks)
			r.Post("/agent/generate-synthetic-tasks", h.agentGenerateSyntheticTasks)
			r.Post("/agent/generate-annotator-brief", h.agentGenerateAnnotatorBrief)
			r.Post("/agent/generate-expert-brief", h.agentGenerateExpertBrief)
		})
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Health(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "error", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) demoDatasets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.DemoDatasets())
}

func (h *Handler) createDemoProject(w http.ResponseWriter, r *http.Request) {
	claims := auth.UserFromContext(r.Context())
	var creatorID uuid.UUID
	if claims != nil {
		creatorID = claims.UserID
	}
	result, err := h.service.CreateDemoProject(r.Context(), chi.URLParam(r, "modality"), creatorID)
	writeResult(w, result, err)
}

// ---- Auth ----

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req services.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.Register(r.Context(), req)
	writeResult(w, result, err)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req services.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.Login(r.Context(), req)
	writeResult(w, result, err)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	user, err := h.service.Me(r.Context(), claims.UserID)
	writeResult(w, user, err)
}

// ---- Teams ----

func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req services.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.CreateTeam(r.Context(), req, claims.UserID)
	writeResult(w, result, err)
}

func (h *Handler) listTeams(w http.ResponseWriter, r *http.Request) {
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	result, err := h.service.ListTeams(r.Context(), claims.UserID)
	writeResult(w, result, err)
}

func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	teamID, ok := parseUUIDParam(w, r, "teamId")
	if !ok {
		return
	}
	result, err := h.service.GetTeam(r.Context(), teamID, claims.UserID)
	writeResult(w, result, err)
}

func (h *Handler) addTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID, ok := parseUUIDParam(w, r, "teamId")
	if !ok {
		return
	}
	var req struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.AddTeamMember(r.Context(), teamID, req.UserID)
	writeResult(w, result, err)
}

func (h *Handler) removeTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID, ok := parseUUIDParam(w, r, "teamId")
	if !ok {
		return
	}
	userID, ok := parseUUIDParam(w, r, "userId")
	if !ok {
		return
	}
	err := h.service.RemoveTeamMember(r.Context(), teamID, userID)
	writeResult(w, nil, err)
}

// ---- Project Members ----

func (h *Handler) addProjectMember(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	var req struct {
		UserID uuid.UUID                  `json:"user_id"`
		Role   models.ProjectMemberRole   `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	err := h.service.UpdateProjectRole(r.Context(), projectID, req.UserID, req.Role)
	writeResult(w, nil, err)
}

func (h *Handler) listProjectMembers(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	members, err := h.service.ListProjectMembers(r.Context(), projectID)
	writeResult(w, members, err)
}

func (h *Handler) removeProjectMember(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	userID, ok := parseUUIDParam(w, r, "userId")
	if !ok {
		return
	}
	err := h.service.RemoveProjectMember(r.Context(), projectID, userID)
	writeResult(w, nil, err)
}

// ---- Projects ----

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var req services.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	claims := auth.UserFromContext(r.Context())
	var creatorID uuid.UUID
	if claims != nil {
		creatorID = claims.UserID
	}
	project, err := h.service.CreateProject(r.Context(), req, creatorID)
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
	if err == nil {
		writeJSON(w, http.StatusOK, project)
		return
	}

	// Demo/dev fallback:
	// если прямой GetProject почему-то не нашёл проект,
	// но ListProjects его видит, возвращаем проект из списка.
	projects, listErr := h.service.ListProjects(r.Context())
	if listErr == nil {
		for _, item := range projects {
			if item.ID == projectID {
				writeJSON(w, http.StatusOK, item)
				return
			}
		}
	}

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
	result, err := h.service.GetObjectDetail(r.Context(), projectID, objectID)
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

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	jobID, ok := parseUUIDParam(w, r, "jobId")
	if !ok {
		return
	}
	result, err := h.service.GetJob(r.Context(), jobID)
	writeResult(w, result, err)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.ListJobs(r.Context(), projectID)
	writeResult(w, result, err)
}

// ---- Object Workflow ----

func (h *Handler) listObjects(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	q := r.URL.Query()
	page := queryInt(q, "page", 1)
	perPage := queryInt(q, "per_page", 20)
	scoreMin := queryFloat(q, "score_min", 0)
	scoreMax := queryFloat(q, "score_max", 0)

	result, err := h.service.ListObjects(r.Context(), projectID, page, perPage,
		q.Get("status"), q.Get("label"), q.Get("reason"), q.Get("recommendation"),
		q.Get("sort_by"), q.Get("order"), q.Get("search"),
		scoreMin, scoreMax)
	writeResult(w, result, err)
}

func (h *Handler) objectAction(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	objectID, ok := parseUUIDParam(w, r, "objectId")
	if !ok {
		return
	}
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req struct {
		Action   string `json:"action"`
		OldValue string `json:"old_value"`
		NewValue string `json:"new_value"`
		Comment  string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.PerformObjectAction(r.Context(), projectID, objectID, claims.UserID, req.Action, req.OldValue, req.NewValue, req.Comment)
	writeResult(w, result, err)
}

func (h *Handler) addObjectComment(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	objectID, ok := parseUUIDParam(w, r, "objectId")
	if !ok {
		return
	}
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.AddObjectComment(r.Context(), projectID, objectID, claims.UserID, req.Text)
	writeResult(w, result, err)
}

func (h *Handler) listObjectComments(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	objectID, ok := parseUUIDParam(w, r, "objectId")
	if !ok {
		return
	}
	result, err := h.service.ListObjectComments(r.Context(), projectID, objectID)
	writeResult(w, result, err)
}

func (h *Handler) assignReview(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	var req struct {
		ObjectID uuid.UUID `json:"object_id"`
		UserID   uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.AssignReview(r.Context(), projectID, req.ObjectID, req.UserID)
	writeResult(w, result, err)
}

func (h *Handler) assignedToMe(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	result, err := h.service.AssignedToMe(r.Context(), projectID, claims.UserID)
	writeResult(w, result, err)
}

// ---- Tasks ----

func (h *Handler) createCollectionTask(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req services.CreateCollectionTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.CreateCollectionTask(r.Context(), projectID, claims.UserID, req)
	writeResult(w, result, err)
}

func (h *Handler) listCollectionTasks(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.ListCollectionTasks(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) updateCollectionTask(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	taskID, ok := parseUUIDParam(w, r, "taskId")
	if !ok {
		return
	}
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.UpdateCollectionTask(r.Context(), projectID, taskID, updates)
	writeResult(w, result, err)
}

func (h *Handler) createSyntheticTask(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req services.CreateSyntheticTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.CreateSyntheticTask(r.Context(), projectID, claims.UserID, req)
	writeResult(w, result, err)
}

func (h *Handler) listSyntheticTasks(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.ListSyntheticTasks(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) updateSyntheticTask(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	taskID, ok := parseUUIDParam(w, r, "taskId")
	if !ok {
		return
	}
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.service.UpdateSyntheticTask(r.Context(), projectID, taskID, updates)
	writeResult(w, result, err)
}

func (h *Handler) classActionPlan(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.ClassActionPlan(r.Context(), projectID)
	writeResult(w, result, err)
}

// ---- Agent Level 1 ----

func (h *Handler) agentSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentDatasetSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentDatasetSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentRecommendationsSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentRecommendationsSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentRoadmapSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentRoadmapSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

// ---- Agent Level 2: Role-aware ----

func (h *Handler) agentAdminSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentAdminSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentMLEngineerSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentMLEngineerSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentAnnotatorSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentAnnotatorSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentExpertSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentExpertSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentAnalystSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentAnalystSummary(r.Context(), projectID)
	writeResult(w, result, err)
}

// ---- Agent Level 2: Object Explanation ----

func (h *Handler) agentObjectExplanation(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	objectID, ok := parseUUIDParam(w, r, "objectId")
	if !ok {
		return
	}
	result, err := h.service.AgentObjectExplanation(r.Context(), projectID, objectID)
	writeResult(w, result, err)
}

// ---- Agent Level 2: Collection / Synthetic Plans ----

func (h *Handler) agentCollectionPlan(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentCollectionPlan(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentSyntheticPlan(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentSyntheticPlan(r.Context(), projectID)
	writeResult(w, result, err)
}

// ---- Agent Level 2: Task Generation ----

func (h *Handler) agentGenerateCollectionTasks(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentGenerateCollectionTasks(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}

func (h *Handler) agentGenerateSyntheticTasks(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentGenerateSyntheticTasks(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}

func (h *Handler) agentGenerateAnnotatorBrief(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentGenerateAnnotatorBrief(r.Context(), projectID)
	writeResult(w, result, err)
}

func (h *Handler) agentGenerateExpertBrief(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	result, err := h.service.AgentGenerateExpertBrief(r.Context(), projectID)
	writeResult(w, result, err)
}

func queryInt(q map[string][]string, key string, fallback int) int {
	values, ok := q[key]
	if !ok || len(values) == 0 {
		return fallback
	}
	parsed, err := strconv.Atoi(values[0])
	if err != nil {
		return fallback
	}
	return parsed
}

func queryFloat(q map[string][]string, key string, fallback float64) float64 {
	values, ok := q[key]
	if !ok || len(values) == 0 {
		return fallback
	}
	parsed, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return fallback
	}
	return parsed
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
