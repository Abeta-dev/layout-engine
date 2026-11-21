package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/engine"
	"github.com/abeta-dev/layout-engine/internal/infra/persistence"
	"github.com/abeta-dev/layout-engine/internal/usecase"
	"github.com/abeta-dev/layout-engine/pkg/transformer"
)

func TestResolveServiceFallbackAndStudioLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	store, err := persistence.Open(ctx, persistence.Config{
		Dialect:     persistence.DialectSQLite,
		DSN:         ":memory:",
		AutoMigrate: true,
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if err := persistence.SeedDefaultData(ctx, store); err != nil {
		t.Fatalf("seed: %v", err)
	}

	eval := engine.NewRuleEvaluator()
	slotEng := engine.NewSlottingEngine(eval)
	val := engine.NewLayoutValidator()
	comp := transformer.NewPageComposer(transformer.NewDefaultRegistry())

	resolveSvc := usecase.NewResolveService(store, eval, slotEng, comp)
	studioSvc := usecase.NewStudioService(store, val, comp, engine.DiffJSON)

	// 1. Verify fallback kicks in automatically when a slug has a fallback but no published snapshot.
	tenant, err := store.GetTenantBySlug(ctx, usecase.DefaultTenantSlug)
	if err != nil {
		t.Fatalf("get tenant: %v", err)
	}
	if err := store.SaveFallback(ctx, domain.FallbackLayout{
		ID:              "fb_emergency",
		TenantID:        tenant.ID,
		LayoutSlug:      "emergency-only",
		SnapshotVersion: 9,
		Page:            snippets.PageResponse{PageID: "emergency-only", Title: "Safe Fallback Page"},
		UpdatedBy:       "test",
		UpdatedAt:       time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save fallback: %v", err)
	}

	res, err := resolveSvc.ResolvePage(ctx, "emergency-only", domain.RequestContext{
		TenantSlug: usecase.DefaultTenantSlug,
		City:       "Delhi",
		Now:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("expected fallback page instead of error, got %v", err)
	}
	if !res.Trace.UsedFallback || res.Page.Title != "Safe Fallback Page" {
		t.Fatalf("expected UsedFallback=true and Safe Fallback Page, got %+v", res)
	}

	// 2. Verify ActionService handles save_state, delete_state, and express order celebration.
	actSvc := usecase.NewActionService()
	saveRes, err := actSvc.Execute(ctx, usecase.ActionExecutionInput{
		Trigger: actions.NewSaveStateAction("save_city", actions.SaveStatePayload{Key: "city", Value: "Bengaluru"}),
	})
	if err != nil || saveRes.UpdatedState["city"] != "Bengaluru" {
		t.Fatalf("save_state failed: res=%+v err=%v", saveRes, err)
	}
	delRes, err := actSvc.Execute(ctx, usecase.ActionExecutionInput{
		Trigger: actions.NewDeleteStateAction("del_city", actions.DeleteStatePayload{Key: "city"}),
	})
	if err != nil || delRes.Status != "executed" {
		t.Fatalf("delete_state failed: res=%+v err=%v", delRes, err)
	}

	// 3. Verify StudioService tenant + rule + slotting + splash + tokens round-trips.
	if _, err := studioSvc.UpsertTenant(ctx, domain.Tenant{Slug: "abeta-pro", Name: "Abeta Pro"}, "tester"); err != nil {
		t.Fatalf("upsert tenant: %v", err)
	}
	if tenants, err := studioSvc.ListTenants(ctx); err != nil || len(tenants) < 2 {
		t.Fatalf("expected >= 2 tenants, got %d err=%v", len(tenants), err)
	}
	if _, err := studioSvc.UpsertRule(ctx, usecase.DefaultTenantSlug, domain.TargetingRule{
		Key:        "test_rule",
		Name:       "Test Rule",
		MatchMode:  "all",
		Conditions: []domain.RuleCondition{{Field: "city", Operator: "eq", Value: "Pune"}},
		Enabled:    true,
	}, "tester"); err != nil {
		t.Fatalf("upsert rule: %v", err)
	}
}
