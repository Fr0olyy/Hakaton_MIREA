package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hakaton/backend/internal/auth"
	"hakaton/backend/internal/models"
	"hakaton/backend/internal/repositories"
	"hakaton/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func testAuthToken(t *testing.T) string {
	t.Helper()
	token, err := auth.GenerateToken(models.User{
		ID:    uuid.New(),
		Email: "test@test.com",
		Name:  "Test",
		Role:  models.RoleAdmin,
	}, "test-secret")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func authRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testAuthToken(t))
	return req
}

func testHandler(t *testing.T) *Handler {
	t.Helper()
	return New(services.New(nil, nil, nil, nil, "test-secret"), "test-secret")
}

func testRouter(t *testing.T) chi.Router {
	t.Helper()
	router := chi.NewRouter()
	testHandler(t).Register(router)
	return router
}

func TestProtectedEndpointsRequireAuth(t *testing.T) {
	router := testRouter(t)
	protectedRoutes := []struct {
		method, path string
	}{
		{"GET", "/api/projects"},
		{"POST", "/api/projects"},
		{"GET", "/api/projects/" + uuid.NewString() + "/dashboard"},
		{"GET", "/api/jobs/" + uuid.NewString()},
		{"GET", "/api/teams"},
		{"POST", "/api/teams"},
	}
	for _, route := range protectedRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(route.method, route.path, nil)
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s: status = %d, want %d", route.method, route.path, recorder.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthEndpointsArePublic(t *testing.T) {
	router := testRouter(t)

	// Verify register/login routes exist and are not blocked by auth middleware
	// (they may fail with 500 due to nil repo, but NOT 401)
	routes := []struct {
		method, path string
	}{
		{"POST", "/api/auth/register"},
		{"POST", "/api/auth/login"},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(route.method, route.path, strings.NewReader(`{}`))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)
			if recorder.Code == http.StatusUnauthorized {
				t.Fatalf("%s %s: should be public, got 401", route.method, route.path)
			}
		})
	}
}

func TestWriteResultMapsNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeResult(recorder, nil, repositories.ErrNotFound)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestAuthMiddlewareRejectsBadToken(t *testing.T) {
	router := testRouter(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
