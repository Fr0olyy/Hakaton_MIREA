package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dataforge/agent-service/internal/agent"
)

type Handler struct {
	svc *agent.Service
}

func New(svc *agent.Service) *Handler {
	return &Handler{svc: svc}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---- Level 1 ----

type DatasetSummaryRequest struct {
	Readiness       float64  `json:"readiness"`
	TotalObjects    int      `json:"total_objects"`
	ReviewItems     int      `json:"review_items"`
	ImbalanceIndex  float64  `json:"imbalance_index"`
	AvgLabelError   float64  `json:"avg_label_error_probability"`
	MissingFiles    int      `json:"missing_files"`
	Recommendations []string `json:"recommendations"`
	Roadmap         []string `json:"roadmap"`
}

func (h *Handler) DatasetSummary(w http.ResponseWriter, r *http.Request) {
	var req DatasetSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateDatasetSummary(r.Context(),
		req.Readiness, req.TotalObjects, req.ReviewItems,
		req.ImbalanceIndex, req.AvgLabelError, req.MissingFiles,
		req.Recommendations, req.Roadmap)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type RecommendationsSummaryRequest struct {
	Recommendations []string `json:"recommendations"`
}

func (h *Handler) RecommendationsSummary(w http.ResponseWriter, r *http.Request) {
	var req RecommendationsSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateRecommendationsSummary(r.Context(), req.Recommendations)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type RoadmapSummaryRequest struct {
	Roadmap []string `json:"roadmap"`
}

func (h *Handler) RoadmapSummary(w http.ResponseWriter, r *http.Request) {
	var req RoadmapSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateRoadmapSummary(r.Context(), req.Roadmap)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- Level 2: Role-aware ----

type AdminSummaryRequest struct {
	Readiness       float64  `json:"readiness"`
	TotalObjects    int      `json:"total_objects"`
	ReviewItems     int      `json:"review_items"`
	TeamMembers     int      `json:"team_members"`
	ReviewedObjects int      `json:"reviewed_objects"`
	ReadyForExport  string   `json:"ready_for_export"`
	Recommendations []string `json:"recommendations"`
	Roadmap         []string `json:"roadmap"`
}

func (h *Handler) AdminSummary(w http.ResponseWriter, r *http.Request) {
	var req AdminSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateAdminSummary(r.Context(),
		req.Readiness, req.TotalObjects, req.ReviewItems,
		req.TeamMembers, req.ReviewedObjects, req.ReadyForExport,
		req.Recommendations, req.Roadmap)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type MLEngineerSummaryRequest struct {
	Readiness      float64  `json:"readiness"`
	TotalObjects   int      `json:"total_objects"`
	ImbalanceIndex float64  `json:"imbalance_index"`
	AvgLabelError  float64  `json:"avg_label_error_probability"`
	AvgEntropy     float64  `json:"avg_entropy"`
	ReviewItems    int      `json:"review_items"`
	Recommendations []string `json:"recommendations"`
	Roadmap        []string `json:"roadmap"`
}

func (h *Handler) MLEngineerSummary(w http.ResponseWriter, r *http.Request) {
	var req MLEngineerSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateMLEngineerSummary(r.Context(),
		req.Readiness, req.TotalObjects, req.ImbalanceIndex,
		req.AvgLabelError, req.AvgEntropy, req.ReviewItems,
		req.Recommendations, req.Roadmap)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type AnnotatorSummaryRequest struct {
	AssignedCount    int      `json:"assigned_count"`
	ReviewItems      int      `json:"review_items"`
	StatusDistribution string `json:"status_distribution"`
	TopReviewObjects []string `json:"top_review_objects"`
}

func (h *Handler) AnnotatorSummary(w http.ResponseWriter, r *http.Request) {
	var req AnnotatorSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateAnnotatorSummary(r.Context(),
		req.AssignedCount, req.ReviewItems, req.StatusDistribution, req.TopReviewObjects)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type ExpertSummaryRequest struct {
	SentToExpert       int    `json:"sent_to_expert"`
	PendingDecisions   int    `json:"pending_decisions"`
	LabelConflicts     int    `json:"label_conflicts"`
	ClassesNeedingReview string `json:"classes_needing_review"`
}

func (h *Handler) ExpertSummary(w http.ResponseWriter, r *http.Request) {
	var req ExpertSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateExpertSummary(r.Context(),
		req.SentToExpert, req.PendingDecisions, req.LabelConflicts, req.ClassesNeedingReview)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type AnalystSummaryRequest struct {
	Readiness           float64 `json:"readiness"`
	VersionComparison   string  `json:"version_comparison"`
	ObjectsCount        int     `json:"objects_count"`
	DuplicateCandidates int     `json:"duplicate_candidates"`
	LabelErrorCandidates int    `json:"label_error_candidates"`
	QualityIssues       int     `json:"quality_issues"`
	ImbalanceIndex      float64 `json:"imbalance_index"`
}

func (h *Handler) AnalystSummary(w http.ResponseWriter, r *http.Request) {
	var req AnalystSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateAnalystSummary(r.Context(),
		req.Readiness, req.VersionComparison, req.ObjectsCount,
		req.DuplicateCandidates, req.LabelErrorCandidates, req.QualityIssues, req.ImbalanceIndex)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- Level 2: Object Explanation ----

type ObjectExplanationRequest struct {
	FilePath         string   `json:"file_path"`
	CurrentLabel     string   `json:"current_label"`
	PredictedLabel   string   `json:"predicted_label"`
	Confidence       float64  `json:"confidence"`
	Status           string   `json:"status"`
	Entropy          float64  `json:"entropy"`
	Uncertainty      float64  `json:"uncertainty"`
	LabelErrorProb   float64  `json:"label_error_probability"`
	DuplicateScore   float64  `json:"duplicate_score"`
	QualityScore     float64  `json:"quality_score"`
	FinalScore       float64  `json:"final_score"`
	Recommendation   string   `json:"recommendation"`
	Reasons          []string `json:"reasons"`
	Actions          []string `json:"actions"`
	Comments         []string `json:"comments"`
}

func (h *Handler) ObjectExplanation(w http.ResponseWriter, r *http.Request) {
	var req ObjectExplanationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateObjectExplanation(r.Context(),
		req.FilePath, req.CurrentLabel, req.PredictedLabel, req.Confidence, req.Status,
		req.Entropy, req.Uncertainty, req.LabelErrorProb, req.DuplicateScore, req.QualityScore, req.FinalScore,
		req.Recommendation, req.Reasons, req.Actions, req.Comments)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- Level 2: Plans ----

type PlanRequest struct {
	ClassActionPlan []string `json:"class_action_plan"`
}

func (h *Handler) CollectionPlan(w http.ResponseWriter, r *http.Request) {
	var req PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateCollectionPlan(r.Context(), req.ClassActionPlan)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) SyntheticPlan(w http.ResponseWriter, r *http.Request) {
	var req PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateSyntheticPlan(r.Context(), req.ClassActionPlan)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- Level 2: Task Generation ----

type GenerateTasksRequest struct {
	ClassActionPlan []string `json:"class_action_plan"`
}

func (h *Handler) GenerateCollectionTasks(w http.ResponseWriter, r *http.Request) {
	var req GenerateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	result, err := h.svc.GenerateCollectionTaskDrafts(r.Context(), req.ClassActionPlan)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}

func (h *Handler) GenerateSyntheticTasks(w http.ResponseWriter, r *http.Request) {
	var req GenerateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	result, err := h.svc.GenerateSyntheticTaskDrafts(r.Context(), req.ClassActionPlan)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}

type AnnotatorBriefRequest struct {
	ReviewItems  int    `json:"review_items"`
	Classes      string `json:"classes"`
	CommonIssues string `json:"common_issues"`
	TopPriority  string `json:"top_priority"`
}

func (h *Handler) GenerateAnnotatorBrief(w http.ResponseWriter, r *http.Request) {
	var req AnnotatorBriefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateAnnotatorBrief(r.Context(), req.ReviewItems, req.Classes, req.CommonIssues, req.TopPriority)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type ExpertBriefRequest struct {
	ExpertObjects  int    `json:"expert_objects"`
	ConflictClasses string `json:"conflict_classes"`
	KeyDecisions   string `json:"key_decisions"`
}

func (h *Handler) GenerateExpertBrief(w http.ResponseWriter, r *http.Request) {
	var req ExpertBriefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	resp, err := h.svc.GenerateExpertBrief(r.Context(), req.ExpertObjects, req.ConflictClasses, req.KeyDecisions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)

	mux.HandleFunc("POST /api/agent/dataset-summary", h.DatasetSummary)
	mux.HandleFunc("POST /api/agent/recommendations-summary", h.RecommendationsSummary)
	mux.HandleFunc("POST /api/agent/roadmap-summary", h.RoadmapSummary)

	mux.HandleFunc("POST /api/agent/admin-summary", h.AdminSummary)
	mux.HandleFunc("POST /api/agent/ml-engineer-summary", h.MLEngineerSummary)
	mux.HandleFunc("POST /api/agent/annotator-summary", h.AnnotatorSummary)
	mux.HandleFunc("POST /api/agent/expert-summary", h.ExpertSummary)
	mux.HandleFunc("POST /api/agent/analyst-summary", h.AnalystSummary)

	mux.HandleFunc("POST /api/agent/object-explanation", h.ObjectExplanation)

	mux.HandleFunc("POST /api/agent/collection-plan", h.CollectionPlan)
	mux.HandleFunc("POST /api/agent/synthetic-plan", h.SyntheticPlan)

	mux.HandleFunc("POST /api/agent/generate-collection-tasks", h.GenerateCollectionTasks)
	mux.HandleFunc("POST /api/agent/generate-synthetic-tasks", h.GenerateSyntheticTasks)
	mux.HandleFunc("POST /api/agent/generate-annotator-brief", h.GenerateAnnotatorBrief)
	mux.HandleFunc("POST /api/agent/generate-expert-brief", h.GenerateExpertBrief)
}
