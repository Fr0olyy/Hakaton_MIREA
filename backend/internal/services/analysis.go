package services

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

type localAnalysis struct {
	readiness       float64
	metrics         []models.ObjectMetric
	recommendations []models.Recommendation
	roadmap         []models.RoadmapItem
	summary         datasetSummary
	objectStatuses  map[uuid.UUID]string
}

func (s *Service) Analyze(ctx context.Context, projectID uuid.UUID) (models.AnalysisJob, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return models.AnalysisJob{}, err
	}
	version, err := s.repo.LatestDatasetVersion(ctx, projectID)
	if err != nil {
		return models.AnalysisJob{}, err
	}
	objects, err := s.repo.ListObjects(ctx, version.ID)
	if err != nil {
		return models.AnalysisJob{}, err
	}

	job := models.AnalysisJob{
		ID:               uuid.New(),
		ProjectID:        projectID,
		DatasetVersionID: version.ID,
		Status:           "running",
	}
	job, err = s.repo.CreateAnalysisJob(ctx, job)
	if err != nil {
		return job, err
	}

	result, err := s.analyzeWithMLService(ctx, project, version, objects)
	if err != nil {
		result = s.analyzeLocal(project, version, objects)
	}
	if err := s.repo.FinishAnalysis(ctx, job.ID, result.readiness, result.metrics, result.recommendations, result.roadmap, result.objectStatuses); err != nil {
		_ = s.repo.FailAnalysisJob(ctx, job.ID, err.Error())
		return job, err
	}
	job.Status = "done"
	job.FinishedAt = time.Now()
	return job, nil
}

func (s *Service) analyzeLocal(project models.Project, version models.DatasetVersion, objects []models.DataObject) localAnalysis {
	classes := classUniverse(project, objects)
	classCounts, labeledTotal := labelCounts(objects, classes)
	distribution := classDistribution(classCounts, classes, labeledTotal)
	imbalance := imbalanceIndex(distribution, classes, labeledTotal)
	duplicateScores := s.duplicateScores(project.ID, objects)

	metrics := make([]models.ObjectMetric, 0, len(objects))
	var totalQuality, totalLabelRisk, totalUncertainty float64
	for i, object := range objects {
		probabilities := probabilitiesForObject(object, classes, i)
		entropyBits := entropyFromProbabilities(probabilities)
		uncertainty := normalizedEntropy(entropyBits, len(probabilities))
		labelRisk := labelErrorProbability(object, probabilities, uncertainty)
		duplicateScore := duplicateScores[object.ID]
		rarity := rarityScore(object.Label, classCounts, labeledTotal)
		deficit := classDeficitScore(object.Label, classCounts, classes, labeledTotal)
		qualityRisk := qualityRisk(object, duplicateScore)
		qualityScore := clamp(1-qualityRisk, 0, 1)
		novelty := clamp(0.5*rarity+0.5*(1-duplicateScore), 0, 1)
		final := clamp(
			0.35*uncertainty+
				0.25*labelRisk+
				0.20*deficit+
				0.10*novelty+
				0.10*qualityRisk,
			0,
			1,
		)

		metric := models.ObjectMetric{
			ID:                    uuid.New(),
			ObjectID:              object.ID,
			Entropy:               entropyBits,
			UncertaintyScore:      uncertainty,
			LabelErrorProbability: labelRisk,
			DuplicateScore:        duplicateScore,
			RarityScore:           rarity,
			ClassDeficitScore:     deficit,
			QualityScore:          qualityScore,
			NoveltyScore:          novelty,
			ObjectUtilityScore:    final,
			FinalScore:            final,
			Reasons:               metricReasons(object, uncertainty, labelRisk, deficit, duplicateScore, final),
			Recommendation:        metricRecommendation(object, uncertainty, labelRisk, deficit, duplicateScore),
			Probabilities:         probabilitiesPayload(probabilities),
		}
		metrics = append(metrics, metric)
		totalQuality += qualityScore
		totalLabelRisk += labelRisk
		totalUncertainty += uncertainty
	}

	avgQuality := average(totalQuality, len(metrics))
	avgLabelRisk := average(totalLabelRisk, len(metrics))
	avgUncertainty := average(totalUncertainty, len(metrics))
	readiness := clamp(100*(0.35*avgQuality+0.25*(1-avgLabelRisk)+0.20*(1-avgUncertainty)+0.20*(1-imbalance)), 0, 100)

	summary := datasetSummary{
		ClassDistribution:        distribution,
		ImbalanceIndex:           imbalance,
		AvgEntropy:               averageMetric(metrics, func(metric models.ObjectMetric) float64 { return metric.Entropy }),
		AvgLabelErrorProbability: averageMetric(metrics, func(metric models.ObjectMetric) float64 { return metric.LabelErrorProbability }),
		MissingFilesCount:        countObjects(objects, func(object models.DataObject) bool { return object.Status == statusMissingFile }),
		ReviewItems:              countMetrics(metrics, func(metric models.ObjectMetric) bool { return metric.FinalScore >= reviewThreshold }),
		AnalysisSource:           analysisSourceBackendLocal,
	}
	recommendations := buildRecommendations(project, version, objects, metrics, classCounts, classes, labeledTotal)
	roadmap := buildRoadmap(project, version, recommendations, summary)

	return localAnalysis{readiness: readiness, metrics: metrics, recommendations: recommendations, roadmap: roadmap, summary: summary}
}

