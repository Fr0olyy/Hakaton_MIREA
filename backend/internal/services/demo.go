package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

type DemoDataset struct {
	Modality    string `json:"modality"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type DemoProjectResult struct {
	Project      models.Project     `json:"project"`
	UploadResult UploadResult       `json:"upload_result"`
	AnalysisJob  models.AnalysisJob `json:"analysis_job"`
}

func (s *Service) DemoDatasets() []DemoDataset {
	return []DemoDataset{
		{Modality: "image_classification", Label: "Image Classification", Description: "Animals-10 image classification demo with dataset.csv and images.zip."},
		{Modality: "tabular_classification", Label: "Tabular Classification", Description: "Customer churn style tabular classification demo with missing values and outliers."},
		{Modality: "image_detection_yolo", Label: "YOLO Detection", Description: "YOLO label quality demo with valid, invalid, and tiny bounding boxes."},
	}
}

func (s *Service) CreateDemoProject(ctx context.Context, modality string, creatorID uuid.UUID) (DemoProjectResult, error) {
	demo, err := demoSpec(modality)
	if err != nil {
		return DemoProjectResult{}, err
	}
	project, err := s.CreateProject(ctx, CreateProjectRequest{
		Name:     demo.name,
		Modality: demo.modality,
		TaskType: demo.taskType,
		Classes:  demo.classes,
	}, creatorID)
	if err != nil {
		return DemoProjectResult{}, err
	}

	datasetPath := s.storage.DatasetPath(project.ID)
	if err := copyFile(datasetPath, filepath.Join(demoDataRoot(), demo.dataset)); err != nil {
		return DemoProjectResult{}, err
	}
	if demo.imagesZip != "" {
		imagesZipPath := filepath.Join(s.storage.RawDir(project.ID), "images.zip")
		if err := copyFile(imagesZipPath, filepath.Join(demoDataRoot(), demo.imagesZip)); err != nil {
			return DemoProjectResult{}, err
		}
		if err := s.storage.UnzipImages(project.ID, imagesZipPath); err != nil {
			return DemoProjectResult{}, err
		}
	}

	uploadResult, err := s.uploadFromSavedDataset(ctx, project, datasetPath)
	if err != nil {
		return DemoProjectResult{}, err
	}
	job, err := s.Analyze(ctx, project.ID)
	if err != nil {
		return DemoProjectResult{}, err
	}
	return DemoProjectResult{Project: project, UploadResult: uploadResult, AnalysisJob: job}, nil
}

type demoDatasetSpec struct {
	name      string
	modality  string
	taskType  string
	classes   []string
	dataset   string
	imagesZip string
}

func demoSpec(modality string) (demoDatasetSpec, error) {
	switch modality {
	case "image", "image_classification":
		return demoDatasetSpec{
			name:      "Demo Image Classification",
			modality:  "image_classification",
			taskType:  "classification",
			classes:   []string{"butterfly", "cat", "chicken", "cow", "dog", "elephant", "horse", "sheep", "spider", "squirrel"},
			dataset:   "demo/dataset.csv",
			imagesZip: "demo/images.zip",
		}, nil
	case "tabular_classification":
		return demoDatasetSpec{
			name:     "Demo Tabular Classification",
			modality: "tabular_classification",
			taskType: "classification",
			classes:  []string{"retain", "churn"},
			dataset:  "demo_tabular/tabular.csv",
		}, nil
	case "image_detection_yolo":
		return demoDatasetSpec{
			name:     "Demo YOLO Detection",
			modality: "image_detection_yolo",
			taskType: "detection",
			classes:  []string{"person", "vehicle", "sign"},
			dataset:  "demo_yolo/dataset.csv",
		}, nil
	default:
		return demoDatasetSpec{}, badRequest("unknown demo dataset: " + modality)
	}
}

func demoDataRoot() string {
	if value := os.Getenv("DEMO_DATA_DIR"); value != "" {
		return value
	}
	return "/app/ml-data"
}

func copyFile(target, source string) error {
	src, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open demo file %s: %w", source, err)
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	dst, err := os.Create(target)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
