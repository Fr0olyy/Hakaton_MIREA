package services

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"sort"
	"strconv"
	"strings"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func (s *Service) Upload(ctx context.Context, projectID uuid.UUID, datasetHeader, imagesHeader *multipart.FileHeader) (UploadResult, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return UploadResult{}, err
	}
	if datasetHeader == nil {
		return UploadResult{}, badRequest("dataset.csv is required")
	}

	datasetFile, err := datasetHeader.Open()
	if err != nil {
		return UploadResult{}, err
	}
	defer datasetFile.Close()

	datasetPath, err := s.storage.SaveMultipartFile(projectID, datasetFile, "dataset.csv")
	if err != nil {
		return UploadResult{}, err
	}

	if imagesHeader != nil {
		imagesFile, err := imagesHeader.Open()
		if err != nil {
			return UploadResult{}, err
		}
		defer imagesFile.Close()

		imagesPath, err := s.storage.SaveMultipartFile(projectID, imagesFile, "images.zip")
		if err != nil {
			return UploadResult{}, err
		}
		if err := s.storage.UnzipImages(projectID, imagesPath); err != nil {
			return UploadResult{}, badRequest("images.zip is invalid or unsafe")
		}
	}

	return s.uploadFromSavedDataset(ctx, project, datasetPath)
}

func (s *Service) uploadFromSavedDataset(ctx context.Context, project models.Project, datasetPath string) (UploadResult, error) {
	objects, invalid, err := s.parseCSV(project, datasetPath)
	if err != nil {
		return UploadResult{}, err
	}
	if err := s.storage.EnsureImageAliases(project.ID, objectFilePaths(objects)); err != nil {
		return UploadResult{}, err
	}
	if err := writeNormalizedDataset(datasetPath, objects); err != nil {
		return UploadResult{}, err
	}

	versionNumber, err := s.repo.NextDatasetVersionNumber(ctx, project.ID)
	if err != nil {
		return UploadResult{}, err
	}
	version := models.DatasetVersion{
		ID:             uuid.New(),
		ProjectID:      project.ID,
		VersionName:    "v" + strconv.Itoa(versionNumber),
		Status:         "uploaded",
		ObjectsCount:   len(objects),
		ReadinessScore: 0,
	}
	version, err = s.repo.CreateDatasetWithObjects(ctx, version, objects)
	if err != nil {
		return UploadResult{}, err
	}
	return UploadResult{DatasetVersion: version, ObjectsCount: len(objects), InvalidObjects: invalid}, nil
}

func objectFilePaths(objects []models.DataObject) []string {
	paths := make([]string, 0, len(objects))
	for _, object := range objects {
		if object.FilePath != "" {
			paths = append(paths, object.FilePath)
		}
	}
	return paths
}

