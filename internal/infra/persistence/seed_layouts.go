package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

func seedLayouts(tenantID string, now time.Time) []domain.Layout {
	return []domain.Layout{
		buildSeedStoreHomeLayout(tenantID, now),
		buildSeedFreshHomeLayout(tenantID, now),
		buildSeedFlashSaleLayout(tenantID, now),
		buildSeedHomeLayout(tenantID, now),
		buildSeedCatalogLayout(tenantID, now),
		buildSeedFormDemoLayout(tenantID, now),
		buildSeedContainersLayout(tenantID, now),
		buildSeedWatchStatusLayout(tenantID, now),
	}
}

func buildSeedCheckoutSheetAction() actions.ActionTrigger {
	placeOrderAct := actions.NewCallAPIAction("place_express_order", actions.CallAPIPayload{
		Endpoint: "/api/v1/orders/express",
		Method:   "POST",
		Body: map[string]any{
			"store_brand":       "Abeta Fresh Express",
			"delivery_eta_mins": 8,
			"coupon_code":       "FRESH20",
		},
	}).WithSecondaryActions(
		actions.NewDismissOverlayAction("dismiss_express_checkout", actions.DismissOverlayPayload{
			OverlayID: "express_checkout_sheet",
		}),
	)

	cancelAct := actions.NewDismissOverlayAction("cancel_express_checkout", actions.DismissOverlayPayload{
		OverlayID: "express_checkout_sheet",
	})

	return actions.NewOpenDialogAction("open_express_checkout", actions.OpenDialogPayload{
		DialogID:       "express_checkout_sheet",
		Title:          "Place Express Order",
		Message:        "Abeta Fresh Express · 8-Min Chilled Delivery · Free delivery applied.",
		Presentation:   actions.PresentationBottomSheet,
		Dismissible:    atoms.BoolPtr(true),
		PositiveAction: &placeOrderAct,
		NegativeAction: &cancelAct,
	})
}

func buildSeedBackIcon(actionID string) *atoms.IconAtom {
	backAct := actions.NewBackPressAction(actionID, actions.BackPressPayload{Steps: 1})
	icon := atoms.NewIcon("arrow_back", 22, theming.ColorTextPrimary)
	icon.ContentDescription = "Navigate back"
	icon.OnClick = &backAct
	return &icon
}

