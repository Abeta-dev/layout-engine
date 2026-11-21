package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

const (
	defaultSeedTenantID   = "tenant_abeta_fresh"
	defaultSeedTenantSlug = "abeta-fresh"
	defaultSeedActor      = "system-seed"
)

type seedWorkspaceSpec struct {
	Tenant           domain.Tenant
	TemplateIDPrefix string
	Widgets          []domain.Widget
	Rules            []domain.TargetingRule
	SlottingRules    []domain.SlottingRule
	Layouts          []domain.Layout
	SplashConfigs    []domain.SplashConfig
	TokenOverrides   []domain.TokenOverride
	ChangeLogID      string
	ChangeLogSummary string
}

// SeedDefaultData idempotently populates all 6 multi-tenant vertical workspaces
// (abeta-fresh, abeta-bites, abeta-pulse, abeta-mart, abeta-spark, abeta-trade),
// each with all 17 canonical KMP v1 widget templates, curated widgets, targeting
// and slotting rules, published layouts, immutable v1 snapshots, hydrated
// fallbacks, splash configs, and design token overrides. Safe on every startup.
func SeedDefaultData(ctx context.Context, store domain.Store) error {
	now := time.Now().UTC()
	workspaces := []seedWorkspaceSpec{
		buildFreshWorkspaceSpec(now),
		buildBitesWorkspaceSpec(now),
		buildPulseWorkspaceSpec(now),
		buildMartWorkspaceSpec(now),
		buildSparkWorkspaceSpec(now),
		buildTradeWorkspaceSpec(now),
	}

	for _, ws := range workspaces {
		if err := seedWorkspace(ctx, store, ws, now); err != nil {
			return err
		}
	}
	return nil
}

func seedWorkspace(ctx context.Context, store domain.Store, ws seedWorkspaceSpec, now time.Time) error {
	if _, err := store.GetTenantBySlug(ctx, ws.Tenant.Slug); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("persistence: check seed tenant %s: %w", ws.Tenant.Slug, err)
	}

	tenant, err := store.UpsertTenant(ctx, ws.Tenant)
	if err != nil {
		return fmt.Errorf("persistence: seed tenant %s: %w", ws.Tenant.Slug, err)
	}

	for _, tmpl := range seedTemplates(tenant.ID, ws.TemplateIDPrefix, now) {
		if _, err := store.UpsertTemplate(ctx, tmpl); err != nil {
			return fmt.Errorf("persistence: seed template %s/%s: %w", tenant.Slug, tmpl.Key, err)
		}
	}

	widgetMap := make(map[string]domain.Widget, len(ws.Widgets))
	for _, w := range ws.Widgets {
		saved, err := store.UpsertWidget(ctx, w)
		if err != nil {
			return fmt.Errorf("persistence: seed widget %s/%s: %w", tenant.Slug, w.ID, err)
		}
		widgetMap[saved.ID] = saved
	}

	for _, rule := range ws.Rules {
		if _, err := store.UpsertRule(ctx, rule); err != nil {
			return fmt.Errorf("persistence: seed rule %s/%s: %w", tenant.Slug, rule.ID, err)
		}
	}

	for _, sr := range ws.SlottingRules {
		if _, err := store.UpsertSlottingRule(ctx, sr); err != nil {
			return fmt.Errorf("persistence: seed slotting rule %s/%s: %w", tenant.Slug, sr.ID, err)
		}
	}

	snapPrefix := "snap_"
	fbPrefix := "fb_"
	if ws.TemplateIDPrefix != "" {
		snapPrefix = "snap_" + ws.TemplateIDPrefix + "_"
		fbPrefix = "fb_" + ws.TemplateIDPrefix + "_"
	}

	for _, layout := range ws.Layouts {
		savedLayout, err := store.UpsertLayout(ctx, layout)
		if err != nil {
			return fmt.Errorf("persistence: seed layout %s/%s: %w", tenant.Slug, layout.Slug, err)
		}

		frozen := collectFrozenWidgets(savedLayout.Slots, widgetMap)
		snap := domain.PublishedSnapshot{
			ID:              snapPrefix + savedLayout.Slug + "_v1",
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
			return fmt.Errorf("persistence: seed snapshot %s/%s: %w", tenant.Slug, savedLayout.Slug, err)
		}

		fb := domain.FallbackLayout{
			ID:              fbPrefix + savedLayout.Slug,
			TenantID:        tenant.ID,
			LayoutSlug:      savedLayout.Slug,
			SnapshotVersion: 1,
			Page:            composeSeedFallbackPage(savedLayout, frozen),
			UpdatedBy:       defaultSeedActor,
			UpdatedAt:       now,
		}
		if err := store.SaveFallback(ctx, fb); err != nil {
			return fmt.Errorf("persistence: seed fallback %s/%s: %w", tenant.Slug, savedLayout.Slug, err)
		}
	}

	for _, sc := range ws.SplashConfigs {
		if _, err := store.UpsertSplashConfig(ctx, sc); err != nil {
			return fmt.Errorf("persistence: seed splash %s/%s: %w", tenant.Slug, sc.Key, err)
		}
	}

	for _, ov := range ws.TokenOverrides {
		if _, err := store.UpsertTokenOverride(ctx, ov); err != nil {
			return fmt.Errorf("persistence: seed token override %s/%s: %w", tenant.Slug, ov.Key, err)
		}
	}

	if err := store.AppendChangeLog(ctx, domain.ChangeLogEntry{
		ID:         ws.ChangeLogID,
		TenantID:   tenant.ID,
		EntityType: "tenant",
		EntityID:   tenant.ID,
		Action:     "seed",
		Actor:      defaultSeedActor,
		Summary:    ws.ChangeLogSummary,
		CreatedAt:  now,
	}); err != nil {
		return fmt.Errorf("persistence: seed changelog %s: %w", tenant.Slug, err)
	}

	return nil
}

func buildFreshWorkspaceSpec(now time.Time) seedWorkspaceSpec {
	return seedWorkspaceSpec{
		Tenant: domain.Tenant{
			ID:          defaultSeedTenantID,
			Slug:        defaultSeedTenantSlug,
			Name:        "Abeta Fresh — Quick-Commerce Grocery",
			Description: "8-minute quick-commerce grocery storefront powered by Abeta Dynamic Layout Engine",
			Active:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		TemplateIDPrefix: "",
		Widgets:          seedWidgets(defaultSeedTenantID, now),
		Rules:            seedTargetingRules(defaultSeedTenantID, now),
		SlottingRules:    seedSlottingRules(defaultSeedTenantID, now),
		Layouts:          seedLayouts(defaultSeedTenantID, now),
		SplashConfigs:    seedSplashConfigs(defaultSeedTenantID, now),
		TokenOverrides:   seedTokenOverrides(defaultSeedTenantID, now),
		ChangeLogID:      "chg_seed_bootstrap",
		ChangeLogSummary: "Bootstrapped Abeta Fresh quick-commerce workspace with 8 published layouts, 17 templates, and 21 widgets",
	}
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
		{
			ID:             "slotrule_fresh_vip_banner",
			TenantID:       tenantID,
			LayoutSlug:     "fresh-home",
			Name:           "Inject VIP Exclusive Banner on Fresh Home",
			Action:         domain.SlottingActionInsert,
			TargetPosition: 2,
			WidgetID:       "w_vip_flash_banner",
			RuleID:         "rule_vip_segment",
			Priority:       100,
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
