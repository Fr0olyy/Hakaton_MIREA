package services

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func TestWriteDatasetCSVExcludesMissingAndInvalidRows(t *testing.T) {
	keepID := uuid.New()
	objects := []models.DataObject{
		{ID: keepID, ExternalID: "1", FilePath: "cat.png", Label: "cat", Status: statusOK},
		{ID: uuid.New(), ExternalID: "2", FilePath: "missing.png", Label: "cat", Status: statusMissingFile},
		{ID: uuid.New(), ExternalID: "3", FilePath: "", Label: "cat", Status: statusInvalid},
		{ID: uuid.New(), ExternalID: "4", FilePath: "noisy.png", Label: "cat", Status: "suspected_label_error"},
	}
	metrics := map[uuid.UUID]models.ObjectMetric{
		keepID: {FinalScore: 0.52, Recommendation: "verify_label"},
	}
	path := filepath.Join(t.TempDir(), "dataset_v2.csv")
	if err := writeDatasetCSV(path, objects, metrics); err != nil {
		t.Fatalf("writeDatasetCSV: %v", err)
	}
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	content := string(contentBytes)
	if !strings.Contains(content, "review_recommendation") || !strings.Contains(content, "verify_label") {
		t.Fatalf("csv missing metric columns/content:\n%s", content)
	}
	if strings.Contains(content, "missing.png") {
		t.Fatalf("missing file row should not be exported:\n%s", content)
	}
	if strings.Contains(content, "noisy.png") {
		t.Fatalf("suspected label error row should not be exported:\n%s", content)
	}
}

func TestReviewQueueIncludesMLRecommendationsBelowBackendThreshold(t *testing.T) {
	objectID := uuid.New()
	queue := reviewQueueFrom(
		[]models.DataObject{
			{ID: objectID, ExternalID: "1", FilePath: "cat.png", Label: "cat", Status: "suspected_label_error"},
		},
		[]models.ObjectMetric{
			{ObjectID: objectID, FinalScore: 0.41, Recommendation: "recheck_label"},
		},
	)
	if len(queue) != 1 {
		t.Fatalf("queue len = %d, want 1", len(queue))
	}
}

func TestZipFilesContainsExpectedArtifacts(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "dataset_v2.csv"), "id\n1\n")
	writeFile(t, filepath.Join(dir, "dataset_report.json"), "{}\n")
	writeFile(t, filepath.Join(dir, "cat.png"), "fake")

	zipPath := filepath.Join(dir, "export.zip")
	err := zipFiles(zipPath, []zipSource{
		{Name: "dataset_v2.csv", Path: filepath.Join(dir, "dataset_v2.csv")},
		{Name: "dataset_report.json", Path: filepath.Join(dir, "dataset_report.json")},
		{Name: "images/cat.png", Path: filepath.Join(dir, "cat.png")},
	})
	if err != nil {
		t.Fatalf("zipFiles: %v", err)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer reader.Close()
	names := map[string]bool{}
	for _, file := range reader.File {
		names[file.Name] = true
	}
	for _, name := range []string{"dataset_v2.csv", "dataset_report.json", "images/cat.png"} {
		if !names[name] {
			t.Fatalf("zip missing %s; got %+v", name, names)
		}
	}
}
