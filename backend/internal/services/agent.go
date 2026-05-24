package services

import (
	"context"
	"fmt"
	"strings"

	"hakaton/backend/internal/agent"
	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func (s *Service) AgentDatasetSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	dashboard, err := s.Dashboard(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	recs, err := s.Recommendations(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	road, err := s.Roadmap(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	recTexts := make([]string, len(recs))
	for i, r := range recs {
		recTexts[i] = fmt.Sprintf("[%s/%s] %s: %s (%d objects)", r.Priority, r.Type, r.Title, r.Description, r.AffectedObjectsCount)
	}
	roadTexts := make([]string, len(road))
	for i, r := range road {
		roadTexts[i] = fmt.Sprintf("Step %d: %s - %s (impact: %s)", r.Priority, r.Title, r.Description, r.ExpectedImpact)
	}

	return s.agentsvc.GenerateDatasetSummary(ctx,
		dashboard.ReadinessScore,
		dashboard.ObjectsCount,
		dashboard.ReviewItems,
		dashboard.ImbalanceIndex,
		dashboard.AvgLabelErrorProbability,
		dashboard.MissingFilesCount,
		recTexts, roadTexts,
	)
}

func (s *Service) AgentRecommendationsSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	recs, err := s.Recommendations(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	recTexts := make([]string, len(recs))
	for i, r := range recs {
		recTexts[i] = fmt.Sprintf("[%s/%s] %s: %s (%d objects)", r.Priority, r.Type, r.Title, r.Description, r.AffectedObjectsCount)
	}
	return s.agentsvc.GenerateRecommendationsSummary(ctx, recTexts)
}

func (s *Service) AgentRoadmapSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	road, err := s.Roadmap(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	roadTexts := make([]string, len(road))
	for i, r := range road {
		roadTexts[i] = fmt.Sprintf("Step %d: %s - %s (impact: %s)", r.Priority, r.Title, r.Description, r.ExpectedImpact)
	}
	return s.agentsvc.GenerateRoadmapSummary(ctx, roadTexts)
}

// ---- Level 2: Role-aware summaries ----

func (s *Service) AgentAdminSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	dash, err := s.Dashboard(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	members, _ := s.ListProjectMembers(ctx, projectID)
	recs, _ := s.Recommendations(ctx, projectID)
	road, _ := s.Roadmap(ctx, projectID)

	readyForExport := "no"
	if dash.ReadinessScore >= 70 {
		readyForExport = "yes"
	}

	recTexts := make([]string, len(recs))
	for i, r := range recs {
		recTexts[i] = fmt.Sprintf("[%s/%s] %s: %s", r.Priority, r.Type, r.Title, r.Description)
	}
	roadTexts := make([]string, len(road))
	for i, r := range road {
		roadTexts[i] = fmt.Sprintf("Step %d: %s (impact: %s)", r.Priority, r.Title, r.ExpectedImpact)
	}

	return s.agentsvc.GenerateAdminSummary(ctx, dash.ReadinessScore, dash.ObjectsCount, dash.ReviewItems,
		len(members), dash.StatusCounts["reviewed"], readyForExport, recTexts, roadTexts)
}

func (s *Service) AgentMLEngineerSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	dash, err := s.Dashboard(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	recs, _ := s.Recommendations(ctx, projectID)
	road, _ := s.Roadmap(ctx, projectID)

	recTexts := make([]string, len(recs))
	for i, r := range recs {
		recTexts[i] = fmt.Sprintf("[%s] %s: %s (%d objects)", r.Priority, r.Title, r.Description, r.AffectedObjectsCount)
	}
	roadTexts := make([]string, len(road))
	for i, r := range road {
		roadTexts[i] = fmt.Sprintf("Step %d: %s (impact: %s)", r.Priority, r.Title, r.ExpectedImpact)
	}

	return s.agentsvc.GenerateMLEngineerSummary(ctx, dash.ReadinessScore, dash.ObjectsCount,
		dash.ImbalanceIndex, dash.AvgLabelErrorProbability, dash.AvgEntropy,
		dash.ReviewItems, recTexts, roadTexts)
}

func (s *Service) AgentAnnotatorSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	_, objects, metrics, err := s.latestObjectsAndMetrics(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	statusDist := mapStatusDistribution(objects)
	var statusParts []string
	for k, v := range statusDist {
		statusParts = append(statusParts, fmt.Sprintf("%s:%d", k, v))
	}

	var topObjects []string
	for i, m := range metrics {
		if i >= 5 {
			break
		}
		obj := findObjectByID(objects, m.ObjectID)
		if obj != nil {
			topObjects = append(topObjects, fmt.Sprintf("%s (score=%.2f, reason=%s)", obj.FilePath, m.FinalScore, strings.Join(m.Reasons, ", ")))
		}
	}

	return s.agentsvc.GenerateAnnotatorSummary(ctx, len(objects), len(metrics),
		strings.Join(statusParts, ", "), topObjects)
}

func (s *Service) AgentExpertSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	actions, err := s.repo.ListProjectActions(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	var sentToExpert, pendingDecisions, labelConflicts int
	conflictClasses := map[string]bool{}
	for _, a := range actions {
		switch a.Action {
		case "send_to_expert":
			sentToExpert++
		case "expert_approve", "expert_reject":
			pendingDecisions++
		}
		if a.Action == "send_to_expert" {
			conflictClasses[a.NewValue] = true
		}
	}

	return s.agentsvc.GenerateExpertSummary(ctx, sentToExpert, sentToExpert-pendingDecisions,
		labelConflicts, strings.Join(mapKeys(conflictClasses), ", "))
}

func (s *Service) AgentAnalystSummary(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	dash, err := s.Dashboard(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	versionComparison := "single version available"
	if dash.DatasetVersion.ID != uuid.Nil {
		versionComparison = fmt.Sprintf("version %s with %d objects", dash.DatasetVersion.VersionName, dash.ObjectsCount)
	}

	duplicates := dash.StatusCounts["duplicate"]
	labelErrors := dash.StatusCounts["suspected_label_error"]
	qualityIssues := dash.StatusCounts["bad_quality"] + dash.MissingFilesCount

	return s.agentsvc.GenerateAnalystSummary(ctx, dash.ReadinessScore, versionComparison,
		dash.ObjectsCount, duplicates, labelErrors, qualityIssues, dash.ImbalanceIndex)
}

// ---- Level 2: Object Explanation ----

func (s *Service) AgentObjectExplanation(ctx context.Context, projectID, objectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	detail, err := s.GetObjectDetail(ctx, projectID, objectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	actions, _ := s.repo.ListObjectActions(ctx, projectID, objectID)
	comments, _ := s.repo.ListObjectComments(ctx, projectID, objectID)

	actionTexts := make([]string, len(actions))
	for i, a := range actions {
		actionTexts[i] = fmt.Sprintf("%s by %s (old=%s, new=%s)", a.Action, a.UserID.String(), a.OldValue, a.NewValue)
	}
	commentTexts := make([]string, len(comments))
	for i, c := range comments {
		commentTexts[i] = c.Text
	}

	var m *models.ObjectMetric
	if detail.Metrics != nil {
		m = detail.Metrics
	}

	entropy, uncertainty, labelErr, dupScore, qualScore, finalScore := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	rec := ""
	var reasons []string
	if m != nil {
		entropy = m.Entropy
		uncertainty = m.UncertaintyScore
		labelErr = m.LabelErrorProbability
		dupScore = m.DuplicateScore
		qualScore = m.QualityScore
		finalScore = m.FinalScore
		rec = m.Recommendation
		reasons = m.Reasons
	}

	return s.agentsvc.GenerateObjectExplanation(ctx,
		detail.FilePath, detail.Label, detail.PredictedLabel, detail.Confidence, detail.Status,
		entropy, uncertainty, labelErr, dupScore, qualScore, finalScore,
		rec, reasons, actionTexts, commentTexts)
}

// ---- Level 2: Collection / Synthetic Plans ----

func (s *Service) AgentCollectionPlan(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	plan, err := s.ClassActionPlan(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	planTexts := planToStrings(plan)
	return s.agentsvc.GenerateCollectionPlan(ctx, planTexts)
}

func (s *Service) AgentSyntheticPlan(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	plan, err := s.ClassActionPlan(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	planTexts := planToStrings(plan)
	return s.agentsvc.GenerateSyntheticPlan(ctx, planTexts)
}

// ---- Level 2: Task Generation ----

func (s *Service) AgentGenerateCollectionTasks(ctx context.Context, projectID uuid.UUID) (string, error) {
	if s.agentsvc == nil {
		return "", badRequest("agent service is not configured")
	}
	plan, err := s.ClassActionPlan(ctx, projectID)
	if err != nil {
		return "", err
	}
	planTexts := planToStrings(plan)
	return s.agentsvc.GenerateCollectionTaskDrafts(ctx, planTexts)
}

func (s *Service) AgentGenerateSyntheticTasks(ctx context.Context, projectID uuid.UUID) (string, error) {
	if s.agentsvc == nil {
		return "", badRequest("agent service is not configured")
	}
	plan, err := s.ClassActionPlan(ctx, projectID)
	if err != nil {
		return "", err
	}
	planTexts := planToStrings(plan)
	return s.agentsvc.GenerateSyntheticTaskDrafts(ctx, planTexts)
}

func (s *Service) AgentGenerateAnnotatorBrief(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	dash, err := s.Dashboard(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	classes := strings.Join(mapStrKeys(dash.LabelCounts), ", ")
	var commonIssues []string
	for status, count := range dash.StatusCounts {
		if count > 0 && status != "ok" {
			commonIssues = append(commonIssues, fmt.Sprintf("%s: %d", status, count))
		}
	}

	topPriority := "review queue"
	if dash.ImbalanceIndex > 0.3 {
		topPriority = "class imbalance"
	}

	return s.agentsvc.GenerateAnnotatorBrief(ctx, dash.ReviewItems, classes, strings.Join(commonIssues, "; "), topPriority)
}

func (s *Service) AgentGenerateExpertBrief(ctx context.Context, projectID uuid.UUID) (agent.AgentResponse, error) {
	if s.agentsvc == nil {
		return agent.AgentResponse{}, badRequest("agent service is not configured")
	}
	actions, err := s.repo.ListProjectActions(ctx, projectID)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	var expertObjects int
	conflictClasses := map[string]bool{}
	for _, a := range actions {
		if a.Action == "send_to_expert" {
			expertObjects++
			conflictClasses[a.NewValue] = true
		}
	}

	return s.agentsvc.GenerateExpertBrief(ctx, expertObjects,
		strings.Join(mapKeys(conflictClasses), ", "), "resolve label conflicts")
}

// ---- Helpers ----

func findObjectByID(objects []models.DataObject, id uuid.UUID) *models.DataObject {
	for i := range objects {
		if objects[i].ID == id {
			return &objects[i]
		}
	}
	return nil
}

func mapStatusDistribution(objects []models.DataObject) map[string]int {
	result := map[string]int{}
	for _, o := range objects {
		status := o.Status
		if status == "" {
			status = "ok"
		}
		result[status]++
	}
	return result
}

func mapKeys(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapStrKeys(m map[string]int) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func planToStrings(plan []models.ClassActionPlanItem) []string {
	result := make([]string, len(plan))
	for i, item := range plan {
		result[i] = fmt.Sprintf("%s: problem=%s, action=%s, real_collection=%.2f, synthetic_score=%.2f, risk=%s, impact=%s",
			item.Class, item.Problem, item.RecommendedAction, item.RealCollectionPriority,
			item.SyntheticDataCandidateScore, item.Risk, item.ExpectedImpact)
	}
	return result
}
