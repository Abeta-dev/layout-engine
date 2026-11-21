package persistence

import (
	"context"
	"sync"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

func TestSeedDefaultDataIdempotentAndWireAuthentic(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	// Call twice to verify strict idempotency.
	if err := SeedDefaultData(ctx, store); err != nil {
		t.Fatalf("first SeedDefaultData failed: %v", err)
	}
	if err := SeedDefaultData(ctx, store); err != nil {
		t.Fatalf("second SeedDefaultData failed: %v", err)
	}

	tenant, err := store.GetTenantBySlug(ctx, "abeta-fresh")
	if err != nil {
		t.Fatalf("get seeded tenant: %v", err)
	}
	if tenant.Name != "Abeta Fresh Express" || !tenant.Active {
		t.Fatalf("unexpected seeded tenant: %+v", tenant)
	}

	templates, err := store.ListTemplates(ctx, tenant.ID)
	if err != nil || len(templates) != 14 {
		t.Fatalf("expected 14 templates (all KMP v1 SnippetTypes), got %d (err=%v)", len(templates), err)
	}

	widgets, err := store.ListWidgets(ctx, tenant.ID)
	if err != nil || len(widgets) < 16 {
		t.Fatalf("expected at least 16 seeded widgets, got %d (err=%v)", len(widgets), err)
	}

	rules, err := store.ListRules(ctx, tenant.ID)
	if err != nil || len(rules) != 5 {
		t.Fatalf("expected 5 seeded targeting rules, got %d (err=%v)", len(rules), err)
	}

	slottingRules, err := store.ListSlottingRules(ctx, tenant.ID, "store-home")
	if err != nil || len(slottingRules) != 2 {
		t.Fatalf("expected 2 slotting rules on store-home, got %d (err=%v)", len(slottingRules), err)
	}
	if slottingRules[0].Priority != 100 || slottingRules[0].WidgetID != "w_vip_flash_banner" {
		t.Fatalf("expected priority 100 VIP slotting rule first, got %+v", slottingRules[0])
	}

	expectedLayouts := []string{"store-home", "flash-sale", "home", "catalog", "form-demo", "containers", "watch-status"}
	layouts, err := store.ListLayouts(ctx, tenant.ID)
	if err != nil || len(layouts) != len(expectedLayouts) {
		t.Fatalf("expected %d seeded layouts, got %d (err=%v)", len(expectedLayouts), len(layouts), err)
	}

	for _, slug := range expectedLayouts {
		snap, err := store.GetLatestSnapshot(ctx, tenant.ID, slug)
		if err != nil || snap.Version != 1 {
			t.Fatalf("expected v1 snapshot for %s, got %+v (err=%v)", slug, snap, err)
		}
		if len(snap.ResolvedWidgets) == 0 {
			t.Fatalf("expected frozen widgets in snapshot for %s", slug)
		}

		fb, err := store.GetFallback(ctx, tenant.ID, slug)
		if err != nil || fb.Page.PageID != slug || len(fb.Page.Snippets) == 0 {
			t.Fatalf("expected hydrated fallback for %s, got %+v (err=%v)", slug, fb, err)
		}
		verifyPageEnvelopesDecode(t, fb.Page)
	}

	// Verify store-home specific KMP v1 rich chrome (FloatingCart, SplitPane, ProductCards inside Grid & Carousel).
	storeHomeFB, err := store.GetFallback(ctx, tenant.ID, "store-home")
	if err != nil {
		t.Fatalf("get store-home fallback: %v", err)
	}
	if storeHomeFB.Page.FloatingCart == nil || storeHomeFB.Page.AdaptiveLayout == nil || !storeHomeFB.Page.AdaptiveLayout.SplitPaneEnabled {
		t.Fatalf("expected floating cart and split-pane adaptive layout on store-home, got %+v", storeHomeFB.Page)
	}

	splashes, err := store.ListSplashConfigs(ctx, tenant.ID)
	if err != nil || len(splashes) != 2 {
		t.Fatalf("expected 2 splash configs, got %d (err=%v)", len(splashes), err)
	}
	overrides, err := store.ListTokenOverrides(ctx, tenant.ID)
	if err != nil || len(overrides) != 2 {
		t.Fatalf("expected 2 token overrides, got %d (err=%v)", len(overrides), err)
	}
	logs, err := store.ListChangeLogs(ctx, tenant.ID, 10)
	if err != nil || len(logs) == 0 {
		t.Fatalf("expected bootstrap changelog entry, got %d (err=%v)", len(logs), err)
	}
}

func TestConcurrentSQLiteAccessUnderRace(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	if err := SeedDefaultData(ctx, store); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	tenant, err := store.GetTenantBySlug(ctx, "abeta-fresh")
	if err != nil {
		t.Fatalf("get tenant: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 16)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := store.ListWidgets(ctx, tenant.ID); err != nil {
				errCh <- err
				return
			}
			if _, err := store.GetLatestSnapshot(ctx, tenant.ID, "store-home"); err != nil {
				errCh <- err
				return
			}
			if err := store.AppendChangeLog(ctx, domain.ChangeLogEntry{
				TenantID:   tenant.ID,
				EntityType: "layout",
				EntityID:   "store-home",
				Action:     "inspect",
				Actor:      "race-worker",
				Summary:    "concurrent check",
			}); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent store access failed: %v", err)
	}
}

func verifyPageEnvelopesDecode(t *testing.T, page snippets.PageResponse) {
	t.Helper()
	for _, env := range page.Snippets {
		verifySnippetEnvelopeDecode(t, env)
	}
}

func verifySnippetEnvelopeDecode(t *testing.T, env snippets.SnippetEnvelope) {
	t.Helper()
	switch env.SnippetType {
	case snippets.SnippetTypeHeader:
		d, err := snippets.DecodeSnippetData[snippets.HeaderSnippetData](env)
		if err != nil || d.Title.Text == "" {
			t.Fatalf("decode header %s: %v", env.SnippetID, err)
		}
	case snippets.SnippetTypeBanner:
		d, err := snippets.DecodeSnippetData[snippets.BannerSnippetData](env)
		if err != nil || d.Title.Text == "" {
			t.Fatalf("decode banner %s: %v", env.SnippetID, err)
		}
	case snippets.SnippetTypeMetricCard:
		d, err := snippets.DecodeSnippetData[snippets.MetricCardSnippetData](env)
		if err != nil || d.Value.Text == "" {
			t.Fatalf("decode metric_card %s: %v", env.SnippetID, err)
		}
	case snippets.SnippetTypeCard:
		d, err := snippets.DecodeSnippetData[snippets.CardSnippetData](env)
		if err != nil || d.Title.Text == "" {
			t.Fatalf("decode card_v1 %s: %v", env.SnippetID, err)
		}
	case snippets.SnippetTypeActionRow:
		d, err := snippets.DecodeSnippetData[snippets.ActionRowSnippetData](env)
		if err != nil || d.Title.Text == "" {
			t.Fatalf("decode action_row %s: %v", env.SnippetID, err)
		}
	case snippets.SnippetTypeProductCard:
		d, err := snippets.DecodeSnippetData[snippets.ProductCardSnippetData](env)
		if err != nil || d.ProductID == "" || d.UnitPriceCents <= 0 || d.Stepper.FieldID == "" {
			t.Fatalf("decode product_card %s: %v (data=%+v)", env.SnippetID, err, d)
		}
	case snippets.SnippetTypeGrid:
		d, err := snippets.DecodeSnippetData[snippets.GridSnippetData](env)
		if err != nil || d.Columns <= 0 || len(d.Items) == 0 {
			t.Fatalf("decode grid %s: %v (items=%d)", env.SnippetID, err, len(d.Items))
		}
		for _, item := range d.Items {
			verifySnippetEnvelopeDecode(t, item)
		}
	case snippets.SnippetTypeCarousel:
		d, err := snippets.DecodeSnippetData[snippets.CarouselSnippetData](env)
		if err != nil || len(d.Items) == 0 {
			t.Fatalf("decode carousel %s: %v (items=%d)", env.SnippetID, err, len(d.Items))
		}
		for _, item := range d.Items {
			verifySnippetEnvelopeDecode(t, item)
		}
	case snippets.SnippetTypeTabs:
		d, err := snippets.DecodeSnippetData[snippets.TabsSnippetData](env)
		if err != nil || len(d.Tabs) == 0 {
			t.Fatalf("decode tabs %s: %v", env.SnippetID, err)
		}
		for _, tab := range d.Tabs {
			for _, child := range tab.Snippets {
				verifySnippetEnvelopeDecode(t, child)
			}
		}
	case snippets.SnippetTypeAccordion:
		d, err := snippets.DecodeSnippetData[snippets.AccordionSnippetData](env)
		if err != nil || d.Title.Text == "" || len(d.Snippets) == 0 {
			t.Fatalf("decode accordion %s: %v", env.SnippetID, err)
		}
		for _, child := range d.Snippets {
			verifySnippetEnvelopeDecode(t, child)
		}
	case snippets.SnippetTypeForm:
		d, err := snippets.DecodeSnippetData[snippets.FormSnippetData](env)
		if err != nil || d.FormID == "" {
			t.Fatalf("decode form_v1 %s: %v", env.SnippetID, err)
		}
	}
}
