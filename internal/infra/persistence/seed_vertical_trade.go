package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// ── Abeta Trade (B2B Wholesale & RFQ Sourcing) ────────────────────────────────

func buildTradeWorkspaceSpec(now time.Time) seedWorkspaceSpec {
	const tid = "tenant_abeta_trade"
	return seedWorkspaceSpec{
		Tenant: domain.Tenant{
			ID:          tid,
			Slug:        "abeta-trade",
			Name:        "Abeta Trade — B2B Wholesale & RFQ Sourcing",
			Description: "GST-verified wholesale manufacturers, MOQ tiered pricing, and direct RFQ sourcing",
			Active:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		TemplateIDPrefix: "trade",
		Widgets:          seedTradeWidgets(tid, now),
		Rules: []domain.TargetingRule{
			{
				ID: "rule_trade_enterprise", TenantID: tid, Key: "enterprise_buyers",
				Name: "Enterprise & VIP Wholesale Buyers", Description: "Matches enterprise or VIP procurement accounts.",
				MatchMode: "all", Conditions: []domain.RuleCondition{{Field: "segment", Operator: "in", Value: []string{"enterprise", "vip"}}},
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		SlottingRules: []domain.SlottingRule{
			{
				ID: "slotrule_trade_enterprise_desk", TenantID: tid, LayoutSlug: "trade-home",
				Name: "Inject Enterprise Procurement Desk Banner", Action: domain.SlottingActionInsert,
				TargetPosition: 2, WidgetID: "w_trade_enterprise_banner", RuleID: "rule_trade_enterprise",
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		Layouts: seedTradeLayouts(tid, now),
		SplashConfigs: []domain.SplashConfig{
			{
				ID: "splash_default_trade", TenantID: tid, Key: "default_splash",
				ImageURL: "https://cdn.abeta.dev/trade/splash-default.png", BackgroundColorHex: "#115E59",
				DurationMs: 1400, IsDefault: true, Enabled: true, UpdatedAt: now,
			},
		},
		TokenOverrides: []domain.TokenOverride{
			{
				ID: "tok_default_trade", TenantID: tid, Key: "default_trade", ThemePreset: "abeta-trade",
				Tokens: map[string]string{
					string(theming.ColorBrandPrimary):   "#0F766E",
					string(theming.ColorBrandContainer): "#CCFBF1",
					string(theming.ColorAccentPrimary):  "#0D9488",
				},
				Enabled: true, UpdatedAt: now,
			},
		},
		ChangeLogID:      "chg_seed_abeta_trade",
		ChangeLogSummary: "Bootstrapped Abeta Trade B2B wholesale workspace with 2 published layouts and 17 templates",
	}
}

func seedTradeWidgets(tenantID string, now time.Time) []domain.Widget {
	entToast := actions.NewShowToastAction("toast_trade_ent", actions.ShowToastPayload{
		Message: "Enterprise Net-30 credit desk activated for your account!", DurationMs: 2200, Variant: "success",
	})
	entCTA := atoms.NewButton("btn_trade_ent", "Open Credit Line", "primary", &entToast)
	entCTA.Size = "sm"
	entLayout := seedCardLayout()
	entLayout.BgColor = atoms.ColorRef(theming.ColorBrandContainer)

	rfqSuccess := actions.NewShowToastAction("toast_rfq_submitted", actions.ShowToastPayload{
		Message: "RFQ sent to Apex Corrugated! GST quote arrives within 2 hours.", DurationMs: 2600, Variant: "success",
	})
	submitRFQ := actions.NewCallAPIAction("trade_submit_rfq", actions.CallAPIPayload{
		Endpoint:  "/api/v1/trade/rfq",
		Method:    "POST",
		Body:      map[string]any{"supplier_id": "apex_corrugated", "vertical": "abeta-trade"},
		OnSuccess: &rfqSuccess,
	}).WithValidations(
		actions.ValidationRule{FieldID: "rfq_quantity", Rule: "required", Message: "Order quantity is required"},
		actions.ValidationRule{FieldID: "rfq_quantity", Rule: "regex", Value: `^\d+$`, Message: "Enter a valid numeric unit quantity"},
		actions.ValidationRule{FieldID: "rfq_company", Rule: "required", Message: "Company name is required"},
		actions.ValidationRule{FieldID: "rfq_email", Rule: "required", Message: "Business email is required"},
		actions.ValidationRule{FieldID: "rfq_email", Rule: "regex", Value: `^[^@\s]+@[^@\s]+\.[^@\s]+$`, Message: "Enter a valid business email"},
	)

	return []domain.Widget{
		buildSeedSearchBarWidget(
			tenantID, "w_trade_search_bar", "trade-search-bar", "Abeta Trade Wholesale Search Bar",
			"trade_search", "Search GST-verified manufacturers & wholesale lots...", "trade-home", now,
		),
		buildSeedChipGroupWidget(
			tenantID, "w_trade_category_chips", "trade-category-chips", "Abeta Trade Sourcing Filters",
			"trade_filters", "trade-home",
			[]seedChipSpec{
				{id: "all", label: "All", selected: true},
				{id: "gst_verified", label: "GST Verified"},
				{id: "electronics", label: "Electronics"},
				{id: "packaging", label: "Packaging"},
				{id: "textiles", label: "Textiles"},
			}, now,
		),
		{
			ID: "w_trade_enterprise_banner", TenantID: tenantID, Slug: "trade-enterprise-banner",
			Name: "Enterprise Procurement Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("Enterprise Procurement Desk • {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Dedicated sourcing managers, Net-30 trade credit & automated GST invoicing.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("verified", 24, theming.ColorBrandPrimary)),
				Tag:    ptrTag(atoms.NewTag("ENTERPRISE", theming.ColorSuccessContainer)),
				CTA:    &entCTA,
				Layout: entLayout,
			}),
			Tags: []string{"banner", "enterprise", "slotting"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_trade_sup_apex", slug: "trade-sup-apex", name: "Apex Corrugated Packaging Supplier",
			title: "Apex Corrugated Packaging", subtitle: "Packaging · 5-ply kraft shipping cartons",
			details: "$0.68/unit · MOQ 500 · 4-day factory dispatch", badge: "GST VERIFIED",
			badgeColor: theming.ColorSuccessContainer, ctaLabel: "Request Quote", route: "trade-supplier",
			tags: []string{"supplier", "packaging"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_trade_sup_nova", slug: "trade-sup-nova", name: "Nova Circuit Components Supplier",
			title: "Nova Circuit Components", subtitle: "Electronics · IoT telemetry boards",
			details: "$11.40/unit · MOQ 250 · ISO 9001 certified", badge: "MOQ 250",
			badgeColor: theming.ColorInfoContainer, ctaLabel: "View Catalog", route: "trade-supplier",
			tags: []string{"supplier", "electronics"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_trade_sup_veridian", slug: "trade-sup-veridian", name: "Veridian Loom Textiles Supplier",
			title: "Veridian Loom Textiles", subtitle: "Textiles · 320 GSM organic cotton rolls",
			details: "$4.90/meter · MOQ 300m · Export mill", badge: "EXPORT MILL",
			badgeColor: theming.ColorWarningContainer, ctaLabel: "Request Quote", route: "trade-supplier",
			tags: []string{"supplier", "textiles"},
		}),
		{
			ID: "w_trade_sup_header", TenantID: tenantID, Slug: "trade-sup-header",
			Name: "Verified Supplier Header", TemplateKey: "tmpl_header_v1", SnippetType: snippets.SnippetTypeHeader,
			Props: structToMap(snippets.HeaderSnippetData{
				Title:       atoms.NewText("Apex Corrugated Packaging", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("ISO 9001 & FSC Mill · Dispatching to {{city}}", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag: ptrTag(atoms.NewTag("GST VERIFIED", theming.ColorSuccessContainer)),
				Layout:      seedCardLayout(),
			}),
			Tags: []string{"header", "supplier"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_trade_moq_metric", TenantID: tenantID, Slug: "trade-moq-metric",
			Name: "Wholesale Bulk Tier Metric", TemplateKey: "tmpl_metric_card", SnippetType: snippets.SnippetTypeMetricCard,
			Props: structToMap(snippets.MetricCardSnippetData{
				Label: atoms.NewText("Wholesale Bulk Tier", theming.TypographyLabelMedium, theming.ColorTextSecondary),
				Value: atoms.NewText("$0.68 / unit", theming.TypographyHeadlineMedium, theming.ColorTextPrimary),
				Delta: ptrTag(atoms.NewTag("MOQ 500", theming.ColorSuccessContainer)),
				Progress: &atoms.ProgressAtom{
					Kind: "linear", Progress: atoms.Float64Ptr(0.98), Label: "98.6% on-time dispatch · 4-day lead time",
					Color: atoms.ColorRef(theming.ColorSuccess),
				},
				Layout: seedCardLayout(),
			}),
			Tags: []string{"metric", "wholesale"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_trade_rfq_form", TenantID: tenantID, Slug: "trade-rfq-form",
			Name: "Direct Wholesale RFQ Form", TemplateKey: "tmpl_form_v1", SnippetType: snippets.SnippetTypeForm,
			Props: structToMap(snippets.FormSnippetData{
				FormID:      "trade_rfq_form",
				Title:       ptrText(atoms.NewText("Request Direct Wholesale Quote (RFQ)", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				Description: ptrText(atoms.NewText("Receive tiered bulk pricing & GST invoice within 2 hours.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Fields: []atoms.TextFieldAtom{
					{FieldID: "rfq_quantity", Label: "Order Quantity (Units)", Placeholder: "500", InitialValue: "500", HelperText: "MOQ: 500 units", Required: true, Regex: `^\d+$`, KeyboardType: "number"},
					{FieldID: "rfq_company", Label: "Company Name", Placeholder: "Northstar Retail Pvt Ltd", Required: true, KeyboardType: "text"},
					{FieldID: "rfq_email", Label: "Business Email", Placeholder: "procurement@northstar.example", Required: true, Regex: `^[^@\s]+@[^@\s]+\.[^@\s]+$`, KeyboardType: "email"},
				},
				SubmitButton: atoms.NewButton("btn_submit_rfq", "Submit RFQ Quote", "primary", &submitRFQ),
				Layout:       seedCardLayout(),
			}),
			Tracking: map[string]any{"event_name": "rfq_form_viewed", "properties": map[string]any{"supplier_id": "apex_corrugated"}},
			Tags:     []string{"form", "rfq"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedTradeLayouts(tenantID string, now time.Time) []domain.Layout {
	pullTrade := actions.NewRefreshPageAction("pull_trade_home", actions.RefreshPagePayload{PageID: "trade-home", ShowLoader: true})
	return []domain.Layout{
		{
			ID: "lay_trade_home", TenantID: tenantID, Slug: "trade-home",
			Title: "Abeta Trade — B2B Wholesale", Description: "GST-verified B2B supplier directory with search, sourcing chips, and enterprise slotting.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Abeta Trade", "Verified B2B Wholesale & Direct RFQ", "back_trade_home"),
				AdaptiveLayout: buildSeedCenteredAdaptive(1120),
				PullToRefresh:  &pullTrade,
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_trade_search", WidgetID: "w_trade_search_bar", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_trade_chips", WidgetID: "w_trade_category_chips", Position: 1, Enabled: true},
				{SlotID: "slot_trade_apex", WidgetID: "w_trade_sup_apex", Position: 2, Enabled: true},
				{SlotID: "slot_trade_nova", WidgetID: "w_trade_sup_nova", Position: 3, Enabled: true},
				{SlotID: "slot_trade_veridian", WidgetID: "w_trade_sup_veridian", Position: 4, Enabled: true},
			},
			Tags: []string{"vertical", "b2b", "wholesale"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_trade_supplier", TenantID: tenantID, Slug: "trade-supplier",
			Title: "Apex Corrugated Packaging", Description: "Direct manufacturer storefront with MOQ pricing metrics and validated RFQ submission form.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Apex Corrugated", "GST Verified · Direct Manufacturer", "back_trade_sup"),
				AdaptiveLayout: buildSeedCenteredAdaptive(840),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_sup_hdr", WidgetID: "w_trade_sup_header", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_sup_moq", WidgetID: "w_trade_moq_metric", Position: 1, Enabled: true},
				{SlotID: "slot_sup_rfq", WidgetID: "w_trade_rfq_form", Position: 2, Enabled: true},
			},
			Tags: []string{"vertical", "supplier", "rfq"}, CreatedAt: now, UpdatedAt: now,
		},
	}
}
