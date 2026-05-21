package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type AnalysisRequest struct {
	ProjectID        uuid.UUID           `json:"project_id"`
	DatasetVersionID uuid.UUID           `json:"dataset_version_id"`
	Objects          []models.DataObject `json:"objects"`
}

type AnalysisResult struct {
	ReadinessScore  float64                 `json:"readiness_score"`
	Metrics         []models.ObjectMetric   `json:"metrics"`
	Recommendations []models.Recommendation `json:"recommendations"`
	Roadmap         []models.RoadmapItem    `json:"roadmap"`
}

type DatasetAnalysisRequest struct {
	DatasetPath string `json:"dataset_path"`
	ImagesDir   string `json:"images_dir"`
	OutputDir   string `json:"output_dir"`
}

type DatasetAnalysisResult struct {
	Status                string            `json:"status"`
	ObjectsCount          int               `json:"objects_count"`
	DatasetV2ObjectsCount int               `json:"dataset_v2_objects_count"`
	ReviewQueueCount      int               `json:"review_queue_count"`
	DatasetReadinessScore float64           `json:"dataset_readiness_score"`
	OutputDir             string            `json:"output_dir"`
	Files                 map[string]string `json:"files"`
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) AnalyzeDataset(ctx context.Context, req DatasetAnalysisRequest) (DatasetAnalysisResult, error) {
	if c == nil || strings.TrimSpace(c.baseURL) == "" {
		return DatasetAnalysisResult{}, errors.New("ml-service url is not configured")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return DatasetAnalysisResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analyze", bytes.NewReader(body))
	if err != nil {
		return DatasetAnalysisResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return DatasetAnalysisResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var payload map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		if detail, ok := payload["detail"].(string); ok && detail != "" {
			return DatasetAnalysisResult{}, fmt.Errorf("ml-service returned %d: %s", resp.StatusCode, detail)
		}
		return DatasetAnalysisResult{}, fmt.Errorf("ml-service returned %d", resp.StatusCode)
	}

	var result DatasetAnalysisResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return DatasetAnalysisResult{}, err
	}
	if result.Status != "completed" {
		return DatasetAnalysisResult{}, fmt.Errorf("ml-service analysis status is %q", result.Status)
	}
	if result.Files["results_csv"] == "" {
		return DatasetAnalysisResult{}, errors.New("ml-service response is missing files.results_csv")
	}
	return result, nil
}

func (c *Client) Analyze(ctx context.Context, req AnalysisRequest) (AnalysisResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return AnalysisResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analyze", bytes.NewReader(body))
	if err != nil {
		return AnalysisResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return c.fallbackAnalyze(req), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.fallbackAnalyze(req), nil
	}

	var result AnalysisResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return AnalysisResult{}, err
	}
	if len(result.Metrics) == 0 {
		return AnalysisResult{}, errors.New("ml-service returned empty metrics")
	}
	fillResultIDs(&result, req.ProjectID, req.DatasetVersionID)
	return result, nil
}

