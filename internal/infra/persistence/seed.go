package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

const (
	defaultSeedTenantID   = "tenant_abeta_fresh"
	defaultSeedTenantSlug = "abeta-fresh"
	defaultSeedActor      = "system-seed"
)

// SeedDefaultData idempotently populates the Abeta Fresh Express quick-commerce
// starter workspace (1 tenant, 14 templates, 19 widgets, 5 targeting rules,
// 2 slotting rules, 5 published layouts + v1 snapshots + fallbacks, 2 splash
// configs, and 2 token overrides). It is safe to call on every server startup.
func SeedDefaultData(ctx context.Context, store domain.Store) error {
	if _, err := store.GetTenantBySlug(ctx, defaultSeedTenantSlug); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("persistence: check seed tenant: %w", err)
	}

	now := time.Now().UTC()

	tenant, err := store.UpsertTenant(ctx, domain.Tenant{
		ID:          defaultSeedTenantID,
		Slug:        defaultSeedTenantSlug,
		Name:        "Abeta Fresh Express",
		Description: "8-minute quick-commerce storefront powered by Abeta Dynamic Layout Engine",
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return fmt.Errorf("persistence: seed tenant: %w", err)
	}

	for _, tmpl := range seedTemplates(tenant.ID, now) {
		if _, err := store.UpsertTemplate(ctx, tmpl); err != nil {
			return fmt.Errorf("persistence: seed template %s: %w", tmpl.Key, err)
		}
	}

	widgets := seedWidgets(tenant.ID, now)
	widgetMap := make(map[string]domain.Widget, len(widgets))
	for _, w := range widgets {
		saved, err := store.UpsertWidget(ctx, w)
		if err != nil {
			return fmt.Errorf("persistence: seed widget %s: %w", w.ID, err)
		}
		widgetMap[saved.ID] = saved
	}

	for _, rule := range seedTargetingRules(tenant.ID, now) {
		if _, err := store.UpsertRule(ctx, rule); err != nil {
			return fmt.Errorf("persistence: seed rule %s: %w", rule.ID, err)
		}
	}

	for _, sr := range seedSlottingRules(tenant.ID, now) {
		if _, err := store.UpsertSlottingRule(ctx, sr); err != nil {
			return fmt.Errorf("persistence: seed slotting rule %s: %w", sr.ID, err)
		}
	}

	for _, layout := range seedLayouts(tenant.ID, now) {
		savedLayout, err := store.UpsertLayout(ctx, layout)
		if err != nil {
			return fmt.Errorf("persistence: seed layout %s: %w", layout.Slug, err)
		}

		frozen := collectFrozenWidgets(savedLayout.Slots, widgetMap)
		snap := domain.PublishedSnapshot{
			ID:              "snap_" + savedLayout.Slug + "_v1",
			TenantID:        tenant.ID,
			LayoutID:        savedLayout.ID,
			LayoutSlug:      savedLayout.Slug,
			Version:         1,
			Title:           savedLayout.Title,
			PageConfig:      savedLayout.PageConfig,
			Slots:           savedLayout.Slots,
			ResolvedWidgets: frozen,
			PublishedBy:     defaultSeedActor,
			Comment:         "Initial starter seed snapshot v1",
			PublishedAt:     now,
		}
		if err := store.SaveSnapshot(ctx, snap); err != nil {
			return fmt.Errorf("persistence: seed snapshot %s: %w", savedLayout.Slug, err)
		}

		fb := domain.FallbackLayout{
			ID:              "fb_" + savedLayout.Slug,
			TenantID:        tenant.ID,
			LayoutSlug:      savedLayout.Slug,
			SnapshotVersion: 1,
			Page:            composeSeedFallbackPage(savedLayout, frozen),
			UpdatedBy:       defaultSeedActor,
			UpdatedAt:       now,
		}
		if err := store.SaveFallback(ctx, fb); err != nil {
			return fmt.Errorf("persistence: seed fallback %s: %w", savedLayout.Slug, err)
		}
	}

	for _, sc := range seedSplashConfigs(tenant.ID, now) {
		if _, err := store.UpsertSplashConfig(ctx, sc); err != nil {
			return fmt.Errorf("persistence: seed splash %s: %w", sc.Key, err)
		}
	}

	for _, ov := range seedTokenOverrides(tenant.ID, now) {
		if _, err := store.UpsertTokenOverride(ctx, ov); err != nil {
			return fmt.Errorf("persistence: seed token override %s: %w", ov.Key, err)
		}
	}

	if err := store.AppendChangeLog(ctx, domain.ChangeLogEntry{
		ID:         "chg_seed_bootstrap",
		TenantID:   tenant.ID,
		EntityType: "tenant",
		EntityID:   tenant.ID,
		Action:     "seed",
		Actor:      defaultSeedActor,
		Summary:    "Bootstrapped Abeta Fresh Express starter workspace with 5 published layouts, 14 templates, and 19 widgets",
		CreatedAt:  now,
	}); err != nil {
		return fmt.Errorf("persistence: seed changelog: %w", err)
	}

	return nil
}

