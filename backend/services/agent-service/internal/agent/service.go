package agent

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type AgentResponse struct {
	Summary             string   `json:"summary"`
	KeyFindings         []string `json:"key_findings"`
	Risks               []string `json:"risks"`
	RecommendedNextSteps []string `json:"recommended_next_steps"`
	UsedContextFields   []string `json:"used_context_fields"`
}

type Service struct {
	client *Client
}

func NewService(baseURL, model string, timeout time.Duration) *Service {
	return &Service{
		client: NewClient(baseURL, model, timeout),
	}
}

func (s *Service) GenerateDatasetSummary(ctx context.Context, readiness float64, totalObjects, reviewItems int, imbalanceIndex, avgLabelError float64, missingFiles int, recommendations, roadmap []string) (AgentResponse, error) {
	recsText := joinBullets(recommendations)
	roadText := joinBullets(roadmap)

	prompt := fmt.Sprintf(datasetSummaryPrompt, readiness, totalObjects, reviewItems, imbalanceIndex, avgLabelError, missingFiles, recsText, roadText)

	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}

	usedFields := []string{
		"readiness_score", "total_objects", "review_queue_count",
		"imbalance_index", "avg_label_error_probability", "missing_files_count",
	}
	if len(recommendations) > 0 {
		usedFields = append(usedFields, "recommendations")
	}
	if len(roadmap) > 0 {
		usedFields = append(usedFields, "roadmap")
	}

	ctxVals := map[string]float64{
		"readiness": readiness, "total": float64(totalObjects),
		"review": float64(reviewItems), "imbalance": imbalanceIndex,
		"error_probability": avgLabelError, "missing": float64(missingFiles),
	}

	resp := s.guardResponse(ctx, AgentResponse{
		Summary:              response,
		UsedContextFields:    usedFields,
		RecommendedNextSteps: extractSteps(response),
	}, ctxVals)

	return resp.Response, nil
}

func (s *Service) GenerateRecommendationsSummary(ctx context.Context, recommendations []string) (AgentResponse, error) {
	if len(recommendations) == 0 {
		return AgentResponse{Summary: "No recommendations available.", UsedContextFields: []string{}}, nil
	}

	prompt := fmt.Sprintf(recommendationsPrompt, joinBullets(recommendations))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}

	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		KeyFindings:       recommendations,
		UsedContextFields: []string{"recommendations"},
	}, nil)

	return resp.Response, nil
}

func (s *Service) GenerateRoadmapSummary(ctx context.Context, roadmap []string) (AgentResponse, error) {
	if len(roadmap) == 0 {
		return AgentResponse{Summary: "No roadmap available.", UsedContextFields: []string{}}, nil
	}

	prompt := fmt.Sprintf(roadmapPrompt, joinBullets(roadmap))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}

	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		UsedContextFields: []string{"roadmap"},
	}, nil)

	return resp.Response, nil
}

func joinBullets(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	return "- " + strings.Join(items, "\n- ")
}

// ---- Level 2: Role-aware summaries ----

func (s *Service) GenerateAdminSummary(ctx context.Context, readiness float64, totalObjects, reviewItems, teamMembers, reviewedObjects int, readyForExport string, recommendations, roadmap []string) (AgentResponse, error) {
	prompt := fmt.Sprintf(adminSummaryPrompt, readiness, totalObjects, reviewItems, teamMembers, reviewedObjects, readyForExport, joinBullets(recommendations), joinBullets(roadmap))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		UsedContextFields: []string{"readiness_score", "total_objects", "review_queue_count", "team_members", "reviewed_objects", "recommendations", "roadmap"},
	}, map[string]float64{
		"readiness": readiness, "total": float64(totalObjects),
		"review": float64(reviewItems), "team": float64(teamMembers),
		"reviewed": float64(reviewedObjects),
	})
	return resp.Response, nil
}

func (s *Service) GenerateMLEngineerSummary(ctx context.Context, readiness float64, totalObjects int, imbalanceIndex, avgLabelError, avgEntropy float64, reviewItems int, recommendations, roadmap []string) (AgentResponse, error) {
	prompt := fmt.Sprintf(mlEngineerSummaryPrompt, readiness, totalObjects, imbalanceIndex, avgLabelError, avgEntropy, reviewItems, joinBullets(recommendations), joinBullets(roadmap))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		UsedContextFields: []string{"readiness_score", "total_objects", "imbalance_index", "avg_label_error_probability", "avg_entropy", "review_queue_count"},
	}, map[string]float64{
		"readiness": readiness, "total": float64(totalObjects),
		"imbalance": imbalanceIndex, "error_probability": avgLabelError,
		"entropy": avgEntropy, "review": float64(reviewItems),
	})
	return resp.Response, nil
}

