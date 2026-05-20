package services

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"hakaton/backend/internal/models"
	filestorage "hakaton/backend/internal/storage"

	"github.com/google/uuid"
)

func TestParseCSVSupportsExtendedColumns(t *testing.T) {
	service, project := testService(t, []string{"cat", "dog"})
	writeTestImage(t, service.storage.ImagesDir(project.ID), "images/cat.png")

	csvPath := filepath.Join(t.TempDir(), "dataset.csv")
	writeFile(t, csvPath, "id,file_path,label,predicted_label,confidence,split,source,annotator,prob_cat,prob_dog,custom\n1,images/cat.png,cat,cat,0.91,train,camera_a,ann_1,0.82,0.18,keep-me\n")

	objects, invalid, err := service.parseCSV(project, csvPath)
	if err != nil {
		t.Fatalf("parseCSV returned error: %v", err)
	}
	if invalid != 0 {
		t.Fatalf("invalid count = %d, want 0", invalid)
	}
	if len(objects) != 1 {
		t.Fatalf("objects len = %d, want 1", len(objects))
	}
	object := objects[0]
	if object.ExternalID != "1" || object.Status != statusOK || object.Confidence != 0.91 {
		t.Fatalf("unexpected object: %+v", object)
	}
	if object.Metadata["custom"] != "keep-me" {
		t.Fatalf("custom metadata was not preserved: %+v", object.Metadata)
	}
	probabilities, ok := object.Metadata["probabilities"].(map[string]float64)
	if !ok {
		t.Fatalf("probabilities metadata type = %T", object.Metadata["probabilities"])
	}
	if probabilities["cat"] != 0.82 || probabilities["dog"] != 0.18 {
		t.Fatalf("unexpected probabilities: %+v", probabilities)
	}
}

func TestParseCSVAssignsValidationStatuses(t *testing.T) {
	service, project := testService(t, []string{"cat", "dog"})
	writeTestImage(t, service.storage.ImagesDir(project.ID), "cat.png")

	csvPath := filepath.Join(t.TempDir(), "dataset.csv")
	writeFile(t, csvPath, "id,file_path,label\n1,missing.png,cat\n2,cat.png,\n3,cat.png,fox\n4,,cat\n")

	objects, invalid, err := service.parseCSV(project, csvPath)
	if err != nil {
		t.Fatalf("parseCSV returned error: %v", err)
	}
	if invalid != 4 {
		t.Fatalf("invalid count = %d, want 4", invalid)
	}
	got := []string{objects[0].Status, objects[1].Status, objects[2].Status, objects[3].Status}
	want := []string{statusMissingFile, statusMissingLabel, statusUnknownClass, statusInvalid}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("status[%d] = %q, want %q (all statuses: %+v)", i, got[i], want[i], got)
		}
	}
}

func TestParseCSVRejectsMissingFilePathHeaderForImageProjects(t *testing.T) {
	service, project := testService(t, []string{"cat"})
	csvPath := filepath.Join(t.TempDir(), "dataset.csv")
	writeFile(t, csvPath, "id,label\n1,cat\n")

	_, _, err := service.parseCSV(project, csvPath)
	if !IsBadRequest(err) {
		t.Fatalf("error = %v, want bad request", err)
	}
}

func testService(t *testing.T, classes []string) (*Service, models.Project) {
	t.Helper()
	storage := filestorage.New(t.TempDir())
	project := models.Project{
		ID:       uuid.New(),
		Name:     "Test",
		Modality: "image",
		TaskType: "classification",
		Classes:  classes,
	}
	if err := storage.EnsureProjectDirs(project.ID); err != nil {
		t.Fatalf("EnsureProjectDirs: %v", err)
	}
	return &Service{storage: storage}, project
}

func writeTestImage(t *testing.T, root, name string) {
	t.Helper()
	bytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADUlEQVR4nGJgYGD4DwABBAEAgh7e1wAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir image dir: %v", err)
	}
	writeBytes(t, path, bytes)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	writeBytes(t, path, []byte(content))
}

func writeBytes(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