func seedTemplates(tenantID string, now time.Time) []domain.WidgetTemplate {
	sampleToast := actions.NewShowToastAction("tmpl_toast", actions.ShowToastPayload{
		Message:    "Sample action triggered",
		DurationMs: 1800,
		Variant:    "info",
	})
	sampleBtn := atoms.NewButton("btn_sample", "Action", "primary", &sampleToast)
	sampleStepper := atoms.NewStepper("qty_sample", "ADD", 0, &sampleToast, nil)
	sampleStepper.MinQuantity = 0
	sampleStepper.MaxQuantity = 10

	type tmplSpec struct {
		key         string
		name        string
		snippetType string
		desc        string
		props       any
		tags        []string
	}

	specs := []tmplSpec{
		{
			key:         "tmpl_header_v1",
			name:        "Section Header",
			snippetType: snippets.SnippetTypeHeader,
			desc:        "Screen or section header with title, subtitle, leading icon, trailing badge, and action button.",
			props: snippets.HeaderSnippetData{
				Title:       atoms.NewText("Section Title", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Supporting subtitle", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag: ptrTag(atoms.NewTag("LIVE", theming.ColorSuccessContainer)),
			},
			tags: []string{"layout", "header"},
		},
		{
			key:         "tmpl_banner_v1",
			name:        "Hero Promo Banner",
			snippetType: snippets.SnippetTypeBanner,
			desc:        "High-visibility promotional or status banner with icon, tag, and CTA button.",
			props: snippets.BannerSnippetData{
				Title: atoms.NewText("Fresh Drop", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:  ptrText(atoms.NewText("Delivered in 8 minutes.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:  ptrIcon(atoms.NewIcon("bolt", 24, theming.ColorBrandPrimary)),
				Tag:   ptrTag(atoms.NewTag("NEW", theming.ColorBrandContainer)),
				CTA:   &sampleBtn,
			},
			tags: []string{"promo", "hero"},
		},
		{
			key:         "tmpl_metric_card",
			name:        "KPI Metric Card",
			snippetType: snippets.SnippetTypeMetricCard,
			desc:        "Live SLA or KPI counter with delta pill and optional linear progress bar.",
			props: snippets.MetricCardSnippetData{
				Label: atoms.NewText("Delivery ETA", theming.TypographyLabelMedium, theming.ColorTextSecondary),
				Value: atoms.NewText("8 mins", theming.TypographyHeadlineMedium, theming.ColorTextPrimary),
				Delta: ptrTag(atoms.NewTag("-2 min", theming.ColorSuccessContainer)),
				Progress: &atoms.ProgressAtom{
					Kind:     "linear",
					Progress: atoms.Float64Ptr(0.85),
					Color:    atoms.ColorRef(theming.ColorSuccess),
				},
			},
			tags: []string{"telemetry", "kpi"},
		},
		{
			key:         "tmpl_card_v1",
			name:        "Content Card",
			snippetType: snippets.SnippetTypeCard,
			desc:        "Versatile content card supporting badge, description, media, and primary/secondary CTAs.",
			props: snippets.CardSnippetData{
				Title:         atoms.NewText("Featured Story", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Subtitle:      ptrText(atoms.NewText("Curated selection", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Badge:         ptrTag(atoms.NewTag("FEATURED", theming.ColorInfoContainer)),
				PrimaryButton: &sampleBtn,
			},
			tags: []string{"content", "card"},
		},
		{
			key:         "tmpl_action_row",
			name:        "Navigation Action Row",
			snippetType: snippets.SnippetTypeActionRow,
			desc:        "Interactive list row with leading icon, badge, chevron, and click action.",
			props: snippets.ActionRowSnippetData{
				Title:        atoms.NewText("Quick Action", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:     ptrText(atoms.NewText("Open target screen", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon:  ptrIcon(atoms.NewIcon("bolt", 22, theming.ColorBrandPrimary)),
				TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 20, theming.ColorTextTertiary)),
				OnClick:      &sampleToast,
			},
			tags: []string{"navigation", "list"},
		},
		{
			key:         "tmpl_form_v1",
			name:        "Interactive Form",
			snippetType: snippets.SnippetTypeForm,
			desc:        "Declarative input form supporting text fields, switches, checkboxes, sliders, and validation.",
			props: snippets.FormSnippetData{
				FormID: "sample_form",
				Title:  ptrText(atoms.NewText("Feedback Form", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				Fields: []atoms.TextFieldAtom{
					{FieldID: "name", Label: "Name", Placeholder: "Your name", Required: true, KeyboardType: "text"},
				},
				SubmitButton: sampleBtn,
			},
			tags: []string{"input", "form"},
		},
		{
			key:         "tmpl_empty_state",
			name:        "Empty State Placeholder",
			snippetType: snippets.SnippetTypeEmptyState,
			desc:        "Zero-data placeholder with illustration icon, message, and recovery button.",
			props: snippets.EmptyStateSnippetData{
				Title:        atoms.NewText("Nothing Here Yet", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Message:      ptrText(atoms.NewText("Check back shortly for fresh arrivals.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:         ptrIcon(atoms.NewIcon("inventory_2", 32, theming.ColorTextTertiary)),
				ActionButton: &sampleBtn,
			},
			tags: []string{"state", "empty"},
		},
		{
			key:         "tmpl_divider",
			name:        "Section Divider",
			snippetType: snippets.SnippetTypeDivider,
			desc:        "Horizontal divider separating content blocks.",
			props: snippets.DividerSnippetData{
				Divider: atoms.DividerAtom{
					ThicknessDp:  atoms.Float64Ptr(1),
					Color:        atoms.ColorRef(theming.ColorDivider),
					InsetStartDp: atoms.IntPtr(16),
				},
			},
			tags: []string{"layout", "separator"},
		},
		{
			key:         "tmpl_spacer",
			name:        "Vertical Spacer",
			snippetType: snippets.SnippetTypeSpacer,
			desc:        "Fixed dp whitespace block between snippets.",
			props: snippets.SpacerSnippetData{
				Spacer: atoms.SpacerAtom{HeightDp: atoms.IntPtr(16)},
			},
			tags: []string{"layout", "spacing"},
		},
		{
			key:         "tmpl_carousel",
			name:        "Responsive Carousel",
			snippetType: snippets.SnippetTypeCarousel,
			desc:        "Horizontally scrolling container with responsive item width by window breakpoint.",
			props: map[string]any{
				"title":         structToMap(atoms.NewText("Featured Carousel", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				"item_width_dp": 220,
				"item_width_by_width": map[string]any{
					"compact":  220,
					"medium":   240,
					"expanded": 260,
				},
				"spacing_dp":       12,
				"child_widget_ids": []string{},
			},
			tags: []string{"container", "responsive"},
		},
		{
			key:         "tmpl_grid",
			name:        "Responsive Grid",
			snippetType: snippets.SnippetTypeGrid,
			desc:        "Adaptive multi-column product or card grid keyed by compact/medium/expanded breakpoints.",
			props: map[string]any{
				"title":   structToMap(atoms.NewText("Product Grid", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				"columns": 2,
				"columns_by_width": map[string]any{
					"compact":  2,
					"medium":   3,
					"expanded": 4,
				},
				"spacing_dp":       12,
				"child_widget_ids": []string{},
			},
			tags: []string{"container", "responsive", "commerce"},
		},
		{
			key:         "tmpl_tabs",
			name:        "Category Tabs Container",
			snippetType: snippets.SnippetTypeTabs,
			desc:        "Tabbed section switching between nested snippet groups or widget references.",
			props: map[string]any{
				"selected_index": 0,
				"tabs": []map[string]any{
					{"id": "tab_1", "title": "Tab One", "widget_ids": []string{}},
				},
			},
			tags: []string{"container", "navigation"},
		},
		{
			key:         "tmpl_accordion",
			name:        "Collapsible Accordion",
			snippetType: snippets.SnippetTypeAccordion,
			desc:        "Expandable section wrapping FAQ rows, guarantees, or preference forms.",
			props: snippets.AccordionSnippetData{
				Title:             atoms.NewText("Frequently Asked", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				InitiallyExpanded: false,
				Snippets:          []snippets.SnippetEnvelope{},
			},
			tags: []string{"container", "faq"},
		},
		{
			key:         "tmpl_product_card",
			name:        "Quick-Commerce Product Card",
			snippetType: snippets.SnippetTypeProductCard,
			desc:        "Merchandising card with image, pricing, badge, rating, and live quantity stepper.",
			props: snippets.ProductCardSnippetData{
				ProductID:      "sample_sku",
				Title:          atoms.NewText("Fresh Product", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:       ptrText(atoms.NewText("250 g · Chilled", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Price:          atoms.NewText("$3.99", theming.TypographyLabelLarge, theming.ColorTextPrimary),
				UnitPriceCents: 399,
				Badge:          ptrTag(atoms.NewTag("FRESH", theming.ColorSuccessContainer)),
				Stepper:        sampleStepper,
			},
			tags: []string{"commerce", "product"},
		},
	}

	out := make([]domain.WidgetTemplate, 0, len(specs))
	for _, sp := range specs {
		out = append(out, domain.WidgetTemplate{
			ID:           sp.key,
			TenantID:     tenantID,
			Key:          sp.key,
			Name:         sp.name,
			SnippetType:  sp.snippetType,
			Description:  sp.desc,
			DefaultProps: structToMap(sp.props),
			Schema: map[string]any{
				"snippet_type": sp.snippetType,
				"wire_version": "v1",
			},
			Tags:      sp.tags,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return out
}

func seedTargetingRules(tenantID string, now time.Time) []domain.TargetingRule {
	startMorning := 0
	endMorning := 6

	return []domain.TargetingRule{
		{
			ID:          "rule_vip_segment",
			TenantID:    tenantID,
			Key:         "vip_segment",
			Name:        "VIP Member Segment",
			Description: "Matches callers belonging to the VIP loyalty segment.",
			MatchMode:   "all",
			Conditions: []domain.RuleCondition{
				{Field: "segment", Operator: "eq", Value: "vip"},
			},
			Priority:  100,
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "rule_bengaluru_city",
			TenantID:    tenantID,
			Key:         "bengaluru_city",
			Name:        "Bengaluru Metro Launch",
			Description: "Matches Bengaluru / Bangalore quick-commerce delivery zones.",
			MatchMode:   "all",
			Conditions: []domain.RuleCondition{
				{Field: "city", Operator: "in", Value: []string{"bengaluru", "bangalore"}},
			},
			Priority:  90,
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "rule_morning_window",
			TenantID:    tenantID,
			Key:         "morning_window",
			Name:        "Morning Breakfast Window (00:00–06:00 UTC)",
			Description: "Active daily during IST morning rush (5:30–11:30 AM IST).",
			MatchMode:   "all",
			Conditions:  []domain.RuleCondition{},
			Schedule: &domain.ScheduleWindow{
				StartHourUTC: &startMorning,
				EndHourUTC:   &endMorning,
			},
			Priority:  80,
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "rule_ios_or_android_v1",
			TenantID:    tenantID,
			Key:         "ios_or_android_v1",
			Name:        "Mobile App v1.0+ (iOS & Android)",
			Description: "Targets native iOS and Android KMP clients running version 1.0.0 or higher.",
			MatchMode:   "all",
			Conditions: []domain.RuleCondition{
				{Field: "platform", Operator: "in", Value: []string{"ios", "android"}},
				{Field: "app_version", Operator: "semver_gte", Value: "1.0.0"},
			},
			Priority:  70,
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "rule_rollout_50",
			TenantID:    tenantID,
			Key:         "rollout_50",
			Name:        "50% Deterministic Rollout",
			Description: "Consistent hash rollout targeting 50% of users by user_id.",
			MatchMode:   "all",
			Conditions: []domain.RuleCondition{
				{Field: "user_id", Operator: "bucket_pct", Value: 50},
			},
			Priority:  60,
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func seedSlottingRules(tenantID string, now time.Time) []domain.SlottingRule {
	return []domain.SlottingRule{
		{
			ID:             "slotrule_vip_flash_banner",
			TenantID:       tenantID,
			LayoutSlug:     "store-home",
			Name:           "Inject VIP Exclusive Banner at Slot 1",
			Action:         domain.SlottingActionInsert,
			TargetPosition: 1,
			WidgetID:       "w_vip_flash_banner",
			RuleID:         "rule_vip_segment",
			Priority:       100,
			Enabled:        true,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "slotrule_morning_banner",
			TenantID:       tenantID,
			LayoutSlug:     "store-home",
			Name:           "Inject Morning Breakfast Banner at Slot 1",
			Action:         domain.SlottingActionInsert,
			TargetPosition: 1,
			WidgetID:       "w_morning_banner",
			RuleID:         "rule_morning_window",
			Priority:       80,
			Enabled:        true,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}
}

func seedSplashConfigs(tenantID string, now time.Time) []domain.SplashConfig {
	return []domain.SplashConfig{
		{
			ID:                 "splash_default_fresh",
			TenantID:           tenantID,
			Key:                "default_splash",
			ImageURL:           "https://cdn.abeta.dev/fresh/splash-default.png",
			BackgroundColorHex: "#0B6E4F",
			DurationMs:         1400,
			IsDefault:          true,
			Enabled:            true,
			UpdatedAt:          now,
		},
		{
			ID:                 "splash_vip_fresh",
			TenantID:           tenantID,
			Key:                "vip_splash",
			ImageURL:           "https://cdn.abeta.dev/fresh/splash-vip.png",
			BackgroundColorHex: "#1A1423",
			DurationMs:         1600,
			LottieURL:          "https://cdn.abeta.dev/fresh/splash-vip.json",
			RuleID:             "rule_vip_segment",
			IsDefault:          false,
			Enabled:            true,
			UpdatedAt:          now,
		},
	}
}

func seedTokenOverrides(tenantID string, now time.Time) []domain.TokenOverride {
	return []domain.TokenOverride{
		{
			ID:          "tok_default_fresh",
			TenantID:    tenantID,
			Key:         "default_fresh",
			ThemePreset: "abeta-fresh",
			Tokens: map[string]string{
				string(theming.ColorBrandPrimary):   "#0B6E4F",
				string(theming.ColorBrandContainer): "#DCFCE7",
				string(theming.ColorAccentPrimary):  "#16A34A",
				string(theming.ColorSuccess):        "#15803D",
			},
			Enabled:   true,
			UpdatedAt: now,
		},
		{
			ID:          "tok_vip_gold_accent",
			TenantID:    tenantID,
			Key:         "vip_gold_accent",
			ThemePreset: "abeta-vip-gold",
			Tokens: map[string]string{
				string(theming.ColorBrandPrimary):   "#B45309",
				string(theming.ColorBrandContainer): "#FEF3C7",
				string(theming.ColorAccentPrimary):  "#D97706",
				string(theming.ColorTextBrand):      "#B45309",
			},
			RuleID:    "rule_vip_segment",
			Enabled:   true,
			UpdatedAt: now,
		},
	}
}

func structToMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}

func ptrText(t atoms.TextAtom) *atoms.TextAtom { return &t }
func ptrIcon(i atoms.IconAtom) *atoms.IconAtom { return &i }
func ptrTag(t atoms.TagAtom) *atoms.TagAtom    { return &t }
