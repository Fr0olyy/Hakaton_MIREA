package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *Repository) CreateProject(ctx context.Context, p models.Project) (models.Project, error) {
	classes, err := json.Marshal(p.Classes)
	if err != nil {
		return p, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO projects (id, name, modality, task_type, classes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`, p.ID, p.Name, p.Modality, p.TaskType, classes).Scan(&p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) ListProjects(ctx context.Context) ([]models.Project, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, modality, task_type, COALESCE(classes, '[]'::jsonb), created_at, updated_at
		FROM projects ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (r *Repository) GetProject(ctx context.Context, id uuid.UUID) (models.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, modality, task_type, COALESCE(classes, '[]'::jsonb), created_at, updated_at
		FROM projects WHERE id = $1
	`, id)
	project, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return project, ErrNotFound
	}
	return project, err
}

func (r *Repository) CreateDatasetWithObjects(ctx context.Context, version models.DatasetVersion, objects []models.DataObject) (models.DatasetVersion, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return version, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO dataset_versions (id, project_id, version_name, status, objects_count, readiness_score)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`, version.ID, version.ProjectID, version.VersionName, version.Status, len(objects), version.ReadinessScore).Scan(&version.CreatedAt)
	if err != nil {
		return version, err
	}

	for _, object := range objects {
		metadata, err := json.Marshal(object.Metadata)
		if err != nil {
			return version, err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO data_objects (
				id, dataset_version_id, external_id, file_path, text_content, label,
				predicted_label, confidence, split, source, annotator, metadata, status
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, object.ID, version.ID, object.ExternalID, object.FilePath, object.TextContent, object.Label,
			object.PredictedLabel, object.Confidence, object.Split, object.Source, object.Annotator, metadata, object.Status)
		if err != nil {
			return version, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return version, err
	}
	version.ObjectsCount = len(objects)
	return version, nil
}

func (r *Repository) NextDatasetVersionNumber(ctx context.Context, projectID uuid.UUID) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM dataset_versions WHERE project_id = $1
	`, projectID).Scan(&count); err != nil {
		return 0, err
	}
	return count + 1, nil
}

func (r *Repository) LatestDatasetVersion(ctx context.Context, projectID uuid.UUID) (models.DatasetVersion, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, project_id, version_name, status, objects_count, COALESCE(readiness_score, 0), created_at
		FROM dataset_versions
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, projectID)
	version, err := scanDatasetVersion(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return version, ErrNotFound
	}
	return version, err
}

func (r *Repository) ListObjects(ctx context.Context, datasetVersionID uuid.UUID) ([]models.DataObject, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, dataset_version_id, COALESCE(external_id, ''), COALESCE(file_path, ''),
			COALESCE(text_content, ''), COALESCE(label, ''), COALESCE(predicted_label, ''),
			COALESCE(confidence, 0), COALESCE(split, ''), COALESCE(source, ''),
			COALESCE(annotator, ''), COALESCE(metadata, '{}'::jsonb), COALESCE(status, 'ok'), created_at
		FROM data_objects
		WHERE dataset_version_id = $1
		ORDER BY created_at ASC
	`, datasetVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objects []models.DataObject
	for rows.Next() {
		object, err := scanDataObject(rows)
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return objects, rows.Err()
}

func (r *Repository) GetObject(ctx context.Context, projectID, objectID uuid.UUID) (models.DataObject, error) {
	row := r.db.QueryRow(ctx, `
		SELECT o.id, o.dataset_version_id, COALESCE(o.external_id, ''), COALESCE(o.file_path, ''),
			COALESCE(o.text_content, ''), COALESCE(o.label, ''), COALESCE(o.predicted_label, ''),
			COALESCE(o.confidence, 0), COALESCE(o.split, ''), COALESCE(o.source, ''),
			COALESCE(o.annotator, ''), COALESCE(o.metadata, '{}'::jsonb), COALESCE(o.status, 'ok'), o.created_at
		FROM data_objects o
		JOIN dataset_versions dv ON dv.id = o.dataset_version_id
		WHERE dv.project_id = $1 AND o.id = $2
	`, projectID, objectID)
	object, err := scanDataObject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return object, ErrNotFound
	}
	if err != nil {
		return object, err
	}

	metric, err := r.GetMetricByObject(ctx, object.ID)
	if err == nil {
		object.Metrics = &metric
	}
	return object, nil
}

func (r *Repository) CreateAnalysisJob(ctx context.Context, job models.AnalysisJob) (models.AnalysisJob, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO analysis_jobs (id, project_id, dataset_version_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING started_at
	`, job.ID, job.ProjectID, job.DatasetVersionID, job.Status).Scan(&job.StartedAt)
	if err != nil {
		return job, err
	}
	_, err = r.db.Exec(ctx, `UPDATE dataset_versions SET status = 'analyzing' WHERE id = $1`, job.DatasetVersionID)
	return job, err
}

