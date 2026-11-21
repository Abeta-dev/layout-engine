package persistence

import (
	"context"
	"sync"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

type verticalSeedExpectation struct {
	slug               string
	name               string
	expectedLayouts    []string
	slottingLayoutSlug string
	minWidgets         int
	minRules           int
}

func TestSeedDefaultDataIdempotentAndWireAuthentic(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	// Call twice to verify strict idempotency across all 6 vertical workspaces.
	if err := SeedDefaultData(ctx, store); err != nil {
		t.Fatalf("first SeedDefaultData failed: %v", err)
	}
	if err := SeedDefaultData(ctx, store); err != nil {
		t.Fatalf("second SeedDefaultData failed: %v", err)
	}

	allTenants, err := store.ListTenants(ctx)
	if err != nil || len(allTenants) != 6 {
		t.Fatalf("expected 6 seeded vertical tenants, got %d (err=%v)", len(allTenants), err)
	}

	expectations := []verticalSeedExpectation{
		{
			slug: "abeta-fresh", name: "Abeta Fresh — Quick-Commerce Grocery",
			expectedLayouts:    []string{"store-home", "fresh-home", "flash-sale", "home", "catalog", "form-demo", "containers", "watch-status"},
			slottingLayoutSlug: "store-home", minWidgets: 21, minRules: 5,
		},
		{
			slug: "abeta-bites", name: "Abeta Bites — Food Delivery & Dining",
			expectedLayouts:    []string{"bites-home", "bites-restaurant", "bites-order-track"},
			slottingLayoutSlug: "bites-home", minWidgets: 15, minRules: 2,
		},
		{
			slug: "abeta-pulse", name: "Abeta Pulse — Events, Movies & Dining Tickets",
			expectedLayouts:    []string{"pulse-home", "pulse-event-detail"},
			slottingLayoutSlug: "pulse-home", minWidgets: 11, minRules: 1,
		},
		{
			slug: "abeta-mart", name: "Abeta Mart — Multi-Category Marketplace",
			expectedLayouts:    []string{"mart-home", "mart-product-detail"},
			slottingLayoutSlug: "mart-home", minWidgets: 10, minRules: 1,
		},
		{
			slug: "abeta-spark", name: "Abeta Spark — Swipe Discovery & Matches",
			expectedLayouts:    []string{"spark-deck", "spark-matches"},
			slottingLayoutSlug: "spark-deck", minWidgets: 8, minRules: 1,
		},
		{
			slug: "abeta-trade", name: "Abeta Trade — B2B Wholesale & RFQ Sourcing",
			expectedLayouts:    []string{"trade-home", "trade-supplier"},
			slottingLayoutSlug: "trade-home", minWidgets: 9, minRules: 1,
		},
	}

	for _, exp := range expectations {
		verifySeededVertical(t, ctx, store, exp)
	}

	verifyFreshAndVerticalSpecifics(t, ctx, store)
}

func verifySeededVertical(t *testing.T, ctx context.Context, store *SQLStore, exp verticalSeedExpectation) {
	t.Helper()

	tenant, err := store.GetTenantBySlug(ctx, exp.slug)
	if err != nil {
		t.Fatalf("[%s] get seeded tenant: %v", exp.slug, err)
	}
	if tenant.Name != exp.name || !tenant.Active {
		t.Fatalf("[%s] unexpected seeded tenant: %+v", exp.slug, tenant)
	}

	templates, err := store.ListTemplates(ctx, tenant.ID)
	if err != nil || len(templates) != 17 {
		t.Fatalf("[%s] expected 17 templates (all KMP v1 SnippetTypes), got %d (err=%v)", exp.slug, len(templates), err)
	}

	widgets, err := store.ListWidgets(ctx, tenant.ID)
	if err != nil || len(widgets) < exp.minWidgets {
		t.Fatalf("[%s] expected at least %d widgets, got %d (err=%v)", exp.slug, exp.minWidgets, len(widgets), err)
	}
	for _, w := range widgets {
		verifyTrackingStringProps(t, exp.slug+"/"+w.ID, w.Tracking)
	}

	rules, err := store.ListRules(ctx, tenant.ID)
	if err != nil || len(rules) < exp.minRules {
		t.Fatalf("[%s] expected at least %d rules, got %d (err=%v)", exp.slug, exp.minRules, len(rules), err)
	}

	slottingRules, err := store.ListSlottingRules(ctx, tenant.ID, exp.slottingLayoutSlug)
	if err != nil || len(slottingRules) == 0 {
		t.Fatalf("[%s] expected slotting rules on %s, got %d (err=%v)", exp.slug, exp.slottingLayoutSlug, len(slottingRules), err)
	}

	layouts, err := store.ListLayouts(ctx, tenant.ID)
	if err != nil || len(layouts) != len(exp.expectedLayouts) {
		t.Fatalf("[%s] expected %d layouts, got %d (err=%v)", exp.slug, len(exp.expectedLayouts), len(layouts), err)
	}

	for _, slug := range exp.expectedLayouts {
		snap, err := store.GetLatestSnapshot(ctx, tenant.ID, slug)
		if err != nil || snap.Version != 1 {
			t.Fatalf("[%s/%s] expected v1 snapshot, got %+v (err=%v)", exp.slug, slug, snap, err)
		}
		if len(snap.ResolvedWidgets) == 0 {
			t.Fatalf("[%s/%s] expected frozen widgets in snapshot", exp.slug, slug)
		}

		fb, err := store.GetFallback(ctx, tenant.ID, slug)
		if err != nil || fb.Page.PageID != slug || len(fb.Page.Snippets) == 0 {
			t.Fatalf("[%s/%s] expected hydrated fallback, got %+v (err=%v)", exp.slug, slug, fb, err)
		}
		verifyPageEnvelopesDecode(t, fb.Page)
	}

	splashes, err := store.ListSplashConfigs(ctx, tenant.ID)
	if err != nil || len(splashes) == 0 {
		t.Fatalf("[%s] expected splash configs, got %d (err=%v)", exp.slug, len(splashes), err)
	}
	overrides, err := store.ListTokenOverrides(ctx, tenant.ID)
	if err != nil || len(overrides) == 0 {
		t.Fatalf("[%s] expected token overrides, got %d (err=%v)", exp.slug, len(overrides), err)
	}
	logs, err := store.ListChangeLogs(ctx, tenant.ID, 10)
	if err != nil || len(logs) == 0 {
		t.Fatalf("[%s] expected bootstrap changelog entry, got %d (err=%v)", exp.slug, len(logs), err)
	}
}

func verifyFreshAndVerticalSpecifics(t *testing.T, ctx context.Context, store *SQLStore) {
	t.Helper()

	freshTenant, err := store.GetTenantBySlug(ctx, "abeta-fresh")
	if err != nil {
		t.Fatalf("get abeta-fresh: %v", err)
	}
	slottingRules, err := store.ListSlottingRules(ctx, freshTenant.ID, "store-home")
	if err != nil || len(slottingRules) != 2 {
		t.Fatalf("expected 2 slotting rules on store-home, got %d (err=%v)", len(slottingRules), err)
	}
	if slottingRules[0].Priority != 100 || slottingRules[0].WidgetID != "w_vip_flash_banner" {
		t.Fatalf("expected priority 100 VIP slotting rule first, got %+v", slottingRules[0])
	}

	storeHomeFB, err := store.GetFallback(ctx, freshTenant.ID, "store-home")
	if err != nil {
		t.Fatalf("get store-home fallback: %v", err)
	}
	if storeHomeFB.Page.FloatingCart == nil || storeHomeFB.Page.AdaptiveLayout == nil || !storeHomeFB.Page.AdaptiveLayout.SplitPaneEnabled {
		t.Fatalf("expected floating cart and split-pane adaptive layout on store-home, got %+v", storeHomeFB.Page)
	}

	freshHomeFB, err := store.GetFallback(ctx, freshTenant.ID, "fresh-home")
	if err != nil || freshHomeFB.Page.FloatingCart == nil {
		t.Fatalf("expected fresh-home fallback with floating cart, err=%v", err)
	}

	sparkTenant, err := store.GetTenantBySlug(ctx, "abeta-spark")
	if err != nil {
		t.Fatalf("get abeta-spark: %v", err)
	}
	sparkDeckFB, err := store.GetFallback(ctx, sparkTenant.ID, "spark-deck")
	if err != nil {
		t.Fatalf("get spark-deck fallback: %v", err)
	}
	foundDeck := false
	for _, env := range sparkDeckFB.Page.Snippets {
		if env.SnippetType == snippets.SnippetTypeSwipeDeck {
			deck, err := snippets.DecodeSnippetData[snippets.SwipeDeckSnippetData](env)
			if err != nil || len(deck.Cards) != 4 || deck.OnUndo == nil {
				t.Fatalf("expected 4 curated swipe cards + on_undo on spark-deck, got %+v (err=%v)", deck, err)
			}
			foundDeck = true
		}
	}
	if !foundDeck {
		t.Fatalf("expected swipe_deck snippet on spark-deck")
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

func verifyTrackingStringProps(t *testing.T, owner string, tracking map[string]any) {
	t.Helper()
	if len(tracking) == 0 {
		return
	}
	props, ok := tracking["properties"].(map[string]any)
	if !ok {
		return
	}
	for k, v := range props {
		if _, isStr := v.(string); !isStr {
			t.Fatalf("tracking.properties[%q] on %s must be string, got %T (%v)", k, owner, v, v)
		}
	}
}

func verifyPageEnvelopesDecode(t *testing.T, page snippets.PageResponse) {
	t.Helper()
	if page.Tracking != nil {
		for k, v := range page.Tracking.Properties {
			if _, isStr := v.(string); !isStr {
				t.Fatalf("page %s tracking.properties[%q] must be string, got %T (%v)", page.PageID, k, v, v)
			}
		}
	}
	for _, env := range page.Snippets {
		verifySnippetEnvelopeDecode(t, env)
	}
}

func verifySnippetEnvelopeDecode(t *testing.T, env snippets.SnippetEnvelope) {
	t.Helper()
	if env.Tracking != nil {
		for k, v := range env.Tracking.Properties {
			if _, isStr := v.(string); !isStr {
				t.Fatalf("snippet %s tracking.properties[%q] must be string, got %T (%v)", env.SnippetID, k, v, v)
			}
		}
	}

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
	case snippets.SnippetTypeSearchBar:
		d, err := snippets.DecodeSnippetData[snippets.SearchBarSnippetData](env)
		if err != nil || d.SearchID == "" || d.Placeholder == "" {
			t.Fatalf("decode search_bar %s: %v (data=%+v)", env.SnippetID, err, d)
		}
	case snippets.SnippetTypeChipGroup:
		d, err := snippets.DecodeSnippetData[snippets.ChipGroupSnippetData](env)
		if err != nil || d.GroupID == "" || len(d.Chips) == 0 {
			t.Fatalf("decode chip_group %s: %v (data=%+v)", env.SnippetID, err, d)
		}
	case snippets.SnippetTypeSwipeDeck:
		d, err := snippets.DecodeSnippetData[snippets.SwipeDeckSnippetData](env)
		if err != nil || d.DeckID == "" || len(d.Cards) == 0 {
			t.Fatalf("decode swipe_deck %s: %v (data=%+v)", env.SnippetID, err, d)
		}
		for _, card := range d.Cards {
			if card.CardID == "" || card.Title.Text == "" || card.OnLike == nil || card.OnNope == nil || card.OnSuperLike == nil {
				t.Fatalf("swipe_deck %s card incomplete: %+v", env.SnippetID, card)
			}
		}
	}
}