func classUniverse(project models.Project, objects []models.DataObject) []string {
	seen := map[string]struct{}{}
	classes := make([]string, 0, len(project.Classes))
	add := func(class string) {
		class = strings.TrimSpace(class)
		if class == "" {
			return
		}
		if _, ok := seen[class]; ok {
			return
		}
		seen[class] = struct{}{}
		classes = append(classes, class)
	}
	for _, class := range project.Classes {
		add(class)
	}
	var discovered []string
	for _, object := range objects {
		discovered = append(discovered, object.Label, object.PredictedLabel)
		for class := range probabilitiesFromMetadata(object) {
			discovered = append(discovered, class)
		}
	}
	sort.Strings(discovered)
	for _, class := range discovered {
		add(class)
	}
	if len(classes) == 0 {
		classes = append(classes, "unknown")
	}
	return classes
}

func labelCounts(objects []models.DataObject, classes []string) (map[string]int, int) {
	counts := map[string]int{}
	for _, class := range classes {
		counts[class] = 0
	}
	total := 0
	for _, object := range objects {
		if object.Label == "" {
			continue
		}
		counts[object.Label]++
		total++
	}
	return counts, total
}

func classDistribution(counts map[string]int, classes []string, labeledTotal int) map[string]float64 {
	distribution := map[string]float64{}
	for _, class := range classes {
		if labeledTotal == 0 {
			distribution[class] = 0
			continue
		}
		distribution[class] = float64(counts[class]) / float64(labeledTotal)
	}
	return distribution
}

func imbalanceIndex(distribution map[string]float64, classes []string, labeledTotal int) float64 {
	if len(classes) <= 1 {
		return 0
	}
	if labeledTotal == 0 {
		return 1
	}
	expected := 1 / float64(len(classes))
	var deviation float64
	for _, class := range classes {
		deviation += math.Abs(distribution[class] - expected)
	}
	maxDeviation := 2 * (1 - expected)
	if maxDeviation == 0 {
		return 0
	}
	return clamp(deviation/maxDeviation, 0, 1)
}

func (s *Service) duplicateScores(projectID uuid.UUID, objects []models.DataObject) map[uuid.UUID]float64 {
	keys := map[uuid.UUID]string{}
	counts := map[string]int{}
	for _, object := range objects {
		if object.FilePath == "" || object.Status == statusMissingFile || object.Status == statusInvalid {
			continue
		}
		key, err := s.storage.FileSHA256(projectID, object.FilePath)
		if err != nil {
			key = "path:" + object.FilePath
		}
		keys[object.ID] = key
		counts[key]++
	}

	scores := map[uuid.UUID]float64{}
	for _, object := range objects {
		key := keys[object.ID]
		if key != "" && counts[key] > 1 {
			scores[object.ID] = 1
		} else {
			scores[object.ID] = 0
		}
	}
	return scores
}

func probabilitiesForObject(object models.DataObject, classes []string, index int) map[string]float64 {
	if probabilities := normalizeProbabilities(probabilitiesFromMetadata(object), classes); len(probabilities) > 0 {
		return probabilities
	}
	if object.PredictedLabel != "" || object.Confidence > 0 {
		return probabilitiesFromPrediction(object, classes)
	}
	return heuristicProbabilities(object, classes, index)
}

func probabilitiesFromMetadata(object models.DataObject) map[string]float64 {
	raw, ok := object.Metadata["probabilities"]
	if !ok || raw == nil {
		return nil
	}
	probabilities := map[string]float64{}
	switch typed := raw.(type) {
	case map[string]float64:
		for class, value := range typed {
			probabilities[class] = value
		}
	case map[string]any:
		for class, value := range typed {
			switch v := value.(type) {
			case float64:
				probabilities[class] = v
			case string:
				if parsed, err := strconv.ParseFloat(v, 64); err == nil {
					probabilities[class] = parsed
				}
			}
		}
	}
	return probabilities
}

