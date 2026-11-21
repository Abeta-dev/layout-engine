package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// ── Abeta Bites (Food Delivery & Dining) ──────────────────────────────────────

func buildBitesWorkspaceSpec(now time.Time) seedWorkspaceSpec {
	const tid = "tenant_abeta_bites"
	return seedWorkspaceSpec{
		Tenant: domain.Tenant{
			ID:          tid,
			Slug:        "abeta-bites",
			Name:        "Abeta Bites — Food Delivery & Dining",
			Description: "Hot restaurant delivery, curated dining menus, and live rider tracking",
			Active:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		TemplateIDPrefix: "bites",
		Widgets:          seedBitesWidgets(tid, now),
		Rules:            seedBitesRules(tid, now),
		SlottingRules:    seedBitesSlottingRules(tid, now),
		Layouts:          seedBitesLayouts(tid, now),
		SplashConfigs: []domain.SplashConfig{
			{
				ID: "splash_default_bites", TenantID: tid, Key: "default_splash",
				ImageURL: "https://cdn.abeta.dev/bites/splash-default.png", BackgroundColorHex: "#B45309",
				DurationMs: 1400, IsDefault: true, Enabled: true, UpdatedAt: now,
			},
			{
				ID: "splash_vip_bites", TenantID: tid, Key: "vip_splash",
				ImageURL: "https://cdn.abeta.dev/bites/splash-vip.png", BackgroundColorHex: "#7C2D12",
				DurationMs: 1500, RuleID: "rule_bites_mumbai_or_vip", IsDefault: false, Enabled: true, UpdatedAt: now,
			},
		},
		TokenOverrides: []domain.TokenOverride{
			{
				ID: "tok_default_bites", TenantID: tid, Key: "default_bites", ThemePreset: "abeta-bites",
				Tokens: map[string]string{
					string(theming.ColorBrandPrimary):   "#D97706",
					string(theming.ColorBrandContainer): "#FEF3C7",
					string(theming.ColorAccentPrimary):  "#EA580C",
				},
				Enabled: true, UpdatedAt: now,
			},
		},
		ChangeLogID:      "chg_seed_abeta_bites",
		ChangeLogSummary: "Bootstrapped Abeta Bites food delivery workspace with 3 published layouts and 17 templates",
	}
}

func seedBitesWidgets(tenantID string, now time.Time) []domain.Widget {
	navRest := actions.NewNavigateAction("nav_bites_banner_rest", actions.NavigatePayload{Route: "/screens/bites-restaurant"})
	diningCTA := atoms.NewButton("btn_bites_dining", "Explore Menu", "primary", &navRest)
	diningCTA.Size = "sm"
	diningLayout := seedCardLayout()
	diningLayout.BgColor = atoms.ColorRef(theming.ColorBrandContainer)

	vipApply := actions.NewShowToastAction("toast_bites_vip", actions.ShowToastPayload{
		Message: "BITESVIP unlocked — free delivery on dinner orders!", DurationMs: 2200, Variant: "success",
	})
	vipCTA := atoms.NewButton("btn_bites_vip", "Apply BITESVIP", "primary", &vipApply)
	vipCTA.Size = "sm"
	vipLayout := seedCardLayout()
	vipLayout.BgColor = atoms.ColorRef(theming.ColorWarningContainer)

	dishes := []seedProductSpec{
		{
			widgetID: "w_bites_dish_biryani", slug: "bites-dish-biryani", productID: "dish_biryani",
			title: "Smoked Veg Biryani", subtitle: "Mint raita & mirchi salan", price: "$9.50", originalPrice: "$11.00",
			unitPriceCents: 950, badge: "CHEF PICK", badgeColor: theming.ColorSuccessContainer, rating: "4.9 ★",
			initialQty: 1, delayMs: 0, imageNamespace: "bites",
		},
		{
			widgetID: "w_bites_dish_paneer", slug: "bites-dish-paneer", productID: "dish_paneer",
			title: "Charred Paneer Tikka", subtitle: "Tandoor roasted · 6 pcs", price: "$8.25", originalPrice: "$9.50",
			unitPriceCents: 825, badge: "PURE VEG", badgeColor: theming.ColorInfoContainer, rating: "4.8 ★",
			initialQty: 1, delayMs: 60, imageNamespace: "bites",
		},
		{
			widgetID: "w_bites_dish_naan", slug: "bites-dish-naan", productID: "dish_naan",
			title: "Garlic Herb Naan", subtitle: "2 wood-fired buttered breads", price: "$3.50", originalPrice: "$4.20",
			unitPriceCents: 350, badge: "HOT", badgeColor: theming.ColorWarningContainer, rating: "4.7 ★",
			initialQty: 0, delayMs: 120, imageNamespace: "bites",
		},
		{
			widgetID: "w_bites_dish_lassi", slug: "bites-dish-lassi", productID: "dish_lassi",
			title: "Cardamom Mango Lassi", subtitle: "Chilled · 350 ml", price: "$4.25", originalPrice: "$4.95",
			unitPriceCents: 425, badge: "CHILLED", badgeColor: theming.ColorBrandContainer, rating: "4.9 ★",
			initialQty: 0, delayMs: 180, imageNamespace: "bites",
		},
	}

	widgets := []domain.Widget{
		buildSeedSearchBarWidget(
			tenantID, "w_bites_search_bar", "bites-search-bar", "Abeta Bites Search Bar",
			"bites_search", "Search biryani, pizza, burgers in {{city}}...", "bites-home", now,
		),
		buildSeedChipGroupWidget(
			tenantID, "w_bites_filter_chips", "bites-filter-chips", "Abeta Bites Dining Filters",
			"bites_filters", "bites-home",
			[]seedChipSpec{
				{id: "all", label: "All", selected: true},
				{id: "pure_veg", label: "Pure Veg"},
				{id: "rating_4", label: "Rating 4.0+"},
				{id: "fast_delivery", label: "Fast Delivery"},
			}, now,
		),
		{
			ID: "w_bites_dining_banner", TenantID: tenantID, Slug: "bites-dining-banner",
			Name: "Featured Dining Hero Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("Chef-Curated Dining in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Piping-hot biryani, wood-fired pizza & ramen delivered in insulated eco-bowls.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("star", 24, theming.ColorBrandPrimary)),
				Tag:    ptrTag(atoms.NewTag("24 MINS", theming.ColorSuccessContainer)),
				CTA:    &diningCTA,
				Layout: diningLayout,
			}),
			Tracking: map[string]any{"event_name": "banner_viewed", "properties": map[string]any{"vertical": "abeta-bites"}},
			Tags:     []string{"banner", "dining"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_bites_vip_dinner_banner", TenantID: tenantID, Slug: "bites-vip-dinner-banner",
			Name: "VIP Dinner Rush Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("VIP Dinner Rush • Free Delivery in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Priority kitchen prep and zero delivery fees for Mumbai & VIP members.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("bolt", 24, theming.ColorWarning)),
				Tag:    ptrTag(atoms.NewTag("VIP DINNER", theming.ColorSurfacePrimary)),
				CTA:    &vipCTA,
				Layout: vipLayout,
			}),
			Tracking: map[string]any{"event_name": "slotting_banner_viewed", "properties": map[string]any{"vertical": "abeta-bites", "slot": "vip_dinner"}},
			Tags:     []string{"banner", "vip", "slotting"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_bites_rest_saffron", slug: "bites-rest-saffron", name: "Saffron Hearth Kitchen Card",
			title: "Saffron Hearth Kitchen", subtitle: "North Indian · Biryani · Kebabs",
			details: "4.8 ★ · 24 mins · $14 for two · Free Delivery", badge: "TOP RATED",
			badgeColor: theming.ColorSuccessContainer, ctaLabel: "View Menu", route: "bites-restaurant",
			tags: []string{"restaurant", "card"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_bites_rest_cedar", slug: "bites-rest-cedar", name: "Cedar Crust Pizzeria Card",
			title: "Cedar Crust Pizzeria", subtitle: "Wood-Fired Pizza · Fresh Pasta",
			details: "4.7 ★ · 20 mins · $18 for two · Flat $3 OFF", badge: "FAST 20M",
			badgeColor: theming.ColorWarningContainer, ctaLabel: "View Menu", route: "bites-restaurant",
			tags: []string{"restaurant", "card"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_bites_rest_bamboo", slug: "bites-rest-bamboo", name: "Bamboo Noodle Foundry Card",
			title: "Bamboo Noodle Foundry", subtitle: "Pan-Asian · Hand-Pulled Ramen",
			details: "4.6 ★ · 28 mins · Pure Veg Options", badge: "PURE VEG+",
			badgeColor: theming.ColorInfoContainer, ctaLabel: "View Menu", route: "bites-restaurant",
			tags: []string{"restaurant", "card"},
		}),
		{
			ID: "w_bites_rest_header", TenantID: tenantID, Slug: "bites-rest-header",
			Name: "Restaurant Menu Header", TemplateKey: "tmpl_header_v1", SnippetType: snippets.SnippetTypeHeader,
			Props: structToMap(snippets.HeaderSnippetData{
				Title:       atoms.NewText("Saffron Hearth Kitchen", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Cooked fresh in {{city}} · Eco-insulated bowls", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag: ptrTag(atoms.NewTag("24 MINS", theming.ColorSuccessContainer)),
				Layout:      seedCardLayout(),
			}),
			Tags: []string{"header", "restaurant"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_bites_menu_grid", TenantID: tenantID, Slug: "bites-menu-grid",
			Name: "Signature Menu Product Grid", TemplateKey: "tmpl_grid", SnippetType: snippets.SnippetTypeGrid,
			Props: map[string]any{
				"title":            structToMap(atoms.NewText("Signature Dishes", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				"columns":          2,
				"columns_by_width": map[string]any{"compact": 2, "medium": 3, "expanded": 4},
				"spacing_dp":       12,
				"child_widget_ids": []string{"w_bites_dish_biryani", "w_bites_dish_paneer", "w_bites_dish_naan", "w_bites_dish_lassi"},
				"layout":           structToMap(seedHorizontalPaddingLayout()),
			},
			Tracking: map[string]any{"event_name": "grid_viewed", "properties": map[string]any{"grid_id": "w_bites_menu_grid"}},
			Tags:     []string{"grid", "menu"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_bites_track_banner", TenantID: tenantID, Slug: "bites-track-banner",
			Name: "Live Order ETA Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("Rider On The Way in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Order #AB-4821 sealed hot · Arriving in 14 mins.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("local_shipping", 24, theming.ColorBrandPrimary)),
				Tag:    ptrTag(atoms.NewTag("14 MINS", theming.ColorSuccessContainer)),
				Layout: diningLayout,
			}),
			Tags: []string{"tracking", "banner"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_bites_track_metric", TenantID: tenantID, Slug: "bites-track-metric",
			Name: "Order Progress Metric", TemplateKey: "tmpl_metric_card", SnippetType: snippets.SnippetTypeMetricCard,
			Props: structToMap(snippets.MetricCardSnippetData{
				Label: atoms.NewText("Delivery Progress", theming.TypographyLabelMedium, theming.ColorTextSecondary),
				Value: atoms.NewText("72% Complete", theming.TypographyHeadlineMedium, theming.ColorTextPrimary),
				Delta: ptrTag(atoms.NewTag("ON TIME", theming.ColorSuccessContainer)),
				Progress: &atoms.ProgressAtom{
					Kind: "linear", Progress: atoms.Float64Ptr(0.72), Label: "1.2 km remaining",
					Color: atoms.ColorRef(theming.ColorBrandPrimary),
				},
				Layout: seedCardLayout(),
			}),
			Tags: []string{"tracking", "metric"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_bites_track_confirmed", TenantID: tenantID, Slug: "bites-track-confirmed",
			Name: "Order Confirmed Step", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:       atoms.NewText("Order Confirmed & Prepared", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Saffron Hearth Kitchen sealed #AB-4821 hot", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon: ptrIcon(atoms.NewIcon("check", 20, theming.ColorSuccess)),
				TrailingTag: ptrTag(atoms.NewTag("DONE", theming.ColorSuccessContainer)),
				Layout:      seedCardLayout(),
			}),
			Tags: []string{"tracking", "step"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_bites_track_rider", TenantID: tenantID, Slug: "bites-track-rider",
			Name: "Rider En Route Step", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:       atoms.NewText("Rider En Route", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Karan M. is 1.2 km away on electric bike", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon: ptrIcon(atoms.NewIcon("bolt", 20, theming.ColorBrandPrimary)),
				TrailingTag: ptrTag(atoms.NewTag("LIVE", theming.ColorBrandContainer)),
				Layout:      seedCardLayout(),
			}),
			Tags: []string{"tracking", "step"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
	}

	for _, sp := range dishes {
		widgets = append(widgets, buildSeedProductWidget(tenantID, now, sp))
	}
	return widgets
}

func seedBitesRules(tenantID string, now time.Time) []domain.TargetingRule {
	dinnerStart, dinnerEnd := 13, 18
	return []domain.TargetingRule{
		{
			ID: "rule_bites_mumbai_or_vip", TenantID: tenantID, Key: "mumbai_or_vip",
			Name: "Mumbai City or VIP Diners", Description: "Matches callers in Mumbai or belonging to the VIP segment.",
			MatchMode: "any",
			Conditions: []domain.RuleCondition{
				{Field: "city", Operator: "eq", Value: "Mumbai"},
				{Field: "segment", Operator: "eq", Value: "vip"},
			},
			Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "rule_bites_dinner_rush", TenantID: tenantID, Key: "dinner_rush_window",
			Name: "Evening Dinner Rush (13:00–18:00 UTC)", Description: "Active during peak dinner ordering hours.",
			MatchMode: "all", Conditions: []domain.RuleCondition{},
			Schedule: &domain.ScheduleWindow{StartHourUTC: &dinnerStart, EndHourUTC: &dinnerEnd},
			Priority: 80, Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedBitesSlottingRules(tenantID string, now time.Time) []domain.SlottingRule {
	return []domain.SlottingRule{
		{
			ID: "slotrule_bites_vip_dinner", TenantID: tenantID, LayoutSlug: "bites-home",
			Name: "Inject VIP & Mumbai Dinner Banner", Action: domain.SlottingActionInsert,
			TargetPosition: 2, WidgetID: "w_bites_vip_dinner_banner", RuleID: "rule_bites_mumbai_or_vip",
			Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "slotrule_bites_dinner_window", TenantID: tenantID, LayoutSlug: "bites-home",
			Name: "Inject Peak Dinner Banner", Action: domain.SlottingActionInsert,
			TargetPosition: 2, WidgetID: "w_bites_vip_dinner_banner", RuleID: "rule_bites_dinner_rush",
			Priority: 80, Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedBitesLayouts(tenantID string, now time.Time) []domain.Layout {
	navTrack := actions.NewNavigateAction("nav_bites_order_track", actions.NavigatePayload{Route: "/screens/bites-order-track"})
	trackIcon := atoms.NewIcon("notifications", 22, theming.ColorBrandPrimary)
	trackIcon.ContentDescription = "Track active order"
	trackIcon.OnClick = &navTrack

	cart := buildSeedFloatingCart(
		"2 items · $17.75", "24 MINS", "bites_place_order", "Place Meal Order",
		"/api/v1/bites/orders", map[string]any{"restaurant_id": "saffron_hearth", "vertical": "abeta-bites"},
	)
	pullHome := actions.NewRefreshPageAction("pull_bites_home", actions.RefreshPagePayload{PageID: "bites-home", ShowLoader: true})

	return []domain.Layout{
		{
			ID: "lay_bites_home", TenantID: tenantID, Slug: "bites-home",
			Title: "Abeta Bites — Food Delivery", Description: "Food delivery discovery home with filters, dining hero, restaurants, and menu grid.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Abeta Bites", "Hot meals · Live tracking in {{city}}", "back_bites_home", trackIcon),
				FloatingCart:   cart,
				AdaptiveLayout: buildSeedCenteredAdaptive(1120),
				PullToRefresh:  &pullHome,
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_bites_search", WidgetID: "w_bites_search_bar", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_bites_chips", WidgetID: "w_bites_filter_chips", Position: 1, Enabled: true},
				{SlotID: "slot_bites_banner", WidgetID: "w_bites_dining_banner", Position: 2, Enabled: true},
				{SlotID: "slot_bites_saffron", WidgetID: "w_bites_rest_saffron", Position: 3, Enabled: true},
				{SlotID: "slot_bites_cedar", WidgetID: "w_bites_rest_cedar", Position: 4, Enabled: true},
				{SlotID: "slot_bites_bamboo", WidgetID: "w_bites_rest_bamboo", Position: 5, Enabled: true},
				{SlotID: "slot_bites_menu", WidgetID: "w_bites_menu_grid", Position: 6, Enabled: true},
			},
			Tags: []string{"vertical", "food", "delivery"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_bites_restaurant", TenantID: tenantID, Slug: "bites-restaurant",
			Title: "Saffron Hearth Kitchen", Description: "Restaurant menu storefront with dish cards and meal cart checkout.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Saffron Hearth Kitchen", "4.8 ★ · 24 mins · North Indian", "back_bites_rest"),
				FloatingCart:   cart,
				AdaptiveLayout: buildSeedCenteredAdaptive(1120),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_bites_rest_hdr", WidgetID: "w_bites_rest_header", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_bites_rest_menu", WidgetID: "w_bites_menu_grid", Position: 1, Enabled: true},
			},
			Tags: []string{"vertical", "restaurant", "menu"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_bites_order_track", TenantID: tenantID, Slug: "bites-order-track",
			Title: "Track Order #AB-4821", Description: "Real-time meal order milestone progress and rider telemetry.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Order #AB-4821", "Arriving in 14 mins · Abeta Bites", "back_bites_track"),
				AdaptiveLayout: buildSeedCenteredAdaptive(840),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_track_banner", WidgetID: "w_bites_track_banner", Position: 0, Enabled: true},
				{SlotID: "slot_track_metric", WidgetID: "w_bites_track_metric", Position: 1, Enabled: true},
				{SlotID: "slot_track_step_1", WidgetID: "w_bites_track_confirmed", Position: 2, Enabled: true},
				{SlotID: "slot_track_step_2", WidgetID: "w_bites_track_rider", Position: 3, Enabled: true},
			},
			Tags: []string{"vertical", "order-tracking"}, CreatedAt: now, UpdatedAt: now,
		},
	}
}
