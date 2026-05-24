package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// ---- Users ----

func (r *Repository) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, user.ID, user.Email, user.PasswordHash, user.Name, user.Role).Scan(&user.CreatedAt)
	return user, err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, created_at
		FROM users WHERE email = $1
	`, email)
	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, created_at
		FROM users WHERE id = $1
	`, id)
	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, err
}

// ---- Teams ----

func (r *Repository) CreateTeam(ctx context.Context, team models.Team) (models.Team, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO teams (id, name, created_by)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`, team.ID, team.Name, team.CreatedBy).Scan(&team.CreatedAt)
	return team, err
}

func (r *Repository) ListTeamsByUser(ctx context.Context, userID uuid.UUID) ([]models.Team, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.name, t.created_by, t.created_at
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = $1
		ORDER BY t.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []models.Team
	for rows.Next() {
		var team models.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (r *Repository) GetTeam(ctx context.Context, teamID uuid.UUID) (models.Team, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, created_by, created_at FROM teams WHERE id = $1
	`, teamID)
	var team models.Team
	err := row.Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return team, ErrNotFound
	}
	return team, err
}

func (r *Repository) AddTeamMember(ctx context.Context, member models.TeamMember) (models.TeamMember, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO team_members (id, team_id, user_id, role)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, member.ID, member.TeamID, member.UserID, member.Role).Scan(&member.CreatedAt)
	return member, err
}

func (r *Repository) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM team_members WHERE team_id = $1 AND user_id = $2
	`, teamID, userID)
	return err
}

func (r *Repository) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, team_id, user_id, role, created_at
		FROM team_members WHERE team_id = $1
		ORDER BY created_at ASC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.ID, &member.TeamID, &member.UserID, &member.Role, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

// ---- Project Members ----

func (r *Repository) AddProjectMember(ctx context.Context, member models.ProjectMember) (models.ProjectMember, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO project_members (id, project_id, user_id, role)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, member.ID, member.ProjectID, member.UserID, member.Role).Scan(&member.CreatedAt)
	return member, err
}

func (r *Repository) RemoveProjectMember(ctx context.Context, projectID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID)
	return err
}

func (r *Repository) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, user_id, role, created_at
		FROM project_members WHERE project_id = $1
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []models.ProjectMember
	for rows.Next() {
		var member models.ProjectMember
		if err := rows.Scan(&member.ID, &member.ProjectID, &member.UserID, &member.Role, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (r *Repository) GetProjectMemberRole(ctx context.Context, projectID, userID uuid.UUID) (models.ProjectMemberRole, error) {
	row := r.db.QueryRow(ctx, `
		SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID)
	var role models.ProjectMemberRole
	err := row.Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

// ---- Projects ----

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

type ListObjectsFilter struct {
	DatasetVersionID uuid.UUID
	Status           string
	Label            string
	Recommendation   string
	ScoreMin         float64
	ScoreMax         float64
	Search           string
	SortBy           string
	SortOrder        string
	Offset           int
	Limit            int
}