func writeNormalizedDataset(path string, objects []models.DataObject) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	probabilityClasses := normalizedProbabilityClasses(objects)
	headers := []string{"id", "file_path", "label", "predicted_label", "confidence", "split", "source", "annotator"}
	for _, className := range probabilityClasses {
		headers = append(headers, "prob_"+className)
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, object := range objects {
		record := []string{
			object.ExternalID,
			object.FilePath,
			object.Label,
			object.PredictedLabel,
			strconv.FormatFloat(object.Confidence, 'f', -1, 64),
			object.Split,
			object.Source,
			object.Annotator,
		}
		probabilities, _ := object.Metadata["probabilities"].(map[string]float64)
		for _, className := range probabilityClasses {
			record = append(record, strconv.FormatFloat(probabilities[className], 'f', -1, 64))
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return writer.Error()
}

func normalizedProbabilityClasses(objects []models.DataObject) []string {
	seen := map[string]struct{}{}
	var classes []string
	for _, object := range objects {
		probabilities, ok := object.Metadata["probabilities"].(map[string]float64)
		if !ok {
			continue
		}
		for className := range probabilities {
			if _, exists := seen[className]; exists {
				continue
			}
			seen[className] = struct{}{}
			classes = append(classes, className)
		}
	}
	sort.Strings(classes)
	return classes
}

func (s *Service) parseCSV(project models.Project, datasetPath string) ([]models.DataObject, int, error) {
	file, err := os.Open(datasetPath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return nil, 0, err
	}
	indexes := headerIndexes(headers)
	if isImageClassificationProject(project) {
		if _, ok := indexes["file_path"]; !ok {
			return nil, 0, badRequest("image classification csv must contain file_path column")
		}
	}

	classSet := map[string]struct{}{}
	for _, class := range project.Classes {
		classSet[class] = struct{}{}
	}

	var objects []models.DataObject
	invalid := 0
	rowNumber := 1
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		rowNumber++
		if err != nil {
			return nil, invalid, badRequest("csv contains malformed rows")
		}

		object := models.DataObject{
			ID:       uuid.New(),
			Status:   statusOK,
			Metadata: map[string]any{"row_number": rowNumber},
		}
		object.ExternalID = firstNonEmpty(csvValue(record, indexes, "external_id"), csvValue(record, indexes, "id"))
		object.FilePath = csvValue(record, indexes, "file_path")
		object.TextContent = csvValue(record, indexes, "text_content")
		object.Label = firstNonEmpty(csvValue(record, indexes, "label"), csvValue(record, indexes, "target"), csvValue(record, indexes, "target_label"), csvValue(record, indexes, "class"))
		object.PredictedLabel = csvValue(record, indexes, "predicted_label")
		object.Split = csvValue(record, indexes, "split")
		object.Source = csvValue(record, indexes, "source")
		object.Annotator = csvValue(record, indexes, "annotator")
		object.Confidence = parseOptionalFloat(csvValue(record, indexes, "confidence"))

		probabilities := probabilityColumns(record, headers)
		if len(probabilities) > 0 {
			object.Metadata["probabilities"] = probabilities
		}
		for name, idx := range indexes {
			if isKnownColumn(name) || strings.HasPrefix(name, "prob_") || idx >= len(record) {
				continue
			}
			object.Metadata[name] = strings.TrimSpace(record[idx])
		}

		validationErrors := s.validateObject(project, classSet, &object)
		if len(validationErrors) > 0 {
			object.Metadata["validation_errors"] = validationErrors
		}
		if object.Status != statusOK {
			invalid++
		}
		objects = append(objects, object)
	}
	if len(objects) == 0 {
		return nil, invalid, badRequest("csv contains no data rows")
	}
	return objects, invalid, nil
}

func (s *Service) validateObject(project models.Project, classSet map[string]struct{}, object *models.DataObject) []string {
	var issues []string
	setStatus := func(status string) {
		if statusPriority(status) > statusPriority(object.Status) {
			object.Status = status
		}
	}

	if isImageClassificationProject(project) {
		switch {
		case strings.TrimSpace(object.FilePath) == "":
			issues = append(issues, "file_path is required")
			setStatus(statusInvalid)
		case !s.storage.IsSafeRelativePath(object.FilePath):
			issues = append(issues, "file_path must be a safe relative path")
			setStatus(statusInvalid)
		case !s.storage.HasSupportedImageExtension(object.FilePath):
			issues = append(issues, "file_path has unsupported image extension")
			setStatus(statusInvalid)
		case !s.storage.ImageExists(project.ID, object.FilePath):
			issues = append(issues, "file_path is missing in images.zip")
			setStatus(statusMissingFile)
		case !s.storage.ImageReadable(project.ID, object.FilePath):
			issues = append(issues, "image file is not readable")
			setStatus(statusInvalid)
		}
	}

	if strings.TrimSpace(object.Label) == "" && (isImageClassificationProject(project) || len(classSet) > 0) {
		issues = append(issues, "label is required")
		setStatus(statusMissingLabel)
	} else if len(classSet) > 0 {
		if _, ok := classSet[object.Label]; !ok {
			issues = append(issues, "label is not present in project classes")
			setStatus(statusUnknownClass)
		}
	}

	if isImageLikeProject(project) && object.FilePath == "" && object.TextContent == "" {
		issues = append(issues, "object has no file_path or text_content")
		setStatus(statusInvalid)
	}
	return issues
}

func headerIndexes(headers []string) map[string]int {
	indexes := map[string]int{}
	for i, header := range headers {
		indexes[strings.ToLower(strings.TrimSpace(header))] = i
	}
	return indexes
}

func csvValue(record []string, indexes map[string]int, name string) string {
	index, ok := indexes[name]
	if !ok || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func probabilityColumns(record, headers []string) map[string]float64 {
	probabilities := map[string]float64{}
	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		if !strings.HasPrefix(header, "prob_") || i >= len(record) {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(record[i]), 64)
		if err != nil {
			continue
		}
		class := strings.TrimSpace(strings.TrimPrefix(header, "prob_"))
		if class != "" {
			probabilities[class] = value
		}
	}
	return probabilities
}

func isKnownColumn(name string) bool {
	switch name {
	case "id", "external_id", "file_path", "text_content", "label", "target", "target_label", "class", "predicted_label", "confidence", "split", "source", "annotator":
		return true
	default:
		return false
	}
}

func parseOptionalFloat(value string) float64 {
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func statusPriority(status string) int {
	switch status {
	case statusInvalid:
		return 5
	case statusMissingFile:
		return 4
	case statusMissingLabel:
		return 3
	case statusUnknownClass:
		return 2
	case statusOK:
		return 1
	default:
		return 0
	}
}
