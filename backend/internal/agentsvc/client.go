package agentsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"hakaton/backend/internal/agent"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) do(ctx context.Context, endpoint string, reqBody, respBody any) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(reqBody); err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, &body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("agent-service request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(raw, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("agent-service error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("agent-service returned %d", resp.StatusCode)
	}

	if err := json.Unmarshal(raw, respBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// ---- Level 1 ----

type datasetSummaryReq struct {
	Readiness       float64  `json:"readiness"`
	TotalObjects    int      `json:"total_objects"`
	ReviewItems     int      `json:"review_items"`
	ImbalanceIndex  float64  `json:"imbalance_index"`
	AvgLabelError   float64  `json:"avg_label_error_probability"`
	MissingFiles    int      `json:"missing_files"`
	Recommendations []string `json:"recommendations"`
	Roadmap         []string `json:"roadmap"`
}

func (c *Client) GenerateDatasetSummary(ctx context.Context, readiness float64, totalObjects, reviewItems int, imbalanceIndex, avgLabelError float64, missingFiles int, recommendations, roadmap []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/dataset-summary", datasetSummaryReq{
		Readiness: readiness, TotalObjects: totalObjects, ReviewItems: reviewItems,
		ImbalanceIndex: imbalanceIndex, AvgLabelError: avgLabelError, MissingFiles: missingFiles,
		Recommendations: recommendations, Roadmap: roadmap,
	}, &resp)
	return resp, err
}

type textListReq struct {
	Recommendations []string `json:"recommendations"`
}

func (c *Client) GenerateRecommendationsSummary(ctx context.Context, recommendations []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/recommendations-summary", textListReq{Recommendations: recommendations}, &resp)
	return resp, err
}

type roadmapReq struct {
	Roadmap []string `json:"roadmap"`
}

func (c *Client) GenerateRoadmapSummary(ctx context.Context, roadmap []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/roadmap-summary", roadmapReq{Roadmap: roadmap}, &resp)
	return resp, err
}

// ---- Level 2: Role-aware ----

type adminSummaryReq struct {
	Readiness       float64  `json:"readiness"`
	TotalObjects    int      `json:"total_objects"`
	ReviewItems     int      `json:"review_items"`
	TeamMembers     int      `json:"team_members"`
	ReviewedObjects int      `json:"reviewed_objects"`
	ReadyForExport  string   `json:"ready_for_export"`
	Recommendations []string `json:"recommendations"`
	Roadmap         []string `json:"roadmap"`
}

func (c *Client) GenerateAdminSummary(ctx context.Context, readiness float64, totalObjects, reviewItems, teamMembers, reviewedObjects int, readyForExport string, recommendations, roadmap []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/admin-summary", adminSummaryReq{
		Readiness: readiness, TotalObjects: totalObjects, ReviewItems: reviewItems,
		TeamMembers: teamMembers, ReviewedObjects: reviewedObjects, ReadyForExport: readyForExport,
		Recommendations: recommendations, Roadmap: roadmap,
	}, &resp)
	return resp, err
}

type mlEngineerSummaryReq struct {
	Readiness       float64  `json:"readiness"`
	TotalObjects    int      `json:"total_objects"`
	ImbalanceIndex  float64  `json:"imbalance_index"`
	AvgLabelError   float64  `json:"avg_label_error_probability"`
	AvgEntropy      float64  `json:"avg_entropy"`
	ReviewItems     int      `json:"review_items"`
	Recommendations []string `json:"recommendations"`
	Roadmap         []string `json:"roadmap"`
}

func (c *Client) GenerateMLEngineerSummary(ctx context.Context, readiness float64, totalObjects int, imbalanceIndex, avgLabelError, avgEntropy float64, reviewItems int, recommendations, roadmap []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/ml-engineer-summary", mlEngineerSummaryReq{
		Readiness: readiness, TotalObjects: totalObjects, ImbalanceIndex: imbalanceIndex,
		AvgLabelError: avgLabelError, AvgEntropy: avgEntropy, ReviewItems: reviewItems,
		Recommendations: recommendations, Roadmap: roadmap,
	}, &resp)
	return resp, err
}

type annotatorSummaryReq struct {
	AssignedCount     int      `json:"assigned_count"`
	ReviewItems       int      `json:"review_items"`
	StatusDistribution string  `json:"status_distribution"`
	TopReviewObjects  []string `json:"top_review_objects"`
}

func (c *Client) GenerateAnnotatorSummary(ctx context.Context, assignedCount, reviewItems int, statusDistribution string, topReviewObjects []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/annotator-summary", annotatorSummaryReq{
		AssignedCount: assignedCount, ReviewItems: reviewItems,
		StatusDistribution: statusDistribution, TopReviewObjects: topReviewObjects,
	}, &resp)
	return resp, err
}

