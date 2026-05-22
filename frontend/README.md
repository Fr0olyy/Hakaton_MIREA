# DataForge AI Frontend

Next.js + TypeScript + Tailwind CSS frontend for the Level 1 image classification dataset workflow.

## Pages

- `/projects`
- `/projects/new`
- `/projects/{projectId}/upload`
- `/projects/{projectId}/dashboard`
- `/projects/{projectId}/probabilistic`
- `/projects/{projectId}/queue`
- `/projects/{projectId}/recommendations`
- `/projects/{projectId}/roadmap`
- `/projects/{projectId}/export`

The frontend proxies backend calls through `/api/backend/*`. In Docker, set `BACKEND_INTERNAL_URL=http://backend:8080`.
