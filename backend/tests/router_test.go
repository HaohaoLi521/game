package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"this-is-pun/backend/internal/data"
	"this-is-pun/backend/internal/router"
	"this-is-pun/backend/internal/service"
)

func newTestRouter() http.Handler {
	game := service.NewGameService(data.NewMemoryRepository(data.Seed()))
	return router.New(game)
}

func TestExplanationRouteReturnsAnswerAndExplanation(t *testing.T) {
	engine := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/puzzles/101/explanation", nil)
	res := httptest.NewRecorder()

	engine.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Data struct {
			Answer      string `json:"answer"`
			Explanation string `json:"explanation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Answer == "" || body.Data.Explanation == "" {
		t.Fatalf("expected answer and explanation, got %+v", body.Data)
	}
}

func TestAdminLogoutInvalidatesSession(t *testing.T) {
	engine := newTestRouter()
	token := registerTestAdmin(t, engine)

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	logoutRes := httptest.NewRecorder()
	engine.ServeHTTP(logoutRes, logoutReq)

	if logoutRes.Code != http.StatusOK {
		t.Fatalf("expected logout 200, got %d: %s", logoutRes.Code, logoutRes.Body.String())
	}

	protectedReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/puzzles", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+token)
	protectedRes := httptest.NewRecorder()
	engine.ServeHTTP(protectedRes, protectedReq)

	if protectedRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected logged out token to be unauthorized, got %d", protectedRes.Code)
	}
}

func registerTestAdmin(t *testing.T, engine http.Handler) string {
	t.Helper()
	payload := []byte(`{"username":"admin","password":"secret123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	engine.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected register 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Token == "" {
		t.Fatal("expected auth token")
	}
	return body.Data.Token
}