func (r *Repository) FinishAnalysis(ctx context.Context, jobID uuid.UUID, readiness float64, metrics []models.ObjectMetric, recommendations []models.Recommendation, roadmap []models.RoadmapItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var datasetVersionID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT dataset_version_id FROM analysis_jobs WHERE id = $1`, jobID).Scan(&datasetVersionID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM object_metrics WHERE object_id IN (SELECT id FROM data_objects WHERE dataset_version_id = $1)`, datasetVersionID)
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		reasons, err := json.Marshal(metric.Reasons)
		if err != nil {
			return err
		}
		probabilities, err := json.Marshal(metric.Probabilities)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO object_metrics (
				id, object_id, entropy, uncertainty_score, label_error_probability,
				duplicate_score, rarity_score, class_deficit_score, quality_score,
				novelty_score, object_utility_score, final_score, reasons,
				recommendation, probabilities
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`, metric.ID, metric.ObjectID, metric.Entropy, metric.UncertaintyScore, metric.LabelErrorProbability,
			metric.DuplicateScore, metric.RarityScore, metric.ClassDeficitScore, metric.QualityScore,
			metric.NoveltyScore, metric.ObjectUtilityScore, metric.FinalScore, reasons, metric.Recommendation, probabilities)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `DELETE FROM recommendations WHERE dataset_version_id = $1`, datasetVersionID)
	if err != nil {
		return err
	}
	for _, recommendation := range recommendations {
		objectIDs, err := json.Marshal(recommendation.ObjectIDs)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO recommendations (
				id, project_id, dataset_version_id, type, priority, title,
				description, affected_objects_count, object_ids
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, recommendation.ID, recommendation.ProjectID, datasetVersionID, recommendation.Type, recommendation.Priority,
			recommendation.Title, recommendation.Description, recommendation.AffectedObjectsCount, objectIDs)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `DELETE FROM roadmap_items WHERE dataset_version_id = $1`, datasetVersionID)
	if err != nil {
		return err
	}
	for _, item := range roadmap {
		_, err = tx.Exec(ctx, `
			INSERT INTO roadmap_items (
				id, project_id, dataset_version_id, priority, title,
				description, action_type, expected_impact
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, item.ID, item.ProjectID, datasetVersionID, item.Priority, item.Title,
			item.Description, item.ActionType, item.ExpectedImpact)
		if err != nil {
			return err
		}
	}

	if _, err = tx.Exec(ctx, `UPDATE dataset_versions SET status = 'analyzed', readiness_score = $1 WHERE id = $2`, readiness, datasetVersionID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE analysis_jobs SET status = 'done', finished_at = NOW() WHERE id = $1`, jobID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) FailAnalysisJob(ctx context.Context, jobID uuid.UUID, message string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `
		UPDATE dataset_versions
		SET status = 'analysis_failed'
		WHERE id = (SELECT dataset_version_id FROM analysis_jobs WHERE id = $1)
	`, jobID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `
		UPDATE analysis_jobs SET status = 'failed', error_message = $2, finished_at = NOW() WHERE id = $1
	`, jobID, message); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListMetrics(ctx context.Context, datasetVersionID uuid.UUID) ([]models.ObjectMetric, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.id, m.object_id, COALESCE(m.entropy, 0), COALESCE(m.uncertainty_score, 0),
			COALESCE(m.label_error_probability, 0), COALESCE(m.duplicate_score, 0),
			COALESCE(m.rarity_score, 0), COALESCE(m.class_deficit_score, 0),
			COALESCE(m.quality_score, 0), COALESCE(m.novelty_score, 0),
			COALESCE(m.object_utility_score, 0), COALESCE(m.final_score, 0),
			COALESCE(m.reasons, '[]'::jsonb), COALESCE(m.recommendation, ''),
			COALESCE(m.probabilities, '{}'::jsonb), m.created_at
		FROM object_metrics m
		JOIN data_objects o ON o.id = m.object_id
		WHERE o.dataset_version_id = $1
		ORDER BY m.final_score DESC
	`, datasetVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.ObjectMetric
	for rows.Next() {
		metric, err := scanObjectMetric(rows)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (r *Repository) GetMetricByObject(ctx context.Context, objectID uuid.UUID) (models.ObjectMetric, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, object_id, COALESCE(entropy, 0), COALESCE(uncertainty_score, 0),
			COALESCE(label_error_probability, 0), COALESCE(duplicate_score, 0),
			COALESCE(rarity_score, 0), COALESCE(class_deficit_score, 0),
			COALESCE(quality_score, 0), COALESCE(novelty_score, 0),
			COALESCE(object_utility_score, 0), COALESCE(final_score, 0),
			COALESCE(reasons, '[]'::jsonb), COALESCE(recommendation, ''),
			COALESCE(probabilities, '{}'::jsonb), created_at
		FROM object_metrics WHERE object_id = $1
	`, objectID)
	metric, err := scanObjectMetric(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return metric, ErrNotFound
	}
	return metric, err
}

func (r *Repository) ListRecommendations(ctx context.Context, projectID uuid.UUID) ([]models.Recommendation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, dataset_version_id, COALESCE(type, ''), COALESCE(priority, ''),
			COALESCE(title, ''), COALESCE(description, ''), COALESCE(affected_objects_count, 0),
			COALESCE(object_ids, '[]'::jsonb), created_at
		FROM recommendations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Recommendation
	for rows.Next() {
		item, err := scanRecommendation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListRecommendationsForVersion(ctx context.Context, projectID, datasetVersionID uuid.UUID) ([]models.Recommendation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, dataset_version_id, COALESCE(type, ''), COALESCE(priority, ''),
			COALESCE(title, ''), COALESCE(description, ''), COALESCE(affected_objects_count, 0),
			COALESCE(object_ids, '[]'::jsonb), created_at
		FROM recommendations
		WHERE project_id = $1 AND dataset_version_id = $2
		ORDER BY
			CASE COALESCE(priority, '')
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				ELSE 3
			END,
			created_at DESC
	`, projectID, datasetVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Recommendation
	for rows.Next() {
		item, err := scanRecommendation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListRoadmap(ctx context.Context, projectID uuid.UUID) ([]models.RoadmapItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, dataset_version_id, COALESCE(priority, 0), COALESCE(title, ''),
			COALESCE(description, ''), COALESCE(action_type, ''), COALESCE(expected_impact, ''), created_at
		FROM roadmap_items
		WHERE project_id = $1
		ORDER BY priority ASC, created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.RoadmapItem
	for rows.Next() {
		item, err := scanRoadmapItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListRoadmapForVersion(ctx context.Context, projectID, datasetVersionID uuid.UUID) ([]models.RoadmapItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, dataset_version_id, COALESCE(priority, 0), COALESCE(title, ''),
			COALESCE(description, ''), COALESCE(action_type, ''), COALESCE(expected_impact, ''), created_at
		FROM roadmap_items
		WHERE project_id = $1 AND dataset_version_id = $2
		ORDER BY priority ASC, created_at DESC
	`, projectID, datasetVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.RoadmapItem
	for rows.Next() {
		item, err := scanRoadmapItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateExport(ctx context.Context, export models.Export) (models.Export, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO exports (id, project_id, dataset_version_id, file_path, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, export.ID, export.ProjectID, export.DatasetVersionID, export.FilePath, export.Status).Scan(&export.CreatedAt)
	return export, err
}

func (r *Repository) GetExport(ctx context.Context, projectID, exportID uuid.UUID) (models.Export, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, project_id, dataset_version_id, file_path, status, created_at
		FROM exports WHERE project_id = $1 AND id = $2
	`, projectID, exportID)
	var export models.Export
	err := row.Scan(&export.ID, &export.ProjectID, &export.DatasetVersionID, &export.FilePath, &export.Status, &export.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return export, ErrNotFound
	}
	return export, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProject(row scanner) (models.Project, error) {
	var p models.Project
	var classes []byte
	if err := row.Scan(&p.ID, &p.Name, &p.Modality, &p.TaskType, &classes, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return p, err
	}
	if len(classes) > 0 {
		if err := json.Unmarshal(classes, &p.Classes); err != nil {
			return p, err
		}
	}
	return p, nil
}

func scanDatasetVersion(row scanner) (models.DatasetVersion, error) {
	var version models.DatasetVersion
	err := row.Scan(&version.ID, &version.ProjectID, &version.VersionName, &version.Status, &version.ObjectsCount, &version.ReadinessScore, &version.CreatedAt)
	return version, err
}

func scanDataObject(row scanner) (models.DataObject, error) {
	var object models.DataObject
	var metadata []byte
	err := row.Scan(&object.ID, &object.DatasetVersionID, &object.ExternalID, &object.FilePath,
		&object.TextContent, &object.Label, &object.PredictedLabel, &object.Confidence,
		&object.Split, &object.Source, &object.Annotator, &metadata, &object.Status, &object.CreatedAt)
	if err != nil {
		return object, err
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &object.Metadata); err != nil {
			return object, err
		}
	}
	return object, nil
}

func scanObjectMetric(row scanner) (models.ObjectMetric, error) {
	var metric models.ObjectMetric
	var reasons, probabilities []byte
	err := row.Scan(&metric.ID, &metric.ObjectID, &metric.Entropy, &metric.UncertaintyScore,
		&metric.LabelErrorProbability, &metric.DuplicateScore, &metric.RarityScore,
		&metric.ClassDeficitScore, &metric.QualityScore, &metric.NoveltyScore,
		&metric.ObjectUtilityScore, &metric.FinalScore, &reasons, &metric.Recommendation,
		&probabilities, &metric.CreatedAt)
	if err != nil {
		return metric, err
	}
	if len(reasons) > 0 {
		if err := json.Unmarshal(reasons, &metric.Reasons); err != nil {
			return metric, err
		}
	}
	if len(probabilities) > 0 {
		if err := json.Unmarshal(probabilities, &metric.Probabilities); err != nil {
			return metric, err
		}
	}
	return metric, nil
}

func scanRecommendation(row scanner) (models.Recommendation, error) {
	var item models.Recommendation
	var objectIDs []byte
	err := row.Scan(&item.ID, &item.ProjectID, &item.DatasetVersionID, &item.Type, &item.Priority,
		&item.Title, &item.Description, &item.AffectedObjectsCount, &objectIDs, &item.CreatedAt)
	if err != nil {
		return item, err
	}
	if len(objectIDs) > 0 {
		if err := json.Unmarshal(objectIDs, &item.ObjectIDs); err != nil {
			return item, err
		}
	}
	return item, nil
}

func scanRoadmapItem(row scanner) (models.RoadmapItem, error) {
	var item models.RoadmapItem
	err := row.Scan(&item.ID, &item.ProjectID, &item.DatasetVersionID, &item.Priority,
		&item.Title, &item.Description, &item.ActionType, &item.ExpectedImpact, &item.CreatedAt)
	return item, err
}

func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

func WrapNotFound(entity string, id uuid.UUID) error {
	return fmt.Errorf("%s %s: %w", entity, id, ErrNotFound)
}
