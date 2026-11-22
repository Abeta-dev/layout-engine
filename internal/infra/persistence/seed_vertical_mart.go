package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// ── Abeta Mart (Multi-Category Marketplace) ───────────────────────────────────

func buildMartWorkspaceSpec(now time.Time) seedWorkspaceSpec {
	const tid = "tenant_abeta_mart"
	return seedWorkspaceSpec{
		Tenant: domain.Tenant{
			ID:          tid,
			Slug:        "abeta-mart",
			Name:        "Abeta Mart — Multi-Category Marketplace",
			Description: "Multi-category electronics, audio, home, and fashion marketplace with lightning deals",
			Active:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		TemplateIDPrefix: "mart",
		Widgets:          seedMartWidgets(tid, now),
		Rules: []domain.TargetingRule{
			{
				ID: "rule_mart_vip", TenantID: tid, Key: "mart_vip_segment",
				Name: "Mart VIP Lightning Access", Description: "Unlocks VIP lightning deals banner on mart-home.",
				MatchMode: "all", Conditions: []domain.RuleCondition{{Field: "segment", Operator: "eq", Value: "vip"}},
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		SlottingRules: []domain.SlottingRule{
			{
				ID: "slotrule_mart_vip_lightning", TenantID: tid, LayoutSlug: "mart-home",
				Name: "Inject VIP Lightning Deal Banner", Action: domain.SlottingActionInsert,
				TargetPosition: 2, WidgetID: "w_mart_vip_lightning_banner", RuleID: "rule_mart_vip",
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		Layouts: seedMartLayouts(tid, now),
		SplashConfigs: []domain.SplashConfig{
			{
				ID: "splash_default_mart", TenantID: tid, Key: "default_splash",
				ImageURL: "https://cdn.abeta.dev/mart/splash-default.png", BackgroundColorHex: "#1E293B",
				DurationMs: 1400, IsDefault: true, Enabled: true, UpdatedAt: now,
			},
		},
		TokenOverrides: []domain.TokenOverride{
			{
				ID: "tok_default_mart", TenantID: tid, Key: "default_mart", ThemePreset: "abeta-mart",
				Tokens: map[string]string{
					string(theming.ColorBrandPrimary):   "#B45309",
					string(theming.ColorBrandContainer): "#FEF3C7",
					string(theming.ColorAccentPrimary):  "#0F172A",
				},
				Enabled: true, UpdatedAt: now,
			},
		},
		ChangeLogID:      "chg_seed_abeta_mart",
		ChangeLogSummary: "Bootstrapped Abeta Mart marketplace workspace with 2 published layouts and 21 templates",
	}
}

func seedMartWidgets(tenantID string, now time.Time) []domain.Widget {
	vipToast := actions.NewShowToastAction("toast_mart_vip", actions.ShowToastPayload{
		Message: "VIP Lightning 15% savings applied to eligible tech & audio!", DurationMs: 2200, Variant: "success",
	})
	vipCTA := atoms.NewButton("btn_mart_vip", "Unlock 15% Off", "primary", &vipToast)
	vipCTA.Size = "sm"
	vipLayout := seedCardLayout()
	vipLayout.BgColor = atoms.ColorRef(theming.ColorWarningContainer)

	buyNow := actions.NewCallAPIAction("mart_buy_now", actions.CallAPIPayload{
		Endpoint: "/api/v1/mart/orders", Method: "POST",
		Body: map[string]any{"product_id": "mart_anc", "vertical": "abeta-mart"},
	})
	shareAct := actions.NewShareAction("share_mart_anc", actions.SharePayload{
		Title: "Abeta Pulse ANC Headphones",
		Text:  "44 dB adaptive ANC headphones for $149 on Abeta Mart!",
		URL:   "https://app.abeta.dev/screens/mart-product-detail",
	})
	buyBtn := atoms.NewButton("btn_mart_hero_buy", "Buy Now · $149", "primary", &buyNow)
	shareBtn := atoms.NewButton("btn_mart_hero_share", "Share", "outline", &shareAct)

	specRows := []snippets.SnippetEnvelope{
		snippets.NewActionRowSnippet("mart_spec_anc", atoms.NewText("Noise Cancellation", theming.TypographyTitleSmall, theming.ColorTextPrimary)).
			WithParentID("w_mart_specs_accordion").
			WithSubtitle(atoms.NewText("Hybrid 44 dB adaptive ANC + ambient transparency mode", theming.TypographyBodySmall, theming.ColorTextSecondary)).
			WithLeadingIcon(atoms.NewIcon("check", 20, theming.ColorSuccess)).
			WithLayout(*seedCardLayout()).Build(),
		snippets.NewActionRowSnippet("mart_spec_battery", atoms.NewText("Battery & USB-C Fast Charge", theming.TypographyTitleSmall, theming.ColorTextPrimary)).
			WithParentID("w_mart_specs_accordion").
			WithSubtitle(atoms.NewText("42h continuous playback · 10-min charge delivers 5h", theming.TypographyBodySmall, theming.ColorTextSecondary)).
			WithLeadingIcon(atoms.NewIcon("bolt", 20, theming.ColorBrandPrimary)).
			WithLayout(*seedCardLayout()).Build(),
		snippets.NewActionRowSnippet("mart_spec_bt", atoms.NewText("Connectivity & Warranty", theming.TypographyTitleSmall, theming.ColorTextPrimary)).
			WithParentID("w_mart_specs_accordion").
			WithSubtitle(atoms.NewText("Bluetooth 5.4 Multipoint · 2-year Abeta Care replacement", theming.TypographyBodySmall, theming.ColorTextSecondary)).
			WithLeadingIcon(atoms.NewIcon("verified", 20, theming.ColorInfo)).
			WithLayout(*seedCardLayout()).Build(),
	}

	products := []seedProductSpec{
		{
			widgetID: "w_mart_prod_anc", slug: "mart-prod-anc", productID: "mart_anc",
			title: "Pulse ANC Headset", subtitle: "44 dB ANC · 42h battery", price: "$149.00", originalPrice: "$215.00",
			unitPriceCents: 14900, badge: "-30%", badgeColor: theming.ColorWarningContainer, rating: "4.9 ★",
			initialQty: 1, delayMs: 0, imageNamespace: "mart", onClickRoute: "mart-product-detail",
		},
		{
			widgetID: "w_mart_prod_lamp", slug: "mart-prod-lamp", productID: "mart_lamp",
			title: "Nordic Desk Lamp", subtitle: "CRI 97+ · Touch dimmer", price: "$59.00", originalPrice: "$79.00",
			unitPriceCents: 5900, badge: "HOME PICK", badgeColor: theming.ColorBrandContainer, rating: "4.8 ★",
			initialQty: 0, delayMs: 60, imageNamespace: "mart", onClickRoute: "mart-product-detail",
		},
		{
			widgetID: "w_mart_prod_kb", slug: "mart-prod-kb", productID: "mart_kb",
			title: "Loom Mechanical 75%", subtitle: "Gasket mount · Tri-mode", price: "$119.00", originalPrice: "$145.00",
			unitPriceCents: 11900, badge: "NEW DROP", badgeColor: theming.ColorInfoContainer, rating: "4.9 ★",
			initialQty: 0, delayMs: 120, imageNamespace: "mart", onClickRoute: "mart-product-detail",
		},
		{
			widgetID: "w_mart_prod_hoodie", slug: "mart-prod-hoodie", productID: "mart_hoodie",
			title: "Merino Knit Hoodie", subtitle: "Breathable all-season weave", price: "$84.00", originalPrice: "$98.00",
			unitPriceCents: 8400, badge: "FASHION", badgeColor: theming.ColorSuccessContainer, rating: "4.7 ★",
			initialQty: 0, delayMs: 180, imageNamespace: "mart", onClickRoute: "mart-product-detail",
		},
	}

	widgets := []domain.Widget{
		buildSeedSearchBarWidget(
			tenantID, "w_mart_search_bar", "mart-search-bar", "Abeta Mart Search Bar",
			"mart_search", "Search electronics, audio, home, fashion...", "mart-home", now,
		),
		buildSeedChipGroupWidget(
			tenantID, "w_mart_category_chips", "mart-category-chips", "Abeta Mart Category Chips",
			"mart_categories", "mart-home",
			[]seedChipSpec{
				{id: "all", label: "All", selected: true},
				{id: "electronics", label: "Electronics"},
				{id: "audio", label: "Audio"},
				{id: "home", label: "Home"},
				{id: "fashion", label: "Fashion"},
			}, now,
		),
		{
			ID: "w_mart_vip_lightning_banner", TenantID: tenantID, Slug: "mart-vip-lightning-banner",
			Name: "VIP Lightning Deal Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("VIP Early Access • Lightning Deals in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Same-day dispatch + extra 15% member savings on flagship audio & electronics.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("bolt", 24, theming.ColorWarning)),
				Tag:    ptrTag(atoms.NewTag("VIP DEAL", theming.ColorSurfacePrimary)),
				CTA:    &vipCTA,
				Layout: vipLayout,
			}),
			Tags: []string{"banner", "vip", "slotting"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_mart_lightning_carousel", TenantID: tenantID, Slug: "mart-lightning-carousel",
			Name: "Lightning Deals Carousel", TemplateKey: "tmpl_carousel", SnippetType: snippets.SnippetTypeCarousel,
			Props: map[string]any{
				"title":               structToMap(atoms.NewText("Lightning Deals", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				"item_width_dp":       220,
				"item_width_by_width": map[string]any{"compact": 220, "medium": 250, "expanded": 270},
				"spacing_dp":          12,
				"child_widget_ids":    []string{"w_mart_prod_anc", "w_mart_prod_kb", "w_mart_prod_lamp"},
				"layout":              structToMap(seedHorizontalPaddingLayout()),
			},
			Tracking: map[string]any{"event_name": "carousel_viewed", "properties": map[string]any{"carousel_id": "w_mart_lightning_carousel"}},
			Tags:     []string{"carousel", "deals"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_mart_products_grid", TenantID: tenantID, Slug: "mart-products-grid",
			Name: "Trending Marketplace Grid", TemplateKey: "tmpl_grid", SnippetType: snippets.SnippetTypeGrid,
			Props: map[string]any{
				"title":            structToMap(atoms.NewText("Trending on Abeta Mart", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				"columns":          2,
				"columns_by_width": map[string]any{"compact": 2, "medium": 3, "expanded": 4},
				"spacing_dp":       12,
				"child_widget_ids": []string{"w_mart_prod_anc", "w_mart_prod_lamp", "w_mart_prod_kb", "w_mart_prod_hoodie"},
				"layout":           structToMap(seedHorizontalPaddingLayout()),
			},
			Tracking: map[string]any{"event_name": "grid_viewed", "properties": map[string]any{"grid_id": "w_mart_products_grid"}},
			Tags:     []string{"grid", "marketplace"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_mart_detail_hero", TenantID: tenantID, Slug: "mart-detail-hero",
			Name: "Product Detail Hero Card", TemplateKey: "tmpl_card_v1", SnippetType: snippets.SnippetTypeCard,
			Props: structToMap(snippets.CardSnippetData{
				Title:       atoms.NewText("Abeta Pulse ANC Headphones", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("4.9 ★ (1,284 reviews) · $149.00", theming.TypographyTitleMedium, theming.ColorTextBrand)),
				Description: ptrText(atoms.NewText("Studio 40mm drivers with adaptive 44 dB ANC and same-day dispatch in {{city}}.", theming.TypographyBodyMedium, theming.ColorTextSecondary)),
				Media: &atoms.ImageAtom{
					URL: "https://cdn.abeta.dev/mart/pulse-anc.png", ContentDescription: "Abeta Pulse ANC Headphones",
					HeightDp: atoms.IntPtr(172), ScaleMode: "crop", CornerRadius: atoms.UniformRadius(theming.RadiusMD.Dp()),
				},
				Badge:           ptrTag(atoms.NewTag("LIGHTNING DEAL", theming.ColorWarningContainer)),
				PrimaryButton:   &buyBtn,
				SecondaryButton: &shareBtn,
				Layout:          seedCardLayout(),
			}),
			Tags: []string{"product-detail", "card"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_mart_specs_accordion", TenantID: tenantID, Slug: "mart-specs-accordion",
			Name: "Product Technical Specs Accordion", TemplateKey: "tmpl_accordion", SnippetType: snippets.SnippetTypeAccordion,
			Props: structToMap(snippets.AccordionSnippetData{
				Title:             atoms.NewText("Specifications & Warranty", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Subtitle:          ptrText(atoms.NewText("Drivers, battery life & multipoint wireless", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				InitiallyExpanded: true,
				Snippets:          specRows,
				Layout:            seedCardLayout(),
			}),
			Tags: []string{"accordion", "specs"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
	}

	for _, sp := range products {
		widgets = append(widgets, buildSeedProductWidget(tenantID, now, sp))
	}
	return widgets
}

func seedMartLayouts(tenantID string, now time.Time) []domain.Layout {
	buyAction := actions.NewCallAPIAction("mart_bottom_buy", actions.CallAPIPayload{
		Endpoint: "/api/v1/mart/orders", Method: "POST",
		Body: map[string]any{"product_id": "mart_anc", "vertical": "abeta-mart"},
	})
	buyBtn := atoms.NewButton("btn_mart_bottom_buy", "Buy Now · $149.00", "primary", &buyAction)
	buyBtn.Layout = &atoms.LayoutConfig{FillMaxWidth: atoms.BoolPtr(true)}
	pullMart := actions.NewRefreshPageAction("pull_mart_home", actions.RefreshPagePayload{PageID: "mart-home", ShowLoader: true})

	return []domain.Layout{
		{
			ID: "lay_mart_home", TenantID: tenantID, Slug: "mart-home",
			Title: "Abeta Mart — Marketplace", Description: "Multi-category storefront with search, filters, lightning deals carousel, and responsive product grid.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar: buildSeedVerticalToolbar("Abeta Mart", "Same-day dispatch · {{city}}", "back_mart_home"),
				FloatingCart: buildSeedFloatingCart(
					"1 item · $149.00", "SAME DAY", "mart_checkout", "Checkout Now",
					"/api/v1/mart/orders", map[string]any{"vertical": "abeta-mart"},
				),
				AdaptiveLayout: buildSeedCenteredAdaptive(1120),
				PullToRefresh:  &pullMart,
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_mart_search", WidgetID: "w_mart_search_bar", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_mart_chips", WidgetID: "w_mart_category_chips", Position: 1, Enabled: true},
				{SlotID: "slot_mart_carousel", WidgetID: "w_mart_lightning_carousel", Position: 2, Enabled: true},
				{SlotID: "slot_mart_grid", WidgetID: "w_mart_products_grid", Position: 3, Enabled: true},
			},
			Tags: []string{"vertical", "marketplace"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_mart_product_detail", TenantID: tenantID, Slug: "mart-product-detail",
			Title: "Abeta Pulse ANC Headphones", Description: "Product detail screen with hero showcase, collapsible technical specs, and bottom bar CTA.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar: buildSeedVerticalToolbar("Product Detail", "Abeta Mart · In Stock", "back_mart_detail"),
				BottomBar: &snippets.BottomBarSpec{
					PrimaryButton: &buyBtn,
					HelperText:    ptrText(atoms.NewText("Free same-day delivery · 30-day returns", theming.TypographyLabelSmall, theming.ColorTextSecondary)),
				},
				AdaptiveLayout: buildSeedCenteredAdaptive(960),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_mart_detail_hero", WidgetID: "w_mart_detail_hero", Position: 0, Enabled: true},
				{SlotID: "slot_mart_detail_specs", WidgetID: "w_mart_specs_accordion", Position: 1, Enabled: true},
			},
			Tags: []string{"vertical", "product-detail"}, CreatedAt: now, UpdatedAt: now,
		},
	}
}
