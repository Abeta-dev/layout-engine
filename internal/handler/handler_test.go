package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/engine"
	"github.com/abeta-dev/layout-engine/internal/handler"
	"github.com/abeta-dev/layout-engine/internal/infra/persistence"
	"github.com/abeta-dev/layout-engine/internal/usecase"
	"github.com/abeta-dev/layout-engine/pkg/transformer"
	"github.com/gin-gonic/gin"
)

type apiEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail"`
	} `json:"error,omitempty"`
}

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	store, err := persistence.Open(ctx, persistence.Config{
		Dialect:     persistence.DialectSQLite,
		DSN:         ":memory:",
		AutoMigrate: true,
	})
	if err != nil {
		t.Fatalf("open sqlite store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if err := persistence.SeedDefaultData(ctx, store); err != nil {
		t.Fatalf("seed default data: %v", err)
	}

	eval := engine.NewRuleEvaluator()
	slotEng := engine.NewSlottingEngine(eval)
	val := engine.NewLayoutValidator()
	comp := transformer.NewPageComposer(transformer.NewDefaultRegistry())

	return handler.NewRouter(handler.RouterConfig{
		ServiceName:       "abeta-layout-engine-test",
		Store:             store,
		ResolveService:    usecase.NewResolveService(store, eval, slotEng, comp),
		StudioService:     usecase.NewStudioService(store, val, comp, engine.DiffJSON),
		ActionService:     usecase.NewActionService(),
		EnableCompression: true,
		EnableETag:        true,
	})
}

func doJSON(t *testing.T, r http.Handler, method, path string, body any, headers map[string]string) (*httptest.ResponseRecorder, apiEnvelope) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var env apiEnvelope
	if rec.Body.Len() > 0 && strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("decode envelope for %s %s (status %d): %v | raw=%s", method, path, rec.Code, err, rec.Body.String())
		}
	}
	return rec, env
}

func TestRuntimeDynamicResolutionSlottingAndETag(t *testing.T) {
	t.Parallel()
	r := setupTestRouter(t)

	// 1. List published screens (KMP drop-in endpoint)
	rec, env := doJSON(t, r, http.MethodGet, "/api/v1/screens", nil, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("GET /api/v1/screens failed: code=%d body=%s", rec.Code, rec.Body.String())
	}
	var screensPayload struct {
		Screens []string `json:"screens"`
	}
	if err := json.Unmarshal(env.Data, &screensPayload); err != nil {
		t.Fatalf("decode screens: %v", err)
	}
	if len(screensPayload.Screens) < 5 {
		t.Fatalf("expected >= 5 seeded screens, got %v", screensPayload.Screens)
	}

	// 2. Resolve store-home for VIP segment in Mumbai with trace enabled
	rec, env = doJSON(t, r, http.MethodGet, "/api/v1/pages/store-home?include_trace=true&segment=vip&city=Mumbai&at=2026-11-21T14:00:00Z", nil, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("GET store-home failed: code=%d body=%s", rec.Code, rec.Body.String())
	}
	var resolved domain.ResolvedPageResult
	if err := json.Unmarshal(env.Data, &resolved); err != nil {
		t.Fatalf("decode resolved page: %v", err)
	}
	if resolved.Trace.SnapshotVersion != 1 {
		t.Fatalf("expected snapshot v1, got %d", resolved.Trace.SnapshotVersion)
	}
	if len(resolved.Trace.AppliedSlottingIDs) == 0 {
		t.Fatalf("expected VIP slotting rule to apply, got trace=%+v", resolved.Trace)
	}
	rawPage, _ := json.Marshal(resolved.Page)
	if !strings.Contains(string(rawPage), "Mumbai") {
		t.Fatalf("expected {{city}} placeholder to interpolate Mumbai in composed page")
	}
	if !strings.Contains(string(rawPage), "VIP") {
		t.Fatalf("expected VIP flash banner widget injected via slotting rule")
	}

	// 3. ETag 304 Not Modified on /api/v1/screens/store-home
	rec1, _ := doJSON(t, r, http.MethodGet, "/api/v1/screens/store-home", nil, nil)
	etag := rec1.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header on GET /api/v1/screens/store-home")
	}
	rec2, _ := doJSON(t, r, http.MethodGet, "/api/v1/screens/store-home", nil, map[string]string{"If-None-Match": etag})
	if rec2.Code != http.StatusNotModified {
		t.Fatalf("expected 304 Not Modified, got %d", rec2.Code)
	}

	// 4. Brotli compression negotiation
	recBr := httptest.NewRecorder()
	reqBr := httptest.NewRequest(http.MethodGet, "/api/v1/screens/store-home", nil)
	reqBr.Header.Set("Accept-Encoding", "br, gzip")
	r.ServeHTTP(recBr, reqBr)
	if recBr.Header().Get("Content-Encoding") != "br" {
		t.Fatalf("expected br encoding, got %q", recBr.Header().Get("Content-Encoding"))
	}
}