func (r *Repository) ListObjectsPaginated(ctx context.Context, filter ListObjectsFilter) ([]models.DataObject, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "o.dataset_version_id = $"+itoa(argIdx))
	args = append(args, filter.DatasetVersionID)
	argIdx++

	if filter.Status != "" {
		conditions = append(conditions, "o.status = $"+itoa(argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Label != "" {
		conditions = append(conditions, "o.label = $"+itoa(argIdx))
		args = append(args, filter.Label)
		argIdx++
	}
	if filter.Recommendation != "" {
		conditions = append(conditions, "m.recommendation = $"+itoa(argIdx))
		args = append(args, filter.Recommendation)
		argIdx++
	}
	if filter.ScoreMin > 0 {
		conditions = append(conditions, "m.final_score >= $"+itoa(argIdx))
		args = append(args, filter.ScoreMin)
		argIdx++
	}
	if filter.ScoreMax > 0 {
		conditions = append(conditions, "m.final_score <= $"+itoa(argIdx))
		args = append(args, filter.ScoreMax)
		argIdx++
	}
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		conditions = append(conditions, "(o.file_path ILIKE $"+itoa(argIdx)+" OR o.label ILIKE $"+itoa(argIdx+1)+" OR o.external_id ILIKE $"+itoa(argIdx+2)+")")
		args = append(args, search, search, search)
		argIdx += 3
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count
	var total int
	countQuery := `SELECT COUNT(*) FROM data_objects o LEFT JOIN object_metrics m ON m.object_id = o.id WHERE ` + whereClause
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Sort
	orderBy := "o.created_at ASC"
	switch filter.SortBy {
	case "entropy":
		orderBy = "COALESCE(m.entropy, 0) " + filter.SortOrder
	case "label_error_probability":
		orderBy = "COALESCE(m.label_error_probability, 0) " + filter.SortOrder
	case "quality_score":
		orderBy = "COALESCE(m.quality_score, 0) " + filter.SortOrder
	case "duplicate_score":
		orderBy = "COALESCE(m.duplicate_score, 0) " + filter.SortOrder
	case "object_utility_score":
		orderBy = "COALESCE(m.object_utility_score, 0) " + filter.SortOrder
	case "final_score":
		orderBy = "COALESCE(m.final_score, 0) " + filter.SortOrder
	}

	query := `
		SELECT o.id, o.dataset_version_id, COALESCE(o.external_id, ''), COALESCE(o.file_path, ''),
			COALESCE(o.text_content, ''), COALESCE(o.label, ''), COALESCE(o.predicted_label, ''),
			COALESCE(o.confidence, 0), COALESCE(o.split, ''), COALESCE(o.source, ''),
			COALESCE(o.annotator, ''), COALESCE(o.metadata, '{}'::jsonb), COALESCE(o.status, 'ok'), o.created_at
		FROM data_objects o
		LEFT JOIN object_metrics m ON m.object_id = o.id
		WHERE ` + whereClause + `
		ORDER BY ` + orderBy + `
		LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var objects []models.DataObject
	for rows.Next() {
		object, err := scanDataObject(rows)
		if err != nil {
			return nil, 0, err
		}
		objects = append(objects, object)
	}
	return objects, total, rows.Err()
}

// ---- Object Actions ----

func (r *Repository) CreateObjectAction(ctx context.Context, action models.ObjectAction) (models.ObjectAction, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO object_actions (id, project_id, object_id, user_id, action, old_value, new_value, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`, action.ID, action.ProjectID, action.ObjectID, action.UserID, action.Action, action.OldValue, action.NewValue, action.Comment).Scan(&action.CreatedAt)
	return action, err
}

func (r *Repository) ListProjectActions(ctx context.Context, projectID uuid.UUID) ([]models.ObjectAction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, object_id, user_id, COALESCE(action, ''), COALESCE(old_value, ''), COALESCE(new_value, ''), COALESCE(comment, ''), created_at
		FROM object_actions
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var actions []models.ObjectAction
	for rows.Next() {
		var a models.ObjectAction
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.ObjectID, &a.UserID, &a.Action, &a.OldValue, &a.NewValue, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, rows.Err()
}

func (r *Repository) ListObjectActions(ctx context.Context, projectID, objectID uuid.UUID) ([]models.ObjectAction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, object_id, user_id, COALESCE(action, ''), COALESCE(old_value, ''), COALESCE(new_value, ''), COALESCE(comment, ''), created_at
		FROM object_actions
		WHERE project_id = $1 AND object_id = $2
		ORDER BY created_at DESC
	`, projectID, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var actions []models.ObjectAction
	for rows.Next() {
		var a models.ObjectAction
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.ObjectID, &a.UserID, &a.Action, &a.OldValue, &a.NewValue, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, rows.Err()
}

// ---- Object Comments ----

func (r *Repository) CreateObjectComment(ctx context.Context, comment models.ObjectComment) (models.ObjectComment, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO object_comments (id, project_id, object_id, user_id, text)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, comment.ID, comment.ProjectID, comment.ObjectID, comment.UserID, comment.Text).Scan(&comment.CreatedAt)
	return comment, err
}

func (r *Repository) ListObjectComments(ctx context.Context, projectID, objectID uuid.UUID) ([]models.ObjectComment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, object_id, user_id, text, created_at
		FROM object_comments
		WHERE project_id = $1 AND object_id = $2
		ORDER BY created_at ASC
	`, projectID, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.ObjectComment
	for rows.Next() {
		var c models.ObjectComment
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.ObjectID, &c.UserID, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// ---- Review Assignments ----

func (r *Repository) CreateReviewAssignment(ctx context.Context, assignment models.ReviewAssignment) (models.ReviewAssignment, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO review_assignments (id, project_id, object_id, user_id, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, assignment.ID, assignment.ProjectID, assignment.ObjectID, assignment.UserID, assignment.Status).Scan(&assignment.CreatedAt)
	return assignment, err
}

func (r *Repository) ListReviewAssignmentsByUser(ctx context.Context, projectID, userID uuid.UUID) ([]models.ReviewAssignment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, object_id, user_id, COALESCE(status, 'assigned'), created_at
		FROM review_assignments
		WHERE project_id = $1 AND user_id = $2
		ORDER BY created_at DESC
	`, projectID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assignments []models.ReviewAssignment
	for rows.Next() {
		var a models.ReviewAssignment
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.ObjectID, &a.UserID, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
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

func (r *Repository) GetJob(ctx context.Context, jobID uuid.UUID) (models.AnalysisJob, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, project_id, dataset_version_id, COALESCE(status, ''), COALESCE(error_message, ''),
			COALESCE(progress_percent, 0), COALESCE(progress_stage, ''), COALESCE(output_files, '{}'::jsonb),
			COALESCE(started_at, NOW()), finished_at
		FROM analysis_jobs WHERE id = $1
	`, jobID)
	return scanAnalysisJob(row)
}

func (r *Repository) ListJobsByProject(ctx context.Context, projectID uuid.UUID) ([]models.AnalysisJob, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, dataset_version_id, COALESCE(status, ''), COALESCE(error_message, ''),
			COALESCE(progress_percent, 0), COALESCE(progress_stage, ''), COALESCE(output_files, '{}'::jsonb),
			COALESCE(started_at, NOW()), finished_at
		FROM analysis_jobs WHERE project_id = $1
		ORDER BY started_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.AnalysisJob
	for rows.Next() {
		job, err := scanAnalysisJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *Repository) UpdateJobProgress(ctx context.Context, jobID uuid.UUID, percent int, stage string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE analysis_jobs SET progress_percent = $1, progress_stage = $2 WHERE id = $3
	`, percent, stage, jobID)
	return err
}

func (r *Repository) UpdateJobOutput(ctx context.Context, jobID uuid.UUID, outputFiles map[string]string) error {
	data, err := json.Marshal(outputFiles)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		UPDATE analysis_jobs SET output_files = $1 WHERE id = $2
	`, data, jobID)
	return err
}

func (r *Repository) FinishAnalysis(ctx context.Context, jobID uuid.UUID, readiness float64, metrics []models.ObjectMetric, recommendations []models.Recommendation, roadmap []models.RoadmapItem, objectStatuses map[uuid.UUID]string, outputFiles map[string]string) error {
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

	for objectID, status := range objectStatuses {
		if status == "" {
			continue
		}
		_, err = tx.Exec(ctx, `
			UPDATE data_objects
			SET status = $1
			WHERE dataset_version_id = $2 AND id = $3
		`, status, datasetVersionID, objectID)
		if err != nil {
			return err
		}
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
	outputData, err := json.Marshal(outputFiles)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE analysis_jobs SET status = 'completed', progress_percent = 100, progress_stage = 'done', output_files = $1, finished_at = NOW() WHERE id = $2`, outputData, jobID); err != nil {
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

func scanAnalysisJob(row scanner) (models.AnalysisJob, error) {
	var job models.AnalysisJob
	var outputFiles []byte
	var finishedAt *time.Time
	err := row.Scan(&job.ID, &job.ProjectID, &job.DatasetVersionID, &job.Status, &job.ErrorMessage,
		&job.ProgressPercent, &job.ProgressStage, &outputFiles, &job.StartedAt, &finishedAt)
	if err != nil {
		return job, err
	}
	if finishedAt != nil {
		job.FinishedAt = *finishedAt
	}
	if len(outputFiles) > 0 {
		if err := json.Unmarshal(outputFiles, &job.OutputFiles); err != nil {
			return job, err
		}
	}
	return job, nil
}

// ---- Collection Tasks ----

func (r *Repository) CreateCollectionTask(ctx context.Context, task models.CollectionTask) (models.CollectionTask, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO collection_tasks (id, project_id, target_class, target_count, priority, risk, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`, task.ID, task.ProjectID, task.TargetClass, task.TargetCount, task.Priority, task.Risk, task.Status, task.CreatedBy).Scan(&task.CreatedAt)
	return task, err
}

func (r *Repository) ListCollectionTasks(ctx context.Context, projectID uuid.UUID) ([]models.CollectionTask, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, target_class, target_count, COALESCE(priority, 'medium'), COALESCE(risk, 'low'), COALESCE(status, 'draft'), created_by, created_at
		FROM collection_tasks WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []models.CollectionTask
	for rows.Next() {
		var t models.CollectionTask
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.TargetClass, &t.TargetCount, &t.Priority, &t.Risk, &t.Status, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *Repository) UpdateCollectionTask(ctx context.Context, taskID uuid.UUID, updates map[string]any) error {
	setClauses := make([]string, 0, len(updates))
	args := []any{taskID}
	argIdx := 2
	for key, value := range updates {
		setClauses = append(setClauses, key+" = $"+itoa(argIdx))
		args = append(args, value)
		argIdx++
	}
	query := "UPDATE collection_tasks SET " + strings.Join(setClauses, ", ") + " WHERE id = $1"
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// ---- Synthetic Tasks ----

func (r *Repository) CreateSyntheticTask(ctx context.Context, task models.SyntheticTask) (models.SyntheticTask, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO synthetic_tasks (id, project_id, target_class, target_count, prompt, negative_prompt, priority, risk, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at
	`, task.ID, task.ProjectID, task.TargetClass, task.TargetCount, task.Prompt, task.NegativePrompt, task.Priority, task.Risk, task.Status, task.CreatedBy).Scan(&task.CreatedAt)
	return task, err
}

func (r *Repository) ListSyntheticTasks(ctx context.Context, projectID uuid.UUID) ([]models.SyntheticTask, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, target_class, target_count, COALESCE(prompt, ''), COALESCE(negative_prompt, ''), COALESCE(priority, 'medium'), COALESCE(risk, 'low'), COALESCE(status, 'draft'), created_by, created_at
		FROM synthetic_tasks WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []models.SyntheticTask
	for rows.Next() {
		var t models.SyntheticTask
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.TargetClass, &t.TargetCount, &t.Prompt, &t.NegativePrompt, &t.Priority, &t.Risk, &t.Status, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *Repository) UpdateSyntheticTask(ctx context.Context, taskID uuid.UUID, updates map[string]any) error {
	setClauses := make([]string, 0, len(updates))
	args := []any{taskID}
	argIdx := 2
	for key, value := range updates {
		setClauses = append(setClauses, key+" = $"+itoa(argIdx))
		args = append(args, value)
		argIdx++
	}
	query := "UPDATE synthetic_tasks SET " + strings.Join(setClauses, ", ") + " WHERE id = $1"
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// ---- Metrics ----

func itoa(i int) string {
	return strconv.Itoa(i)
}

func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

func WrapNotFound(entity string, id uuid.UUID) error {
	return fmt.Errorf("%s %s: %w", entity, id, ErrNotFound)
}