func buildSeedStoreHomeLayout(tenantID string, now time.Time) domain.Layout {
	checkoutSheet := buildSeedCheckoutSheetAction()

	cartIcon := atoms.NewIcon("shopping_cart", 22, theming.ColorBrandPrimary)
	cartIcon.ContentDescription = "Open express checkout"
	cartIcon.OnClick = &checkoutSheet

	titleAtom := atoms.NewText("Abeta Fresh", theming.TypographyTitleLarge, theming.ColorTextPrimary)
	subAtom := atoms.NewText("8-Min Express Delivery", theming.TypographyBodySmall, theming.ColorTextSecondary)

	etaBadge := atoms.NewTag("8 MINS", theming.ColorSuccessContainer)
	etaBadge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	checkoutCTA := atoms.NewButton("btn_floating_checkout", "Place Express Order", "primary", &checkoutSheet)
	checkoutCTA.Motion = &atoms.MotionSpec{
		Enter:         atoms.MotionTransitionSpringBounce,
		DurationToken: atoms.MotionDurationFast,
	}

	pullRefresh := actions.NewRefreshPageAction("pull_refresh_store_home", actions.RefreshPagePayload{
		PageID:     "store-home",
		ShowLoader: true,
	})

	return domain.Layout{
		ID:               "lay_store_home",
		TenantID:         tenantID,
		Slug:             "store-home",
		Title:            "Abeta Fresh — 8-Min Express",
		Description:      "Flagship quick-commerce storefront with dynamic VIP & morning slotting, responsive product grids, and floating checkout dock.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			Toolbar: &snippets.ToolbarSpec{
				Title:           &titleAtom,
				Subtitle:        &subAtom,
				NavigationIcon:  buildSeedBackIcon("back_from_store_home"),
				TrailingActions: []atoms.IconAtom{cartIcon},
				BgColor:         atoms.ColorRef(theming.ColorSurfacePrimary),
			},
			FloatingCart: &snippets.FloatingCartSpec{
				Label:    "2 items · $8.48",
				EtaBadge: &etaBadge,
				CTA:      checkoutCTA,
				Motion: &atoms.MotionSpec{
					Enter:         atoms.MotionTransitionSlideUp,
					Exit:          atoms.MotionTransitionSlideDown,
					DurationToken: atoms.MotionDurationNormal,
				},
			},
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode:          snippets.LayoutModeSplitPane,
				MaxWidthDp:          atoms.IntPtr(1160),
				SplitPaneEnabled:    true,
				SecondarySnippetIDs: []string{"w_delivery_faq"},
			},
			PullToRefresh: &pullRefresh,
			PageActions: []actions.ActionTrigger{
				actions.NewTrackEventAction("track_store_home_view", actions.TrackEventPayload{
					EventName:  "screen_viewed",
					Properties: map[string]any{"screen_id": "store-home", "brand": "Abeta Fresh Express"},
				}),
			},
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_store_header", WidgetID: "w_store_header", Position: 0, Pinned: true, Enabled: true},
			{SlotID: "slot_hero_banner", WidgetID: "w_hero_banner", Position: 1, Enabled: true},
			{SlotID: "slot_eta_metric", WidgetID: "w_eta_metric", Position: 2, Enabled: true},
			{SlotID: "slot_fresh_grid", WidgetID: "w_fresh_grid", Position: 3, Enabled: true},
			{SlotID: "slot_flash_carousel", WidgetID: "w_flash_carousel", Position: 4, Enabled: true},
			{SlotID: "slot_delivery_faq", WidgetID: "w_delivery_faq", Position: 5, Enabled: true},
		},
		Tags:      []string{"storefront", "express", "flagship"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedFlashSaleLayout(tenantID string, now time.Time) domain.Layout {
	titleAtom := atoms.NewText("Lightning Flash Sale", theming.TypographyTitleLarge, theming.ColorTextPrimary)
	subAtom := atoms.NewText("Limited-time chilled drops", theming.TypographyBodySmall, theming.ColorTextSecondary)
	pullRefresh := actions.NewRefreshPageAction("pull_refresh_flash_sale", actions.RefreshPagePayload{
		PageID:     "flash-sale",
		ShowLoader: true,
	})

	return domain.Layout{
		ID:               "lay_flash_sale",
		TenantID:         tenantID,
		Slug:             "flash-sale",
		Title:            "Lightning Flash Sale",
		Description:      "High-urgency promotional drop screen featuring VIP deals, flash carousel, and express essentials.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			Toolbar: &snippets.ToolbarSpec{
				Title:          &titleAtom,
				Subtitle:       &subAtom,
				NavigationIcon: buildSeedBackIcon("back_from_flash_sale"),
				BgColor:        atoms.ColorRef(theming.ColorSurfacePrimary),
			},
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode: snippets.LayoutModeCentered,
				MaxWidthDp: atoms.IntPtr(1040),
			},
			PullToRefresh: &pullRefresh,
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_flash_vip_banner", WidgetID: "w_vip_flash_banner", Position: 0, Enabled: true},
			{SlotID: "slot_flash_carousel", WidgetID: "w_flash_carousel", Position: 1, Enabled: true},
			{SlotID: "slot_flash_fresh_grid", WidgetID: "w_fresh_grid", Position: 2, Enabled: true},
			{SlotID: "slot_flash_quick_actions", WidgetID: "w_quick_actions", Position: 3, Enabled: true},
		},
		Tags:      []string{"promo", "flash-sale"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedHomeLayout(tenantID string, now time.Time) domain.Layout {
	titleAtom := atoms.NewText("Abeta Dynamic Home", theming.TypographyTitleLarge, theming.ColorTextPrimary)
	subAtom := atoms.NewText("Abeta Fresh Express", theming.TypographyBodySmall, theming.ColorTextSecondary)
	pullRefresh := actions.NewRefreshPageAction("pull_refresh_home", actions.RefreshPagePayload{
		PageID:     "home",
		ShowLoader: true,
	})

	return domain.Layout{
		ID:               "lay_home",
		TenantID:         tenantID,
		Slug:             "home",
		Title:            "Abeta Dynamic Home",
		Description:      "Primary entry home screen dynamically assembled from reusable Layout Engine widgets.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			Toolbar: &snippets.ToolbarSpec{
				Title:    &titleAtom,
				Subtitle: &subAtom,
				BgColor:  atoms.ColorRef(theming.ColorSurfacePrimary),
			},
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode: snippets.LayoutModeCentered,
				MaxWidthDp: atoms.IntPtr(1040),
			},
			PullToRefresh: &pullRefresh,
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_home_header", WidgetID: "w_store_header", Position: 0, Pinned: true, Enabled: true},
			{SlotID: "slot_home_hero", WidgetID: "w_hero_banner", Position: 1, Enabled: true},
			{SlotID: "slot_home_quick_actions", WidgetID: "w_quick_actions", Position: 2, Enabled: true},
			{SlotID: "slot_home_category_tabs", WidgetID: "w_category_tabs", Position: 3, Enabled: true},
			{SlotID: "slot_home_eta_metric", WidgetID: "w_eta_metric", Position: 4, Enabled: true},
		},
		Tags:      []string{"home", "landing"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedCatalogLayout(tenantID string, now time.Time) domain.Layout {
	titleAtom := atoms.NewText("Dynamic Catalog", theming.TypographyTitleLarge, theming.ColorTextPrimary)
	subAtom := atoms.NewText("Browse chilled produce & bakery", theming.TypographyBodySmall, theming.ColorTextSecondary)
	pullRefresh := actions.NewRefreshPageAction("pull_refresh_catalog", actions.RefreshPagePayload{
		PageID:     "catalog",
		ShowLoader: true,
	})

	return domain.Layout{
		ID:               "lay_catalog",
		TenantID:         tenantID,
		Slug:             "catalog",
		Title:            "Dynamic Catalog",
		Description:      "Category-driven catalog combining tabbed aisles and responsive multi-column grids.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			Toolbar: &snippets.ToolbarSpec{
				Title:          &titleAtom,
				Subtitle:       &subAtom,
				NavigationIcon: buildSeedBackIcon("back_from_catalog"),
				BgColor:        atoms.ColorRef(theming.ColorSurfacePrimary),
			},
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode: snippets.LayoutModeCentered,
				MaxWidthDp: atoms.IntPtr(1120),
			},
			PullToRefresh: &pullRefresh,
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_cat_header", WidgetID: "w_store_header", Position: 0, Pinned: true, Enabled: true},
			{SlotID: "slot_cat_tabs", WidgetID: "w_category_tabs", Position: 1, Enabled: true},
			{SlotID: "slot_cat_fresh_grid", WidgetID: "w_fresh_grid", Position: 2, Enabled: true},
		},
		Tags:      []string{"catalog", "aisles"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedFormDemoLayout(tenantID string, now time.Time) domain.Layout {
	titleAtom := atoms.NewText("Dynamic Form Demo", theming.TypographyTitleLarge, theming.ColorTextPrimary)
	return domain.Layout{
		ID:               "lay_form_demo",
		TenantID:         tenantID,
		Slug:             "form-demo",
		Title:            "Dynamic Order Feedback Form",
		Description:      "Server-driven validated form with inputs, switches, checkboxes, and sliders.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			Toolbar: &snippets.ToolbarSpec{
				Title:          &titleAtom,
				NavigationIcon: buildSeedBackIcon("back_from_form"),
				BgColor:        atoms.ColorRef(theming.ColorSurfacePrimary),
			},
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode: snippets.LayoutModeCentered,
				MaxWidthDp: atoms.IntPtr(720),
			},
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_form_hdr", WidgetID: "w_store_header", Position: 0, Enabled: true},
			{SlotID: "slot_form_main", WidgetID: "w_feedback_form", Position: 1, Enabled: true},
		},
		Tags:      []string{"form", "validation"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedContainersLayout(tenantID string, now time.Time) domain.Layout {
	titleAtom := atoms.NewText("Dynamic Containers", theming.TypographyTitleLarge, theming.ColorTextPrimary)
	return domain.Layout{
		ID:               "lay_containers",
		TenantID:         tenantID,
		Slug:             "containers",
		Title:            "Dynamic Containers & Controls",
		Description:      "Composable carousel, responsive grid, tabs, accordion, and interactive controls composed via Layout Engine.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			Toolbar: &snippets.ToolbarSpec{
				Title:          &titleAtom,
				NavigationIcon: buildSeedBackIcon("back_from_containers"),
				BgColor:        atoms.ColorRef(theming.ColorSurfacePrimary),
			},
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode: snippets.LayoutModeCentered,
				MaxWidthDp: atoms.IntPtr(1120),
			},
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_cont_carousel", WidgetID: "w_flash_carousel", Position: 0, Enabled: true},
			{SlotID: "slot_cont_grid", WidgetID: "w_fresh_grid", Position: 1, Enabled: true},
			{SlotID: "slot_cont_tabs", WidgetID: "w_category_tabs", Position: 2, Enabled: true},
			{SlotID: "slot_cont_accordion", WidgetID: "w_delivery_faq", Position: 3, Enabled: true},
			{SlotID: "slot_cont_controls", WidgetID: "w_feedback_form", Position: 4, Enabled: true},
		},
		Tags:      []string{"containers", "controls"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedWatchStatusLayout(tenantID string, now time.Time) domain.Layout {
	return domain.Layout{
		ID:               "lay_watch_status",
		TenantID:         tenantID,
		Slug:             "watch-status",
		Title:            "Watch Express Glance",
		Description:      "Compact single-column order glance optimized for Apple Watch and Wear OS companions.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			AdaptiveLayout: &snippets.AdaptivePageSpec{
				LayoutMode: snippets.LayoutModeSingleColumn,
				MaxWidthDp: atoms.IntPtr(280),
			},
			PageActions: []actions.ActionTrigger{
				actions.NewTrackEventAction("track_watch_glance", actions.TrackEventPayload{
					EventName:  "watch_screen_viewed",
					Properties: map[string]any{"page_id": "watch-status"},
				}),
			},
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_watch_eta", WidgetID: "w_eta_metric", Position: 0, Enabled: true},
			{SlotID: "slot_watch_order_card", WidgetID: "w_watch_status_card", Position: 1, Enabled: true},
		},
		Tags:      []string{"watch", "wearable"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedFreshHomeLayout(tenantID string, now time.Time) domain.Layout {
	pullRefresh := actions.NewRefreshPageAction("pull_refresh_fresh_home", actions.RefreshPagePayload{
		PageID:     "fresh-home",
		ShowLoader: true,
	})

	return domain.Layout{
		ID:               "lay_fresh_home",
		TenantID:         tenantID,
		Slug:             "fresh-home",
		Title:            "Abeta Fresh — 8-Min Grocery",
		Description:      "Quick-commerce grocery vertical home with search bar, category chips, city-personalized hero banner, responsive product grid, and express cart dock.",
		Status:           domain.LayoutStatusPublished,
		PublishedVersion: 1,
		DraftVersion:     1,
		PageConfig: domain.LayoutPageConfig{
			Toolbar: buildSeedVerticalToolbar("Abeta Fresh", "8-Min Express Hub · {{city}}", "back_from_fresh_home"),
			FloatingCart: buildSeedFloatingCart(
				"2 items · $8.48",
				"8 MINS",
				"fresh_express_checkout",
				"Express Checkout",
				"/api/v1/fresh/orders/express",
				map[string]any{"vertical": "abeta-fresh", "delivery_eta_mins": 8},
			),
			AdaptiveLayout: buildSeedCenteredAdaptive(1120),
			PullToRefresh:  &pullRefresh,
			PageActions: []actions.ActionTrigger{
				actions.NewTrackEventAction("track_fresh_home_view", actions.TrackEventPayload{
					EventName:  "screen_viewed",
					Properties: map[string]any{"screen_id": "fresh-home", "vertical": "abeta-fresh"},
				}),
			},
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_fresh_search", WidgetID: "w_fresh_search_bar", Position: 0, Pinned: true, Enabled: true},
			{SlotID: "slot_fresh_chips", WidgetID: "w_fresh_category_chips", Position: 1, Enabled: true},
			{SlotID: "slot_fresh_hero", WidgetID: "w_hero_banner", Position: 2, Enabled: true},
			{SlotID: "slot_fresh_grid", WidgetID: "w_fresh_grid", Position: 3, Enabled: true},
			{SlotID: "slot_fresh_carousel", WidgetID: "w_flash_carousel", Position: 4, Enabled: true},
		},
		Tags:      []string{"vertical", "grocery", "express"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