func probabilitiesFromPrediction(object models.DataObject, classes []string) map[string]float64 {
	candidates := ensureClasses(classes, object.Label, object.PredictedLabel)
	if len(candidates) == 0 {
		candidates = []string{"unknown"}
	}
	predicted := strings.TrimSpace(object.PredictedLabel)
	if predicted == "" {
		predicted = firstNonEmpty(object.Label, candidates[0])
	}
	confidence := clamp(object.Confidence, 0, 1)
	if confidence == 0 {
		confidence = 0.72
	}
	probabilities := map[string]float64{}
	if len(candidates) == 1 {
		probabilities[candidates[0]] = 1
		return probabilities
	}
	remaining := clamp(1-confidence, 0, 1)
	spillover := remaining / float64(len(candidates)-1)
	for _, class := range candidates {
		probabilities[class] = spillover
	}
	probabilities[predicted] = confidence
	return normalizeProbabilities(probabilities, candidates)
}

func heuristicProbabilities(object models.DataObject, classes []string, index int) map[string]float64 {
	candidates := ensureClasses(classes, object.Label)
	if len(candidates) == 0 {
		candidates = []string{"unknown"}
	}
	focus := firstNonEmpty(object.Label, candidates[index%len(candidates)])
	confidence := clamp(0.62+float64(index%4)*0.04, 0, 0.78)
	if object.Status != statusOK {
		confidence = 0.52
	}
	return probabilitiesFromPrediction(models.DataObject{
		Label:          object.Label,
		PredictedLabel: focus,
		Confidence:     confidence,
	}, candidates)
}

func normalizeProbabilities(input map[string]float64, classes []string) map[string]float64 {
	if len(input) == 0 {
		return nil
	}
	probabilities := map[string]float64{}
	for _, class := range classes {
		probabilities[class] = 0
	}
	for class, value := range input {
		class = strings.TrimSpace(class)
		if class == "" {
			continue
		}
		probabilities[class] = clamp(value, 0, 1)
	}
	var sum float64
	for _, value := range probabilities {
		sum += value
	}
	if sum == 0 {
		return nil
	}
	for class, value := range probabilities {
		probabilities[class] = value / sum
	}
	return probabilities
}

func ensureClasses(classes []string, extras ...string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(classes)+len(extras))
	add := func(class string) {
		class = strings.TrimSpace(class)
		if class == "" {
			return
		}
		if _, ok := seen[class]; ok {
			return
		}
		seen[class] = struct{}{}
		result = append(result, class)
	}
	for _, class := range classes {
		add(class)
	}
	for _, class := range extras {
		add(class)
	}
	return result
}

func entropyFromProbabilities(probabilities map[string]float64) float64 {
	var result float64
	for _, value := range probabilities {
		if value > 0 {
			result -= value * math.Log2(value)
		}
	}
	return result
}

func normalizedEntropy(entropy float64, classCount int) float64 {
	if classCount <= 1 {
		return 0
	}
	return clamp(entropy/math.Log2(float64(classCount)), 0, 1)
}

func labelErrorProbability(object models.DataObject, probabilities map[string]float64, uncertainty float64) float64 {
	if object.Label == "" {
		return 1
	}
	if value, ok := probabilities[object.Label]; ok {
		return clamp(1-value, 0, 1)
	}
	if object.PredictedLabel != "" {
		confidence := clamp(object.Confidence, 0, 1)
		if confidence == 0 {
			confidence = 0.75
		}
		if object.PredictedLabel != object.Label {
			return confidence
		}
		return 1 - confidence
	}
	return clamp(uncertainty*0.7, 0, 1)
}

func rarityScore(label string, counts map[string]int, labeledTotal int) float64 {
	if label == "" || labeledTotal == 0 {
		return 1
	}
	return clamp(1-float64(counts[label])/float64(labeledTotal), 0, 1)
}

func classDeficitScore(label string, counts map[string]int, classes []string, labeledTotal int) float64 {
	if label == "" || labeledTotal == 0 || len(classes) == 0 {
		return 1
	}
	expected := float64(labeledTotal) / float64(len(classes))
	if expected == 0 {
		return 0
	}
	return clamp((expected-float64(counts[label]))/expected, 0, 1)
}

func qualityRisk(object models.DataObject, duplicateScore float64) float64 {
	switch object.Status {
	case statusInvalid:
		return 1
	case statusMissingFile:
		return 0.9
	case statusMissingLabel:
		return 0.75
	case statusUnknownClass:
		return 0.65
	}
	if duplicateScore >= 0.75 {
		return 0.45
	}
	return 0.1
}