func (s *Service) GenerateAnnotatorSummary(ctx context.Context, assignedCount, reviewItems int, statusDistribution string, topReviewObjects []string) (AgentResponse, error) {
	prompt := fmt.Sprintf(annotatorSummaryPrompt, assignedCount, reviewItems, statusDistribution, joinBullets(topReviewObjects))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		UsedContextFields: []string{"assigned_objects", "review_queue_count", "status_distribution", "top_review_objects"},
	}, map[string]float64{
		"assigned": float64(assignedCount), "review": float64(reviewItems),
	})
	return resp.Response, nil
}

func (s *Service) GenerateExpertSummary(ctx context.Context, sentToExpert, pendingDecisions, labelConflicts int, classesNeedingReview string) (AgentResponse, error) {
	prompt := fmt.Sprintf(domainExpertSummaryPrompt, sentToExpert, pendingDecisions, labelConflicts, classesNeedingReview)
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		UsedContextFields: []string{"sent_to_expert", "pending_expert_decisions", "label_conflicts", "classes_needing_review"},
	}, map[string]float64{
		"expert": float64(sentToExpert), "pending": float64(pendingDecisions),
		"conflicts": float64(labelConflicts),
	})
	return resp.Response, nil
}

func (s *Service) GenerateAnalystSummary(ctx context.Context, readiness float64, versionComparison string, objectsCount, duplicateCandidates, labelErrorCandidates, qualityIssues int, imbalanceIndex float64) (AgentResponse, error) {
	prompt := fmt.Sprintf(dataAnalystSummaryPrompt, readiness, versionComparison, objectsCount, imbalanceIndex, duplicateCandidates, labelErrorCandidates, qualityIssues)
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		UsedContextFields: []string{"readiness_score", "version_comparison", "objects_count", "class_imbalance", "duplicate_candidates", "label_error_candidates", "quality_issues"},
	}, map[string]float64{
		"readiness": readiness, "objects": float64(objectsCount),
		"imbalance": imbalanceIndex, "duplicates": float64(duplicateCandidates),
		"errors": float64(labelErrorCandidates), "quality": float64(qualityIssues),
	})
	return resp.Response, nil
}

// ---- Level 2: Object Explanation ----

func (s *Service) GenerateObjectExplanation(ctx context.Context, filePath, currentLabel, predictedLabel string, confidence float64, status string, entropy, uncertainty, labelErrorProb, duplicateScore, qualityScore, finalScore float64, recommendation string, reasons, actions, comments []string) (AgentResponse, error) {
	prompt := fmt.Sprintf(objectExplanationPrompt,
		filePath, currentLabel, predictedLabel, confidence, status,
		entropy, uncertainty, labelErrorProb, duplicateScore, qualityScore, finalScore,
		recommendation, joinBullets(reasons), joinBullets(actions), joinBullets(comments))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary:           response,
		UsedContextFields: []string{"file_path", "label", "predicted_label", "status", "metrics", "reasons", "actions", "comments"},
	}, map[string]float64{
		"confidence": confidence, "entropy": entropy,
		"uncertainty": uncertainty, "error_probability": labelErrorProb,
		"duplicate": duplicateScore, "quality": qualityScore,
		"score": finalScore,
	})
	return resp.Response, nil
}

// ---- Level 2: Collection / Synthetic Plans ----

func (s *Service) GenerateCollectionPlan(ctx context.Context, classActionPlan []string) (AgentResponse, error) {
	if len(classActionPlan) == 0 {
		return AgentResponse{Summary: "No collection tasks needed at this time.", UsedContextFields: []string{}}, nil
	}
	prompt := fmt.Sprintf(collectionPlanPrompt, joinBullets(classActionPlan))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{Summary: response, UsedContextFields: []string{"class_action_plan"}}, nil)
	return resp.Response, nil
}

func (s *Service) GenerateSyntheticPlan(ctx context.Context, classActionPlan []string) (AgentResponse, error) {
	if len(classActionPlan) == 0 {
		return AgentResponse{Summary: "No synthetic tasks needed at this time.", UsedContextFields: []string{}}, nil
	}
	prompt := fmt.Sprintf(syntheticPlanPrompt, joinBullets(classActionPlan))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{Summary: response, UsedContextFields: []string{"class_action_plan"}}, nil)
	return resp.Response, nil
}

// ---- Level 2: Task Generation ----