type expertSummaryReq struct {
	SentToExpert       int    `json:"sent_to_expert"`
	PendingDecisions   int    `json:"pending_decisions"`
	LabelConflicts     int    `json:"label_conflicts"`
	ClassesNeedingReview string `json:"classes_needing_review"`
}

func (c *Client) GenerateExpertSummary(ctx context.Context, sentToExpert, pendingDecisions, labelConflicts int, classesNeedingReview string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/expert-summary", expertSummaryReq{
		SentToExpert: sentToExpert, PendingDecisions: pendingDecisions,
		LabelConflicts: labelConflicts, ClassesNeedingReview: classesNeedingReview,
	}, &resp)
	return resp, err
}

type analystSummaryReq struct {
	Readiness            float64 `json:"readiness"`
	VersionComparison    string  `json:"version_comparison"`
	ObjectsCount         int     `json:"objects_count"`
	DuplicateCandidates  int     `json:"duplicate_candidates"`
	LabelErrorCandidates int     `json:"label_error_candidates"`
	QualityIssues        int     `json:"quality_issues"`
	ImbalanceIndex       float64 `json:"imbalance_index"`
}

func (c *Client) GenerateAnalystSummary(ctx context.Context, readiness float64, versionComparison string, objectsCount, duplicateCandidates, labelErrorCandidates, qualityIssues int, imbalanceIndex float64) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/analyst-summary", analystSummaryReq{
		Readiness: readiness, VersionComparison: versionComparison, ObjectsCount: objectsCount,
		DuplicateCandidates: duplicateCandidates, LabelErrorCandidates: labelErrorCandidates,
		QualityIssues: qualityIssues, ImbalanceIndex: imbalanceIndex,
	}, &resp)
	return resp, err
}

// ---- Level 2: Object Explanation ----

type objectExplanationReq struct {
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

func (c *Client) GenerateObjectExplanation(ctx context.Context, filePath, currentLabel, predictedLabel string, confidence float64, status string, entropy, uncertainty, labelErrorProb, duplicateScore, qualityScore, finalScore float64, recommendation string, reasons, actions, comments []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/object-explanation", objectExplanationReq{
		FilePath: filePath, CurrentLabel: currentLabel, PredictedLabel: predictedLabel,
		Confidence: confidence, Status: status,
		Entropy: entropy, Uncertainty: uncertainty, LabelErrorProb: labelErrorProb,
		DuplicateScore: duplicateScore, QualityScore: qualityScore, FinalScore: finalScore,
		Recommendation: recommendation, Reasons: reasons, Actions: actions, Comments: comments,
	}, &resp)
	return resp, err
}

// ---- Level 2: Plans ----

type planReq struct {
	ClassActionPlan []string `json:"class_action_plan"`
}

func (c *Client) GenerateCollectionPlan(ctx context.Context, classActionPlan []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/collection-plan", planReq{ClassActionPlan: classActionPlan}, &resp)
	return resp, err
}

func (c *Client) GenerateSyntheticPlan(ctx context.Context, classActionPlan []string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/synthetic-plan", planReq{ClassActionPlan: classActionPlan}, &resp)
	return resp, err
}

// ---- Level 2: Task Generation ----

type genTasksReq struct {
	ClassActionPlan []string `json:"class_action_plan"`
}

func (c *Client) GenerateCollectionTaskDrafts(ctx context.Context, classActionPlan []string) (string, error) {
	var resp json.RawMessage
	err := c.do(ctx, "/api/agent/generate-collection-tasks", genTasksReq{ClassActionPlan: classActionPlan}, &resp)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) GenerateSyntheticTaskDrafts(ctx context.Context, classActionPlan []string) (string, error) {
	var resp json.RawMessage
	err := c.do(ctx, "/api/agent/generate-synthetic-tasks", genTasksReq{ClassActionPlan: classActionPlan}, &resp)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

type annotatorBriefReq struct {
	ReviewItems  int    `json:"review_items"`
	Classes      string `json:"classes"`
	CommonIssues string `json:"common_issues"`
	TopPriority  string `json:"top_priority"`
}

func (c *Client) GenerateAnnotatorBrief(ctx context.Context, reviewItems int, classes, commonIssues, topPriority string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/generate-annotator-brief", annotatorBriefReq{
		ReviewItems: reviewItems, Classes: classes,
		CommonIssues: commonIssues, TopPriority: topPriority,
	}, &resp)
	return resp, err
}

type expertBriefReq struct {
	ExpertObjects   int    `json:"expert_objects"`
	ConflictClasses string `json:"conflict_classes"`
	KeyDecisions    string `json:"key_decisions"`
}

func (c *Client) GenerateExpertBrief(ctx context.Context, expertObjects int, conflictClasses, keyDecisions string) (agent.AgentResponse, error) {
	var resp agent.AgentResponse
	err := c.do(ctx, "/api/agent/generate-expert-brief", expertBriefReq{
		ExpertObjects: expertObjects, ConflictClasses: conflictClasses, KeyDecisions: keyDecisions,
	}, &resp)
	return resp, err
}
