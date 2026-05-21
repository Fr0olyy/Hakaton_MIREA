package services

import (
	"path/filepath"
	"testing"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func TestMetricsFromMLResultsMapsRowsAndStatuses(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	objects := []models.DataObject{
		{
			ID:         firstID,
			ExternalID: "1",
			FilePath:   "cat.png",
			Label:      "cat",
			Status:     statusOK,
		},
		{
			ID:         secondID,
			ExternalID: "2",
			FilePath:   "dog.png",
			Label:      "dog",
			Status:     statusOK,
		},
	}

	path := filepath.Join(t.TempDir(), "results.csv")
	writeFile(t, path, "id,file_path,status,entropy,uncertainty_score,label_error_probability,duplicate_score,class_deficit_score,quality_score,object_utility_score,final_score,reasons,recommendation,prob_cat,prob_dog\n"+
		"1,cat.png,ok,0.1,0.2,0.3,0,0.1,1,0.22,0.22,,keep,0.8,0.2\n"+
		"2,dog.png,suspected_label_error,0.5,0.4,0.9,0,0.2,0.9,0.42,0.42,model_disagreement;suspected_label_error,recheck_label,0.3,0.7\n")

	metrics, statuses, err := metricsFromMLResults(path, objects)
	if err != nil {
		t.Fatalf("metricsFromMLResults: %v", err)
	}
	if len(metrics) != 2 {
		t.Fatalf("metrics len = %d, want 2", len(metrics))
	}
	if metrics[0].ObjectID != firstID || metrics[1].ObjectID != secondID {
		t.Fatalf("metrics mapped to wrong objects: %+v", metrics)
	}
	if statuses[firstID] != "ok" || statuses[secondID] != "suspected_label_error" {
		t.Fatalf("statuses = %+v", statuses)
	}
	if source := metrics[1].Probabilities["analysis_source"]; source != analysisSourceMLService {
		t.Fatalf("analysis source = %v", source)
	}
	if metrics[1].Probabilities["dog"] != 0.7 {
		t.Fatalf("probabilities not preserved: %+v", metrics[1].Probabilities)
	}
	if len(metrics[1].Reasons) != 2 {
		t.Fatalf("reasons = %+v, want two values", metrics[1].Reasons)
	}
}