func TestStudioDynamicPagePublishRollbackAndActionFlow(t *testing.T) {
	t.Parallel()
	r := setupTestRouter(t)

	// 1. Create a brand-new dynamic page via Studio API (zero client deploy required!)
	newLayout := domain.Layout{
		Slug:        "festival-drop",
		Title:       "Festival Express Drop • {{city}}",
		Description: "Dynamic page created live via Studio",
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_hdr", WidgetID: "w_store_header", Position: 0, Enabled: true},
			{SlotID: "slot_grid", WidgetID: "w_fresh_grid", Position: 1, Enabled: true},
		},
	}
	rec, env := doJSON(t, r, http.MethodPut, "/api/v1/studio/layouts/festival-drop", newLayout, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("PUT layout draft failed: %d %s", rec.Code, rec.Body.String())
	}

	// 2. Validate draft
	rec, env = doJSON(t, r, http.MethodPost, "/api/v1/studio/layouts/festival-drop/validate", nil, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("validate failed: %d %s", rec.Code, rec.Body.String())
	}
	var valResp struct {
		Valid bool `json:"valid"`
	}
	_ = json.Unmarshal(env.Data, &valResp)
	if !valResp.Valid {
		t.Fatalf("expected layout draft to validate cleanly: %s", string(env.Data))
	}

	// 3. Publish v1
	rec, env = doJSON(t, r, http.MethodPost, "/api/v1/studio/layouts/festival-drop/publish", map[string]string{
		"actor":   "qa-tester",
		"comment": "Initial festival launch",
	}, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("publish v1 failed: %d %s", rec.Code, rec.Body.String())
	}

	// 4. Immediately serve live on KMP runtime endpoint!
	rec, env = doJSON(t, r, http.MethodGet, "/api/v1/screens/festival-drop?city=Pune", nil, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("GET newly published screen failed: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(string(env.Data), "Festival Express Drop") || !strings.Contains(string(env.Data), "Pune") {
		t.Fatalf("unexpected live page payload: %s", string(env.Data))
	}

	// 5. Update title -> Publish v2 -> Rollback to v1 (becomes v3)
	newLayout.Title = "Festival Drop V2"
	_, _ = doJSON(t, r, http.MethodPut, "/api/v1/studio/layouts/festival-drop", newLayout, nil)
	_, _ = doJSON(t, r, http.MethodPost, "/api/v1/studio/layouts/festival-drop/publish", map[string]string{"comment": "v2"}, nil)

	rec, env = doJSON(t, r, http.MethodPost, "/api/v1/studio/layouts/festival-drop/rollback", map[string]any{
		"version": 1,
		"actor":   "qa-tester",
	}, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("rollback failed: %d %s", rec.Code, rec.Body.String())
	}
	if rec.Code == http.StatusOK {
		recLive, envLive := doJSON(t, r, http.MethodGet, "/api/v1/screens/festival-drop", nil, nil)
		if recLive.Header().Get("X-Layout-Version") != "3" {
			t.Fatalf("expected rolled-back snapshot version 3, got %q", recLive.Header().Get("X-Layout-Version"))
		}
		if !strings.Contains(string(envLive.Data), "Festival Express Drop") {
			t.Fatalf("expected v1 title restored after rollback: %s", string(envLive.Data))
		}
	}

	// 6. Verify changelog, splash, tokens, and express order action
	rec, env = doJSON(t, r, http.MethodGet, "/api/v1/studio/changelog?limit=20", nil, nil)
	if rec.Code != http.StatusOK || !env.Success || !strings.Contains(string(env.Data), "rollback") {
		t.Fatalf("expected rollback audit entry in changelog: %s", string(env.Data))
	}

	rec, env = doJSON(t, r, http.MethodGet, "/api/v1/config/splash?segment=vip", nil, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("GET splash failed: %d %s", rec.Code, rec.Body.String())
	}

	rec, env = doJSON(t, r, http.MethodGet, "/api/v1/catalog/tokens?segment=vip", nil, nil)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("GET tokens failed: %d %s", rec.Code, rec.Body.String())
	}

	rec, env = doJSON(t, r, http.MethodPost, "/api/v1/orders/express", map[string]any{"items": 2}, nil)
	if rec.Code != http.StatusOK || !env.Success || !strings.Contains(string(env.Data), "celebration") {
		t.Fatalf("POST /api/v1/orders/express failed: %d %s", rec.Code, rec.Body.String())
	}
}

func TestEmbeddedStudioDocsAndHealthEndpoints(t *testing.T) {
	t.Parallel()
	r := setupTestRouter(t)

	for _, tc := range []struct {
		path   string
		status int
		substr string
	}{
		{"/healthz", http.StatusOK, ""},
		{"/readyz", http.StatusOK, ""},
		{"/metrics", http.StatusOK, ""},
		{"/openapi.json", http.StatusOK, "Abeta Dynamic Layout Engine API"},
		{"/docs", http.StatusOK, "scalar"},
		{"/studio", http.StatusOK, "Abeta Layout Studio"},
		{"/studio/app.css", http.StatusOK, ""},
		{"/studio/app.js", http.StatusOK, ""},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Errorf("GET %s: expected status %d, got %d", tc.path, tc.status, rec.Code)
		}
		if tc.substr != "" && !strings.Contains(rec.Body.String(), tc.substr) {
			t.Errorf("GET %s: expected body to contain %q", tc.path, tc.substr)
		}
	}
}