func metricReasons(object models.DataObject, uncertainty, labelRisk, deficit, duplicateScore, final float64) []string {
	var reasons []string
	if object.Status != statusOK {
		reasons = append(reasons, "object status is "+object.Status)
	}
	if uncertainty >= 0.65 {
		reasons = append(reasons, "high prediction entropy")
	}
	if labelRisk >= 0.55 {
		reasons = append(reasons, "high label error probability")
	}
	if deficit >= 0.35 {
		reasons = append(reasons, "belongs to an underrepresented class")
	}
	if duplicateScore >= 0.75 {
		reasons = append(reasons, "duplicate image candidate")
	}
	if len(reasons) == 0 && final >= reviewThreshold {
		reasons = append(reasons, "useful object for manual review")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "dataset object looks consistent")
	}
	return reasons
}

func metricRecommendation(object models.DataObject, uncertainty, labelRisk, deficit, duplicateScore float64) string {
	switch object.Status {
	case statusInvalid:
		return "fix_row"
	case statusMissingFile:
		return "restore_file"
	case statusMissingLabel:
		return "add_label"
	case statusUnknownClass:
		return "verify_class"
	}
	if duplicateScore >= 0.75 {
		return "deduplicate"
	}
	if labelRisk >= 0.55 {
		return "verify_label"
	}
	if uncertainty >= 0.65 {
		return "manual_review"
	}
	if deficit >= 0.35 {
		return "keep_for_weak_class"
	}
	return "keep"
}

func probabilitiesPayload(probabilities map[string]float64) map[string]any {
	payload := map[string]any{"analysis_source": analysisSourceBackendLocal}
	for class, value := range probabilities {
		payload[class] = value
	}
	return payload
}

func buildRecommendations(project models.Project, version models.DatasetVersion, objects []models.DataObject, metrics []models.ObjectMetric, counts map[string]int, classes []string, labeledTotal int) []models.Recommendation {
	missingFiles := objectIDs(objects, func(object models.DataObject) bool { return object.Status == statusMissingFile })
	missingLabels := objectIDs(objects, func(object models.DataObject) bool { return object.Label == "" })
	unknownClasses := objectIDs(objects, func(object models.DataObject) bool { return object.Status == statusUnknownClass })
	duplicates := metricObjectIDs(metrics, func(metric models.ObjectMetric) bool { return metric.DuplicateScore >= 0.75 })
	highLabelRisk := metricObjectIDs(metrics, func(metric models.ObjectMetric) bool { return metric.LabelErrorProbability >= 0.55 })
	highUncertainty := metricObjectIDs(metrics, func(metric models.ObjectMetric) bool { return metric.UncertaintyScore >= 0.65 })

	recommendations := make([]models.Recommendation, 0, 7)
	add := func(kind, title, description string, ids []uuid.UUID) {
		if len(ids) == 0 {
			return
		}
		recommendations = append(recommendations, models.Recommendation{
			ID:                   uuid.New(),
			ProjectID:            project.ID,
			DatasetVersionID:     version.ID,
			Type:                 kind,
			Priority:             priorityForCount(len(ids)),
			Title:                title,
			Description:          description,
			AffectedObjectsCount: len(ids),
			ObjectIDs:            ids,
		})
	}
	add("missing_files", "Restore missing image files", "Some CSV rows reference files that were not found in images.zip.", missingFiles)
	add("missing_labels", "Add missing labels", "Objects without labels cannot safely improve the curated dataset.", missingLabels)
	add("unknown_classes", "Verify labels outside project classes", "Some labels do not match the class list configured for the project.", unknownClasses)
	add("duplicates", "Deduplicate repeated images", "Duplicate images reduce novelty and can bias the training set.", duplicates)
	add("label_risk", "Review high label-risk objects", "Objects with high label error probability should be checked by a human.", highLabelRisk)
	add("uncertainty", "Prioritize high-entropy objects", "High-uncertainty objects are strong active learning candidates.", highUncertainty)

	weakClasses := weakClassRecommendation(project, version, objects, counts, classes, labeledTotal)
	if weakClasses.AffectedObjectsCount > 0 {
		recommendations = append(recommendations, weakClasses)
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, models.Recommendation{
			ID:                   uuid.New(),
			ProjectID:            project.ID,
			DatasetVersionID:     version.ID,
			Type:                 "ready",
			Priority:             "low",
			Title:                "Dataset is ready for curated export",
			Description:          "No high-priority backend issues were detected in the current dataset version.",
			AffectedObjectsCount: 0,
		})
	}
	return recommendations
}

