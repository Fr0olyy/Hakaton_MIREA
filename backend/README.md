# Backend Go Level 1

Go backend for the DataForge AI Level 1 MVP. It supports image classification dataset projects, CSV and ZIP upload, local deterministic analysis, dashboard APIs, review queue, recommendations, roadmap, object image serving, and curated ZIP export.

The backend is self-contained for demos: if no ML service exists, `POST /api/projects/{id}/analyze` computes the required metrics locally and marks the analysis source as `backend_local`.

## Quick Start

From the repository root:

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

For local development:

```bash
cd backend
go mod tidy
go run ./cmd/api
```

The backend expects PostgreSQL at `DATABASE_URL`. Migrations run automatically on startup.

## Level 1 Scope

- Only `modality=image` and `task_type=classification` are accepted.
- No frontend, auth, MinIO, or required ML service is included in this backend pass.
- Storage is local under `STORAGE_DIR/projects/{project_id}`.

## CSV Format

Minimal:

```csv
id,file_path,label
1,images/cat_001.jpg,cat
2,images/dog_001.jpg,dog
```

Extended:

```csv
id,file_path,label,predicted_label,confidence,split,source,annotator,prob_cat,prob_dog
1,images/cat_001.jpg,cat,cat,0.94,train,source_a,ann_1,0.82,0.18
2,images/dog_001.jpg,dog,cat,0.87,train,source_b,ann_2,0.41,0.59
```

Supported known columns: `id`, `external_id`, `file_path`, `text_content`, `label`, `predicted_label`, `confidence`, `split`, `source`, `annotator`, and any `prob_*` probability columns. Unknown columns are saved into object `metadata`.

Upload validation stores every row and assigns one of: `ok`, `missing_file`, `missing_label`, `unknown_class`, `invalid`.

## Metrics

Local analysis computes:

- prediction entropy
- uncertainty score
- label error probability
- empirical class distribution
- normalized imbalance index
- class deficit score
- duplicate score
- novelty score
- object utility and final review score
- dataset readiness score

Dashboard and probabilistic analysis responses include summary fields such as `class_distribution`, `imbalance_index`, `avg_entropy`, `avg_label_error_probability`, `missing_files_count`, `review_items`, and `analysis_source`.

## Curl Happy Path

```bash
PROJECT_ID=$(curl -s http://localhost:8080/api/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"Demo","modality":"image","task_type":"classification","classes":["cat","dog"]}' | jq -r .id)

curl -X POST "http://localhost:8080/api/projects/$PROJECT_ID/upload" \
  -F dataset=@dataset.csv \
  -F images=@images.zip

curl -X POST "http://localhost:8080/api/projects/$PROJECT_ID/analyze"
curl "http://localhost:8080/api/projects/$PROJECT_ID/dashboard"
curl "http://localhost:8080/api/projects/$PROJECT_ID/review-queue"

EXPORT_ID=$(curl -s -X POST "http://localhost:8080/api/projects/$PROJECT_ID/export" | jq -r .id)
curl -L "http://localhost:8080/api/projects/$PROJECT_ID/exports/$EXPORT_ID/download" -o export.zip
```

## Endpoints

- `GET /health`
- `POST /api/projects`
- `GET /api/projects`
- `GET /api/projects/{id}`
- `POST /api/projects/{id}/upload`
- `POST /api/projects/{id}/analyze`
- `GET /api/projects/{id}/dashboard`
- `GET /api/projects/{id}/probabilistic-analysis`
- `GET /api/projects/{id}/review-queue`
- `GET /api/projects/{id}/objects/{objectId}`
- `GET /api/projects/{id}/objects/{objectId}/file`
- `GET /api/projects/{id}/recommendations`
- `GET /api/projects/{id}/roadmap`
- `POST /api/projects/{id}/export`
- `GET /api/projects/{id}/exports/{exportId}/download`
- `POST /api/projects/{id}/agent/summary`

## Export Contents

`POST /api/projects/{id}/export` creates a ZIP containing:

- `dataset_v2.csv`
- `dataset_report.json`
- `review_queue.csv`
- `recommendations.json`
- `roadmap.json`
- valid image files under `images/`

Rows with `missing_file` or `invalid` status are excluded from `dataset_v2.csv`; other rows are preserved with review flags and metric columns.

## Verification

```bash
go test ./...
go vet ./...
```
