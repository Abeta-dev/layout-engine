package persistence

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

func openTestStore(t *testing.T) *SQLStore {
	t.Helper()
	ctx := context.Background()
	store, err := Open(ctx, Config{
		Dialect:     DialectSQLite,
		DSN:         ":memory:",
		AutoMigrate: true,
	})
	if err != nil {
		t.Fatalf("open sqlite memory store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	return store
}

func TestOpenConfigAndMigrationsIdempotent(t *testing.T) {
	ctx := context.Background()

	// Unsupported dialect fails fast.
	if _, err := Open(ctx, Config{Dialect: "oracle", DSN: ":memory:"}); err == nil {
		t.Fatal("expected error for unsupported dialect")
	}
	// Missing postgres DSN fails fast.
	if _, err := Open(ctx, Config{Dialect: DialectPostgres, DSN: ""}); err == nil {
		t.Fatal("expected error for empty postgres DSN")
	}

	// Default dialect is SQLite :memory:.
	store := openTestStore(t)
	if store.Dialect() != DialectSQLite {
		t.Fatalf("expected dialect %q, got %q", DialectSQLite, store.Dialect())
	}
	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping failed: %v", err)
	}

	// Double-migrate must be idempotent and keep exactly 1 applied migration row.
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}
	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 schema_migrations row, got %d", count)
	}

	// File-backed SQLite enables WAL and persists across reopen.
	dbPath := filepath.Join(t.TempDir(), "layout_engine.db")
	fileStore, err := Open(ctx, Config{Dialect: DialectSQLite, DSN: dbPath, AutoMigrate: true})
	if err != nil {
		t.Fatalf("open file sqlite: %v", err)
	}
	if _, err := fileStore.UpsertTenant(ctx, domain.Tenant{
		ID:     "ten_persist",
		Slug:   "persist-shop",
		Name:   "Persist Shop",
		Active: true,
	}); err != nil {
		t.Fatalf("upsert tenant on file sqlite: %v", err)
	}
	if err := fileStore.Close(); err != nil {
		t.Fatalf("close file sqlite: %v", err)
	}

	reopened, err := Open(ctx, Config{Dialect: DialectSQLite, DSN: dbPath, AutoMigrate: true})
	if err != nil {
		t.Fatalf("reopen file sqlite: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	got, err := reopened.GetTenantBySlug(ctx, "persist-shop")
	if err != nil || got.ID != "ten_persist" {
		t.Fatalf("expected persisted tenant ten_persist, got %+v (err=%v)", got, err)
	}
}

func TestRebindPostgresPlaceholders(t *testing.T) {
	cases := []struct {
		name     string
		dialect  Dialect
		in       string
		expected string
	}{
		{
			name:     "sqlite leaves placeholders untouched",
			dialect:  DialectSQLite,
			in:       "SELECT * FROM widgets WHERE tenant_id = ? AND id IN (?, ?)",
			expected: "SELECT * FROM widgets WHERE tenant_id = ? AND id IN (?, ?)",
		},
		{
			name:     "postgres converts sequential placeholders",
			dialect:  DialectPostgres,
			in:       "SELECT * FROM widgets WHERE tenant_id = ? AND id IN (?, ?) LIMIT ?",
			expected: "SELECT * FROM widgets WHERE tenant_id = $1 AND id IN ($2, $3) LIMIT $4",
		},
		{
			name:     "postgres preserves question marks inside single-quoted literals",
			dialect:  DialectPostgres,
			in:       "SELECT 'what?' AS q, 'it''s?' AS q2 FROM layouts WHERE tenant_id = ? AND slug = ?",
			expected: "SELECT 'what?' AS q, 'it''s?' AS q2 FROM layouts WHERE tenant_id = $1 AND slug = $2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rebindQuery(tc.dialect, tc.in)
			if got != tc.expected {
				t.Fatalf("rebind mismatch:\n got: %s\nwant: %s", got, tc.expected)
			}
		})
	}
}