func weakClassRecommendation(project models.Project, version models.DatasetVersion, objects []models.DataObject, counts map[string]int, classes []string, labeledTotal int) models.Recommendation {
	if len(classes) == 0 || labeledTotal == 0 {
		return models.Recommendation{}
	}
	expected := float64(labeledTotal) / float64(len(classes))
	var weak []string
	for _, class := range classes {
		if expected-float64(counts[class]) >= math.Max(1, expected*0.25) {
			weak = append(weak, class)
		}
	}
	if len(weak) == 0 {
		return models.Recommendation{}
	}
	weakSet := map[string]struct{}{}
	for _, class := range weak {
		weakSet[class] = struct{}{}
	}
	ids := objectIDs(objects, func(object models.DataObject) bool {
		_, ok := weakSet[object.Label]
		return ok
	})
	return models.Recommendation{
		ID:                   uuid.New(),
		ProjectID:            project.ID,
		DatasetVersionID:     version.ID,
		Type:                 "class_deficit",
		Priority:             "high",
		Title:                "Collect data for weak classes",
		Description:          "Underrepresented classes: " + strings.Join(weak, ", ") + ".",
		AffectedObjectsCount: len(weak),
		ObjectIDs:            ids,
	}
}

func buildRoadmap(project models.Project, version models.DatasetVersion, recommendations []models.Recommendation, summary datasetSummary) []models.RoadmapItem {
	items := make([]models.RoadmapItem, 0, 4)
	add := func(title, description, actionType, impact string) {
		items = append(items, models.RoadmapItem{
			ID:               uuid.New(),
			ProjectID:        project.ID,
			DatasetVersionID: version.ID,
			Priority:         len(items) + 1,
			Title:            title,
			Description:      description,
			ActionType:       actionType,
			ExpectedImpact:   impact,
		})
	}

	if hasRecommendation(recommendations, "missing_files") || hasRecommendation(recommendations, "missing_labels") || hasRecommendation(recommendations, "unknown_classes") {
		add("Fix blocking dataset rows", "Restore missing files, add labels, and align labels with the project class list.", "fix_dataset_rows", "Raises quality score and makes export safer.")
	}
	if hasRecommendation(recommendations, "label_risk") || hasRecommendation(recommendations, "uncertainty") {
		add("Review active learning queue", "Inspect objects ranked by final score and correct labels where needed.", "review_queue", "Reduces uncertainty and label error probability.")
	}
	if hasRecommendation(recommendations, "class_deficit") || summary.ImbalanceIndex >= 0.25 {
		add("Collect weak-class samples", "Collect or annotate more examples for underrepresented classes.", "collect_data", "Improves balance and readiness.")
	}
	if hasRecommendation(recommendations, "duplicates") {
		add("Remove duplicate images", "Drop repeated visual samples before using the dataset for training.", "deduplicate", "Improves novelty and lowers bias.")
	}
	if len(items) == 0 {
		add("Export curated dataset", "The current dataset has no high-priority blocking issues.", "export_dataset", "Creates a clean Level 1 dataset artifact.")
	}
	return items
}

func hasRecommendation(recommendations []models.Recommendation, kind string) bool {
	for _, recommendation := range recommendations {
		if recommendation.Type == kind && recommendation.AffectedObjectsCount > 0 {
			return true
		}
	}
	return false
}

func objectIDs(objects []models.DataObject, keep func(models.DataObject) bool) []uuid.UUID {
	var ids []uuid.UUID
	for _, object := range objects {
		if keep(object) {
			ids = append(ids, object.ID)
		}
	}
	return ids
}

func metricObjectIDs(metrics []models.ObjectMetric, keep func(models.ObjectMetric) bool) []uuid.UUID {
	var ids []uuid.UUID
	for _, metric := range metrics {
		if keep(metric) {
			ids = append(ids, metric.ObjectID)
		}
	}
	return ids
}

func countObjects(objects []models.DataObject, keep func(models.DataObject) bool) int {
	count := 0
	for _, object := range objects {
		if keep(object) {
			count++
		}
	}
	return count
}

func countMetrics(metrics []models.ObjectMetric, keep func(models.ObjectMetric) bool) int {
	count := 0
	for _, metric := range metrics {
		if keep(metric) {
			count++
		}
	}
	return count
}

func average(total float64, count int) float64 {
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func averageMetric(metrics []models.ObjectMetric, value func(models.ObjectMetric) float64) float64 {
	var total float64
	for _, metric := range metrics {
		total += value(metric)
	}
	return average(total, len(metrics))
}

func priorityForCount(count int) string {
	if count >= 20 {
		return "high"
	}
	if count > 0 {
		return "medium"
	}
	return "low"
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