func (s *Service) GenerateCollectionTaskDrafts(ctx context.Context, classActionPlan []string) (string, error) {
	if len(classActionPlan) == 0 {
		return "[]", nil
	}
	prompt := fmt.Sprintf(generateCollectionTasksPrompt, joinBullets(classActionPlan))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}
	return extractJSON(response), nil
}

func (s *Service) GenerateSyntheticTaskDrafts(ctx context.Context, classActionPlan []string) (string, error) {
	if len(classActionPlan) == 0 {
		return "[]", nil
	}
	prompt := fmt.Sprintf(generateSyntheticTasksPrompt, joinBullets(classActionPlan))
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}
	return extractJSON(response), nil
}

func (s *Service) GenerateAnnotatorBrief(ctx context.Context, reviewItems int, classes, commonIssues, topPriority string) (AgentResponse, error) {
	prompt := fmt.Sprintf(annotatorBriefPrompt, reviewItems, classes, commonIssues, topPriority)
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary: response, UsedContextFields: []string{"review_items", "classes", "common_issues", "top_priority"},
	}, map[string]float64{"review": float64(reviewItems)})
	return resp.Response, nil
}

func (s *Service) GenerateExpertBrief(ctx context.Context, expertObjects int, conflictClasses, keyDecisions string) (AgentResponse, error) {
	prompt := fmt.Sprintf(expertBriefPrompt, expertObjects, conflictClasses, keyDecisions)
	response, err := s.client.Generate(ctx, prompt)
	if err != nil {
		return AgentResponse{}, err
	}
	resp := s.guardResponse(ctx, AgentResponse{
		Summary: response, UsedContextFields: []string{"expert_objects", "conflict_classes", "key_decisions"},
	}, map[string]float64{"expert": float64(expertObjects)})
	return resp.Response, nil
}

// ---- Guardrails ----

type GuardrailResult struct {
	Response AgentResponse
	Warnings []string
	Rejected bool
}

func (s *Service) guardResponse(ctx context.Context, resp AgentResponse, contextValues map[string]float64) GuardrailResult {
	var warnings []string

	if strings.TrimSpace(resp.Summary) == "" {
		resp.Summary = "Agent returned an empty response."
		warnings = append(warnings, "empty_response")
	}

	if len(resp.KeyFindings) == 0 {
		resp.KeyFindings = []string{}
	}
	if len(resp.Risks) == 0 {
		resp.Risks = []string{}
	}
	if len(resp.RecommendedNextSteps) == 0 {
		resp.RecommendedNextSteps = []string{}
	}

	// Check for hallucinated numbers against known context values
	for label, actual := range contextValues {
		matches := extractNumbersAroundLabel(resp.Summary, label)
		for _, m := range matches {
			if actual > 0 {
				ratio := m / actual
				if ratio > 1.5 || ratio < 0.5 {
					warnings = append(warnings, fmt.Sprintf("suspicious_number(%s: response=%.2f, actual=%.2f)", label, m, actual))
				}
			}
		}
	}

	if len(warnings) > 0 && len(warnings) >= 3 {
		resp.Summary = fmt.Sprintf("[Quality check] %s\n\n%s",
			"The response contained several inconsistencies and was sanitized.",
			resp.Summary)
	}

	return GuardrailResult{
		Response: resp,
		Warnings: warnings,
		Rejected: false,
	}
}

func extractNumbersAroundLabel(text, label string) []float64 {
	var result []float64
	lower := strings.ToLower(text)
	llabel := strings.ToLower(label)
	idx := strings.Index(lower, llabel)
	if idx < 0 {
		return nil
	}

	window := 120
	start := idx - window
	if start < 0 {
		start = 0
	}
	end := idx + len(label) + window
	if end > len(text) {
		end = len(text)
	}
	segment := text[start:end]

	for _, token := range strings.Fields(segment) {
		token = strings.Trim(token, ".,;:!?%()[]")
		var val float64
		if n, _ := fmt.Sscanf(token, "%f", &val); n == 1 {
			if val > 0 && val < 100000 {
				result = append(result, val)
			}
		}
	}
	return result
}

// ---- Helpers ----

func extractJSON(text string) string {
	start := strings.Index(text, "[")
	if start == -1 {
		start = strings.Index(text, "{")
	}
	if start == -1 {
		return "[]"
	}
	end := strings.LastIndex(text, "]")
	if end == -1 {
		end = strings.LastIndex(text, "}")
	}
	if end == -1 || end < start {
		return "[]"
	}
	return text[start : end+1]
}

func extractSteps(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	var steps []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "*") {
			continue
		}
		if len(steps) < 5 && !strings.HasPrefix(line, "-") {
			step := strings.TrimPrefix(line, "- ")
			steps = append(steps, step)
		}
	}
	return steps
}