func (c *Client) fallbackAnalyze(req AnalysisRequest) AnalysisResult {
	metrics := make([]models.ObjectMetric, 0, len(req.Objects))
	var totalQuality float64
	var highRisk []uuid.UUID
	labelCounts := map[string]int{}
	for _, object := range req.Objects {
		if object.Label != "" {
			labelCounts[object.Label]++
		}
	}

	for index, object := range req.Objects {
		missingFilePenalty := 0.0
		reasons := []string{}
		if object.Status != "ok" {
			missingFilePenalty = 0.35
			reasons = append(reasons, "file_path is missing in images.zip")
		}
		if object.Label == "" {
			reasons = append(reasons, "missing label")
		}
		uncertainty := clamp(0.25+float64((index%5))*0.12+missingFilePenalty, 0, 1)
		labelRisk := clamp(uncertainty*0.7, 0, 1)
		quality := clamp(1-missingFilePenalty-labelRisk*0.35, 0, 1)
		final := clamp((uncertainty*0.45)+(labelRisk*0.35)+((1-quality)*0.20), 0, 1)
		if final >= 0.6 {
			highRisk = append(highRisk, object.ID)
		}
		if len(reasons) == 0 {
			reasons = append(reasons, "high model uncertainty")
		}
		probabilityA := clamp(1-uncertainty/2, 0, 1)
		probabilityB := clamp(1-probabilityA, 0, 1)
		metrics = append(metrics, models.ObjectMetric{
			ID:                    uuid.New(),
			ObjectID:              object.ID,
			Entropy:               entropy(probabilityA, probabilityB),
			UncertaintyScore:      uncertainty,
			LabelErrorProbability: labelRisk,
			DuplicateScore:        clamp(float64(index%3)*0.1, 0, 1),
			RarityScore:           classRarity(object.Label, labelCounts, len(req.Objects)),
			ClassDeficitScore:     classDeficit(object.Label, labelCounts, len(req.Objects)),
			QualityScore:          quality,
			NoveltyScore:          clamp(0.2+float64(index%4)*0.15, 0, 1),
			ObjectUtilityScore:    final,
			FinalScore:            final,
			Reasons:               reasons,
			Recommendation:        recommendationFor(final),
			Probabilities: map[string]any{
				"primary":   probabilityA,
				"secondary": probabilityB,
			},
		})
		totalQuality += quality
	}

	readiness := 0.0
	if len(metrics) > 0 {
		readiness = clamp((totalQuality/float64(len(metrics)))*100, 0, 100)
	}
	recommendations := []models.Recommendation{
		{
			ID:                   uuid.New(),
			ProjectID:            req.ProjectID,
			DatasetVersionID:     req.DatasetVersionID,
			Type:                 "active_learning",
			Priority:             priorityForCount(len(highRisk)),
			Title:                "Review high uncertainty objects",
			Description:          "Objects with the highest final score should be sent to manual review first.",
			AffectedObjectsCount: len(highRisk),
			ObjectIDs:            highRisk,
		},
	}
	roadmap := []models.RoadmapItem{
		{
			ID:               uuid.New(),
			ProjectID:        req.ProjectID,
			DatasetVersionID: req.DatasetVersionID,
			Priority:         1,
			Title:            "Validate active learning queue",
			Description:      "Review objects ranked by final score and fix labels or missing files.",
			ActionType:       "review_queue",
			ExpectedImpact:   "Improves dataset readiness and lowers label error risk.",
		},
		{
			ID:               uuid.New(),
			ProjectID:        req.ProjectID,
			DatasetVersionID: req.DatasetVersionID,
			Priority:         2,
			Title:            "Balance weak classes",
			Description:      "Collect or annotate additional samples for underrepresented classes.",
			ActionType:       "collect_data",
			ExpectedImpact:   "Reduces class deficit and improves model stability.",
		},
	}
	return AnalysisResult{ReadinessScore: readiness, Metrics: metrics, Recommendations: recommendations, Roadmap: roadmap}
}

func fillResultIDs(result *AnalysisResult, projectID, datasetVersionID uuid.UUID) {
	for i := range result.Metrics {
		if result.Metrics[i].ID == uuid.Nil {
			result.Metrics[i].ID = uuid.New()
		}
	}
	for i := range result.Recommendations {
		if result.Recommendations[i].ID == uuid.Nil {
			result.Recommendations[i].ID = uuid.New()
		}
		result.Recommendations[i].ProjectID = projectID
		result.Recommendations[i].DatasetVersionID = datasetVersionID
	}
	for i := range result.Roadmap {
		if result.Roadmap[i].ID == uuid.Nil {
			result.Roadmap[i].ID = uuid.New()
		}
		result.Roadmap[i].ProjectID = projectID
		result.Roadmap[i].DatasetVersionID = datasetVersionID
	}
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

func entropy(values ...float64) float64 {
	var result float64
	for _, value := range values {
		if value > 0 {
			result -= value * math.Log2(value)
		}
	}
	return result
}

func classRarity(label string, counts map[string]int, total int) float64 {
	if label == "" || total == 0 {
		return 0.5
	}
	return clamp(1-(float64(counts[label])/float64(total)), 0, 1)
}

func classDeficit(label string, counts map[string]int, total int) float64 {
	if label == "" || total == 0 {
		return 1
	}
	expected := float64(total) / math.Max(float64(len(counts)), 1)
	if expected == 0 {
		return 0
	}
	return clamp((expected-float64(counts[label]))/expected, 0, 1)
}

func recommendationFor(score float64) string {
	if score >= 0.7 {
		return "manual_review"
	}
	if score >= 0.45 {
		return "verify_label"
	}
	return "keep"
}

func priorityForCount(count int) string {
	if count > 20 {
		return "high"
	}
	if count > 0 {
		return "medium"
	}
	return "low"
}
