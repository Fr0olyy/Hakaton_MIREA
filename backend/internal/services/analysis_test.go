package services

import (
	"path/filepath"
	"testing"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func TestAnalyzeLocalCalculatesRequiredMetrics(t *testing.T) {
	service, project := testService(t, []string{"cat", "dog"})
	writeTestImage(t, service.storage.ImagesDir(project.ID), "a.png")
	writeTestImage(t, service.storage.ImagesDir(project.ID), "b.png")

	objects := []models.DataObject{
		{
			ID:       uuid.New(),
			FilePath: "a.png",
			Label:    "cat",
			Status:   statusOK,
			Metadata: map[string]any{"probabilities": map[string]float64{"cat": 0.55, "dog": 0.45}},
		},
		{
			ID:       uuid.New(),
			FilePath: "a.png",
			Label:    "cat",
			Status:   statusOK,
			Metadata: map[string]any{},
		},
		{
			ID:             uuid.New(),
			FilePath:       "b.png",
			Label:          "dog",
			PredictedLabel: "cat",
			Confidence:     0.88,
			Status:         statusOK,
			Metadata:       map[string]any{},
		},
	}

	result := service.analyzeLocal(project, models.DatasetVersion{ID: uuid.New(), ProjectID: project.ID}, objects)
	if len(result.metrics) != len(objects) {
		t.Fatalf("metrics len = %d, want %d", len(result.metrics), len(objects))
	}
	if result.readiness <= 0 || result.readiness > 100 {
		t.Fatalf("readiness = %.2f, want (0,100]", result.readiness)
	}
	if result.metrics[0].Entropy <= 0 {
		t.Fatalf("entropy = %.4f, want > 0", result.metrics[0].Entropy)
	}
	if result.metrics[0].LabelErrorProbability <= 0.4 {
		t.Fatalf("label risk = %.4f, want > 0.4", result.metrics[0].LabelErrorProbability)
	}
	if result.metrics[0].DuplicateScore != 1 || result.metrics[1].DuplicateScore != 1 {
		t.Fatalf("duplicate scores = %.2f %.2f, want both 1", result.metrics[0].DuplicateScore, result.metrics[1].DuplicateScore)
	}
	if result.summary.AnalysisSource != analysisSourceBackendLocal {
		t.Fatalf("analysis source = %q", result.summary.AnalysisSource)
	}
	if len(result.recommendations) == 0 || len(result.roadmap) == 0 {
		t.Fatalf("recommendations/roadmap should be generated")
	}
}

func TestCollectOutputFilesReturnsEmptyWhenNoFiles(t *testing.T) {
	service, project := testService(t, []string{"cat", "dog"})
	files := service.collectOutputFiles(project.ID, uuid.New())
	if len(files) != 0 {
		t.Fatalf("expected empty output files, got %d", len(files))
	}
}

func TestCollectOutputFilesReturnsExistingFiles(t *testing.T) {
	service, project := testService(t, []string{"cat", "dog"})
	versionID := uuid.New()
	analysisDir := filepath.Join(service.storage.AnalysisDir(project.ID), versionID.String())
	writeFile(t, filepath.Join(analysisDir, "results.csv"), "a,b\n1,2")
	writeFile(t, filepath.Join(analysisDir, "agent_context.json"), "{}")

	files := service.collectOutputFiles(project.ID, versionID)
	if len(files) != 2 {
		t.Fatalf("expected 2 output files, got %d", len(files))
	}
	if files["results.csv"] == "" {
		t.Fatalf("missing results.csv in output files")
	}
	if files["agent_context.json"] == "" {
		t.Fatalf("missing agent_context.json in output files")
	}
}

func TestImbalanceIndexBounds(t *testing.T) {
	balanced := imbalanceIndex(map[string]float64{"cat": 0.5, "dog": 0.5}, []string{"cat", "dog"}, 2)
	if balanced != 0 {
		t.Fatalf("balanced imbalance = %.4f, want 0", balanced)
	}
	skewed := imbalanceIndex(map[string]float64{"cat": 1, "dog": 0}, []string{"cat", "dog"}, 2)
	if skewed != 1 {
		t.Fatalf("skewed imbalance = %.4f, want 1", skewed)
	}
}