func TestEntityStoresCRUD(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	// TenantStore
	if _, err := store.GetTenantBySlug(ctx, "missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing tenant slug, got %v", err)
	}
	if _, err := store.GetTenantByID(ctx, "missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing tenant id, got %v", err)
	}

	created, err := store.UpsertTenant(ctx, domain.Tenant{
		Slug:        "demo-market",
		Name:        "Demo Market",
		Description: "Initial description",
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert tenant: %v", err)
	}
	updated, err := store.UpsertTenant(ctx, domain.Tenant{
		ID:          created.ID,
		Slug:        "demo-market",
		Name:        "Demo Market v2",
		Description: "Updated",
		Active:      true,
	})
	if err != nil {
		t.Fatalf("update tenant: %v", err)
	}
	if updated.Name != "Demo Market v2" || !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Fatalf("tenant update mismatch: got %+v, original created %v", updated, created.CreatedAt)
	}
	if _, err := store.UpsertTenant(ctx, domain.Tenant{
		ID:     "ten_other_conflict",
		Slug:   "demo-market",
		Name:   "Conflicting",
		Active: true,
	}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate tenant slug, got %v", err)
	}
	tenants, err := store.ListTenants(ctx)
	if err != nil || len(tenants) != 1 {
		t.Fatalf("expected 1 tenant, got %d (err=%v)", len(tenants), err)
	}

	// TemplateStore
	if _, err := store.GetTemplateByKey(ctx, created.ID, "missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing template, got %v", err)
	}
	tmpl, err := store.UpsertTemplate(ctx, domain.WidgetTemplate{
		TenantID:     created.ID,
		Key:          "tmpl_banner",
		Name:         "Promo Banner",
		SnippetType:  snippets.SnippetTypeBanner,
		DefaultProps: map[string]any{"title": "Hi"},
		Schema:       map[string]any{"version": "v1"},
		Tags:         []string{"promo"},
	})
	if err != nil {
		t.Fatalf("upsert template: %v", err)
	}
	tmpl2, err := store.UpsertTemplate(ctx, domain.WidgetTemplate{
		TenantID:     created.ID,
		Key:          "tmpl_banner",
		Name:         "Promo Banner v2",
		SnippetType:  snippets.SnippetTypeBanner,
		DefaultProps: map[string]any{"title": "Updated"},
	})
	if err != nil || tmpl2.ID != tmpl.ID || tmpl2.Name != "Promo Banner v2" {
		t.Fatalf("expected upsert template to preserve ID %s, got %+v (err=%v)", tmpl.ID, tmpl2, err)
	}
	tmpls, err := store.ListTemplates(ctx, created.ID)
	if err != nil || len(tmpls) != 1 {
		t.Fatalf("expected 1 template, got %d (err=%v)", len(tmpls), err)
	}

	// WidgetStore
	if _, err := store.GetWidgetByID(ctx, created.ID, "w_missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing widget, got %v", err)
	}
	if err := store.DeleteWidget(ctx, created.ID, "w_missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting missing widget, got %v", err)
	}

	w1, err := store.UpsertWidget(ctx, domain.Widget{
		ID:          "w_one",
		TenantID:    created.ID,
		Slug:        "widget-one",
		Name:        "Widget One",
		SnippetType: snippets.SnippetTypeHeader,
		Props:       map[string]any{"title": "Hello"},
		Tracking:    map[string]any{"event_name": "viewed"},
		Motion:      map[string]any{"enter": "fade_in"},
		Tags:        []string{"top"},
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert w1: %v", err)
	}
	w2, err := store.UpsertWidget(ctx, domain.Widget{
		ID:          "w_two",
		TenantID:    created.ID,
		Slug:        "widget-two",
		Name:        "Widget Two",
		SnippetType: snippets.SnippetTypeSpacer,
		Props:       map[string]any{"spacer": map[string]any{"height_dp": 12}},
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert w2: %v", err)
	}
	batch, err := store.GetWidgetsByIDs(ctx, created.ID, []string{w1.ID, w2.ID, w1.ID})
	if err != nil || len(batch) != 2 || batch[w1.ID].Motion["enter"] != "fade_in" {
		t.Fatalf("unexpected GetWidgetsByIDs result: %+v (err=%v)", batch, err)
	}
	emptyBatch, err := store.GetWidgetsByIDs(ctx, created.ID, nil)
	if err != nil || len(emptyBatch) != 0 {
		t.Fatalf("expected empty batch map, got %+v (err=%v)", emptyBatch, err)
	}
	if err := store.DeleteWidget(ctx, created.ID, w2.ID); err != nil {
		t.Fatalf("delete w2: %v", err)
	}
	remainingWidgets, err := store.ListWidgets(ctx, created.ID)
	if err != nil || len(remainingWidgets) != 1 {
		t.Fatalf("expected 1 widget after delete, got %d (err=%v)", len(remainingWidgets), err)
	}

	// RuleStore
	startHr, endHr := 2, 8
	rHigh, err := store.UpsertRule(ctx, domain.TargetingRule{
		ID:         "rule_high",
		TenantID:   created.ID,
		Key:        "high_pri",
		Name:       "High Priority",
		MatchMode:  "all",
		Conditions: []domain.RuleCondition{{Field: "segment", Operator: "eq", Value: "vip"}},
		Schedule:   &domain.ScheduleWindow{StartHourUTC: &startHr, EndHourUTC: &endHr, DaysOfWeek: []int{1, 2, 3}},
		Priority:   100,
		Enabled:    true,
	})
	if err != nil {
		t.Fatalf("upsert rule_high: %v", err)
	}
	_, err = store.UpsertRule(ctx, domain.TargetingRule{
		ID:       "rule_low",
		TenantID: created.ID,
		Key:      "low_pri",
		Name:     "Low Priority",
		Priority: 10,
		Enabled:  true,
	})
	if err != nil {
		t.Fatalf("upsert rule_low: %v", err)
	}
	rules, err := store.ListRules(ctx, created.ID)
	if err != nil || len(rules) != 2 || rules[0].ID != rHigh.ID || rules[0].Schedule == nil || *rules[0].Schedule.StartHourUTC != 2 {
		t.Fatalf("unexpected rules ordering/decode: %+v (err=%v)", rules, err)
	}
	if err := store.DeleteRule(ctx, created.ID, "rule_low"); err != nil {
		t.Fatalf("delete rule_low: %v", err)
	}
	if err := store.DeleteRule(ctx, created.ID, "rule_low"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting already-deleted rule, got %v", err)
	}
}

func TestLayoutsSlottingSnapshotsFallbackAndGovernance(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	tenant, err := store.UpsertTenant(ctx, domain.Tenant{
		ID:     "ten_gov",
		Slug:   "gov-tenant",
		Name:   "Governance Tenant",
		Active: true,
	})
	if err != nil {
		t.Fatalf("upsert tenant: %v", err)
	}

	// LayoutStore
	layout, err := store.UpsertLayout(ctx, domain.Layout{
		TenantID: tenant.ID,
		Slug:     "landing",
		Title:    "Landing Page",
		Status:   domain.LayoutStatusDraft,
		Slots: []domain.LayoutSlot{
			{SlotID: "s0", WidgetID: "w_hero", Position: 0, Pinned: true, Enabled: true},
		},
		PageConfig: domain.LayoutPageConfig{
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode: snippets.LayoutModeCentered,
				MaxWidthDp: atoms.IntPtr(960),
			},
		},
		Tags: []string{"landing"},
	})
	if err != nil {
		t.Fatalf("upsert layout: %v", err)
	}
	loadedLayout, err := store.GetLayoutBySlug(ctx, tenant.ID, "landing")
	if err != nil || loadedLayout.ID != layout.ID || loadedLayout.PageConfig.AdaptiveLayout == nil {
		t.Fatalf("get layout mismatch: %+v (err=%v)", loadedLayout, err)
	}

	// SlottingStore
	_, err = store.UpsertSlottingRule(ctx, domain.SlottingRule{
		ID:             "sr_low",
		TenantID:       tenant.ID,
		LayoutSlug:     "landing",
		Name:           "Low Priority Insert",
		Action:         domain.SlottingActionInsert,
		TargetPosition: 1,
		WidgetID:       "w_promo_low",
		Priority:       20,
		Enabled:        true,
	})
	if err != nil {
		t.Fatalf("upsert sr_low: %v", err)
	}
	_, err = store.UpsertSlottingRule(ctx, domain.SlottingRule{
		ID:             "sr_high",
		TenantID:       tenant.ID,
		LayoutSlug:     "landing",
		Name:           "High Priority Insert",
		Action:         domain.SlottingActionInsert,
		TargetPosition: 1,
		WidgetID:       "w_promo_high",
		Priority:       90,
		Enabled:        true,
	})
	if err != nil {
		t.Fatalf("upsert sr_high: %v", err)
	}
	slotRules, err := store.ListSlottingRules(ctx, tenant.ID, "landing")
	if err != nil || len(slotRules) != 2 || slotRules[0].ID != "sr_high" {
		t.Fatalf("expected sr_high first, got %+v (err=%v)", slotRules, err)
	}
	if err := store.DeleteSlottingRule(ctx, tenant.ID, "sr_low"); err != nil {
		t.Fatalf("delete slotting rule: %v", err)
	}

	// SnapshotStore versioning, immutability, and newest-first ordering
	if _, err := store.GetLatestSnapshot(ctx, tenant.ID, "landing"); !errors.Is(err, domain.ErrNoPublishedSnapshot) || !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNoPublishedSnapshot/ErrNotFound, got %v", err)
	}
	for v := 1; v <= 3; v++ {
		if err := store.SaveSnapshot(ctx, domain.PublishedSnapshot{
			TenantID:    tenant.ID,
			LayoutID:    layout.ID,
			LayoutSlug:  "landing",
			Version:     v,
			Title:       layout.Title,
			Slots:       layout.Slots,
			PublishedBy: "tester",
			Comment:     "release",
		}); err != nil {
			t.Fatalf("save snapshot v%d: %v", v, err)
		}
	}
	// Duplicate version must fail with ErrConflict (immutable snapshots).
	if err := store.SaveSnapshot(ctx, domain.PublishedSnapshot{
		TenantID:   tenant.ID,
		LayoutID:   layout.ID,
		LayoutSlug: "landing",
		Version:    2,
		Title:      "Overwrite Attempt",
	}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict overwriting snapshot v2, got %v", err)
	}
	latest, err := store.GetLatestSnapshot(ctx, tenant.ID, "landing")
	if err != nil || latest.Version != 3 {
		t.Fatalf("expected latest snapshot v3, got %+v (err=%v)", latest, err)
	}
	v1, err := store.GetSnapshotByVersion(ctx, tenant.ID, "landing", 1)
	if err != nil || v1.Version != 1 {
		t.Fatalf("expected snapshot v1, got %+v (err=%v)", v1, err)
	}
	snaps, err := store.ListSnapshots(ctx, tenant.ID, "landing")
	if err != nil || len(snaps) != 3 || snaps[0].Version != 3 || snaps[1].Version != 2 || snaps[2].Version != 1 {
		t.Fatalf("expected snapshots ordered [3,2,1], got %+v (err=%v)", snaps, err)
	}

	// FallbackStore
	if _, err := store.GetFallback(ctx, tenant.ID, "landing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing fallback, got %v", err)
	}
	fbPage := snippets.NewPage("landing", "Landing Safe Fallback").
		AddSnippet(snippets.NewHeaderSnippet("fb_hdr", atoms.NewText("Safe Mode", theming.TypographyTitleMedium, theming.ColorTextPrimary)).Build()).
		Build()
	if err := store.SaveFallback(ctx, domain.FallbackLayout{
		TenantID:        tenant.ID,
		LayoutSlug:      "landing",
		SnapshotVersion: 3,
		Page:            fbPage,
		UpdatedBy:       "tester",
	}); err != nil {
		t.Fatalf("save fallback: %v", err)
	}
	gotFB, err := store.GetFallback(ctx, tenant.ID, "landing")
	if err != nil || gotFB.SnapshotVersion != 3 || gotFB.Page.PageID != "landing" || len(gotFB.Page.Snippets) != 1 {
		t.Fatalf("unexpected fallback: %+v (err=%v)", gotFB, err)
	}

	// ConfigStore (Splash + TokenOverride)
	_, err = store.UpsertSplashConfig(ctx, domain.SplashConfig{
		TenantID:           tenant.ID,
		Key:                "default_splash",
		BackgroundColorHex: "#0B6E4F",
		DurationMs:         1200,
		IsDefault:          true,
		Enabled:            true,
	})
	if err != nil {
		t.Fatalf("upsert default splash: %v", err)
	}
	_, err = store.UpsertSplashConfig(ctx, domain.SplashConfig{
		TenantID:           tenant.ID,
		Key:                "vip_splash",
		BackgroundColorHex: "#1A1423",
		DurationMs:         1500,
		RuleID:             "rule_vip",
		IsDefault:          false,
		Enabled:            true,
	})
	if err != nil {
		t.Fatalf("upsert vip splash: %v", err)
	}
	splashes, err := store.ListSplashConfigs(ctx, tenant.ID)
	if err != nil || len(splashes) != 2 || splashes[0].Key != "vip_splash" || !splashes[1].IsDefault {
		t.Fatalf("unexpected splash list: %+v (err=%v)", splashes, err)
	}

	_, err = store.UpsertTokenOverride(ctx, domain.TokenOverride{
		TenantID:    tenant.ID,
		Key:         "vip_override",
		ThemePreset: "gold",
		Tokens:      map[string]string{"brand.primary": "#B45309"},
		RuleID:      "rule_vip",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("upsert vip token override: %v", err)
	}
	_, err = store.UpsertTokenOverride(ctx, domain.TokenOverride{
		TenantID:    tenant.ID,
		Key:         "base_override",
		ThemePreset: "fresh",
		Tokens:      map[string]string{"brand.primary": "#0B6E4F"},
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("upsert base token override: %v", err)
	}
	overrides, err := store.ListTokenOverrides(ctx, tenant.ID)
	if err != nil || len(overrides) != 2 || overrides[0].Key != "base_override" || overrides[1].Key != "vip_override" {
		t.Fatalf("expected unscoped base_override before rule-targeted override, got %+v (err=%v)", overrides, err)
	}

	// ChangeLogStore newest-first ordering (including exact-second vs sub-second timestamps!)
	baseTime := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	entries := []domain.ChangeLogEntry{
		{ID: "chg_1", TenantID: tenant.ID, EntityType: "layout", EntityID: "landing", Action: "create", Actor: "alice", Summary: "first", CreatedAt: baseTime},
		{ID: "chg_2", TenantID: tenant.ID, EntityType: "layout", EntityID: "landing", Action: "edit", Actor: "alice", Summary: "second", CreatedAt: baseTime.Add(250 * time.Millisecond), Diff: []domain.DiffItem{{Path: "/title", Op: "replace", Before: "A", After: "B"}}},
		{ID: "chg_3", TenantID: tenant.ID, EntityType: "layout", EntityID: "landing", Action: "publish", Actor: "bob", Summary: "third", CreatedAt: baseTime.Add(500 * time.Millisecond)},
	}
	for _, e := range entries {
		if err := store.AppendChangeLog(ctx, e); err != nil {
			t.Fatalf("append changelog %s: %v", e.ID, err)
		}
	}
	logs, err := store.ListChangeLogs(ctx, tenant.ID, 2)
	if err != nil || len(logs) != 2 || logs[0].ID != "chg_3" || logs[1].ID != "chg_2" || len(logs[1].Diff) != 1 {
		t.Fatalf("expected newest 2 changelogs [chg_3, chg_2] with diff, got %+v (err=%v)", logs, err)
	}

	// Delete layout cleans up cleanly.
	if err := store.DeleteLayout(ctx, tenant.ID, "landing"); err != nil {
		t.Fatalf("delete layout: %v", err)
	}
	if err := store.DeleteLayout(ctx, tenant.ID, "landing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting already-deleted layout, got %v", err)
	}
}
