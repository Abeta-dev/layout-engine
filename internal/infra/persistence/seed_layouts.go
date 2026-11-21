package persistence

import (
	"encoding/json"
	"strings"
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

// collectFrozenWidgets recursively captures every widget referenced by layout slots,
// container child_widget_ids, and tabs[*].widget_ids so snapshots stay self-contained.
func collectFrozenWidgets(slots []domain.LayoutSlot, all map[string]domain.Widget) map[string]domain.Widget {
	frozen := make(map[string]domain.Widget, len(slots))
	var visit func(id string)
	visit = func(id string) {
		if id == "" {
			return
		}
		if _, ok := frozen[id]; ok {
			return
		}
		w, ok := all[id]
		if !ok {
			return
		}
		frozen[id] = w

		for _, cid := range extractSeedIDs(w.Props["child_widget_ids"]) {
			visit(cid)
		}
		if rawTabs, ok := w.Props["tabs"].([]any); ok {
			for _, tabRaw := range rawTabs {
				if tabMap, ok := tabRaw.(map[string]any); ok {
					for _, cid := range extractSeedIDs(tabMap["widget_ids"]) {
						visit(cid)
					}
				}
			}
		}
		if typedTabs, ok := w.Props["tabs"].([]map[string]any); ok {
			for _, tabMap := range typedTabs {
				for _, cid := range extractSeedIDs(tabMap["widget_ids"]) {
					visit(cid)
				}
			}
		}
	}
	for _, sl := range slots {
		visit(sl.WidgetID)
	}
	return frozen
}

func extractSeedIDs(raw any) []string {
	switch list := raw.(type) {
	case []string:
		return list
	case []any:
		out := make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// composeSeedFallbackPage builds a fully hydrated, KMP-v1-valid fallback PageResponse
// for each seeded layout so fallback serving works immediately out of the box.
func composeSeedFallbackPage(layout domain.Layout, widgets map[string]domain.Widget) snippets.PageResponse {
	envs := make([]snippets.SnippetEnvelope, 0, len(layout.Slots))
	for _, sl := range layout.Slots {
		if !sl.Enabled {
			continue
		}
		w, ok := widgets[sl.WidgetID]
		if !ok {
			continue
		}
		envs = append(envs, hydrateSeedEnvelope(w, "", widgets, map[string]bool{}))
	}

	return snippets.PageResponse{
		PageID:         layout.Slug,
		Title:          layout.Title,
		Toolbar:        layout.PageConfig.Toolbar,
		Snippets:       envs,
		BottomBar:      layout.PageConfig.BottomBar,
		FloatingCart:   layout.PageConfig.FloatingCart,
		PageActions:    layout.PageConfig.PageActions,
		Pagination:     layout.PageConfig.Pagination,
		PullToRefresh:  layout.PageConfig.PullToRefresh,
		AdaptiveLayout: layout.PageConfig.AdaptiveLayout,
		Tracking: &atoms.TrackingMeta{
			EventName:  "page_loaded",
			Properties: map[string]any{"page_id": layout.Slug, "fallback": true},
		},
	}
}

func hydrateSeedEnvelope(
	w domain.Widget,
	parentID string,
	widgets map[string]domain.Widget,
	visited map[string]bool,
) snippets.SnippetEnvelope {
	nextVisited := make(map[string]bool, len(visited)+1)
	for k, v := range visited {
		nextVisited[k] = v
	}
	nextVisited[w.ID] = true

	data := cloneAndReplaceCity(w.Props, "Bengaluru")

	switch w.SnippetType {
	case snippets.SnippetTypeGrid, snippets.SnippetTypeCarousel:
		childIDs := extractSeedIDs(data["child_widget_ids"])
		if len(childIDs) > 0 {
			items := make([]snippets.SnippetEnvelope, 0, len(childIDs))
			for _, cid := range childIDs {
				if nextVisited[cid] {
					continue
				}
				if child, ok := widgets[cid]; ok {
					items = append(items, hydrateSeedEnvelope(child, w.ID, widgets, nextVisited))
				}
			}
			data["items"] = items
			delete(data, "child_widget_ids")
		}
	case snippets.SnippetTypeTabs:
		if rawTabs, ok := data["tabs"].([]any); ok {
			hydratedTabs := make([]map[string]any, 0, len(rawTabs))
			for _, tabRaw := range rawTabs {
				tabMap, ok := tabRaw.(map[string]any)
				if !ok {
					continue
				}
				tabCopy := make(map[string]any, len(tabMap))
				for k, v := range tabMap {
					tabCopy[k] = v
				}
				cids := extractSeedIDs(tabCopy["widget_ids"])
				tabSnippets := make([]snippets.SnippetEnvelope, 0, len(cids))
				for _, cid := range cids {
					if nextVisited[cid] {
						continue
					}
					if child, ok := widgets[cid]; ok {
						tabSnippets = append(tabSnippets, hydrateSeedEnvelope(child, w.ID, widgets, nextVisited))
					}
				}
				tabCopy["snippets"] = tabSnippets
				delete(tabCopy, "widget_ids")
				hydratedTabs = append(hydratedTabs, tabCopy)
			}
			data["tabs"] = hydratedTabs
		}
	}

	var tracking *atoms.TrackingMeta
	if len(w.Tracking) > 0 {
		raw, err := json.Marshal(w.Tracking)
		if err == nil {
			var tm atoms.TrackingMeta
			if json.Unmarshal(raw, &tm) == nil && tm.EventName != "" {
				tracking = &tm
			}
		}
	}

	var motion *atoms.MotionSpec
	if len(w.Motion) > 0 {
		raw, err := json.Marshal(w.Motion)
		if err == nil {
			var ms atoms.MotionSpec
			if json.Unmarshal(raw, &ms) == nil {
				motion = &ms
			}
		}
	}

	return snippets.SnippetEnvelope{
		SnippetID:       w.ID,
		SnippetParentID: parentID,
		SnippetType:     w.SnippetType,
		Data:            data,
		Tracking:        tracking,
		Motion:          motion,
	}
}

func cloneAndReplaceCity(props map[string]any, city string) map[string]any {
	raw, err := json.Marshal(props)
	if err != nil {
		return map[string]any{}
	}
	replaced := strings.ReplaceAll(string(raw), "{{city}}", city)
	var out map[string]any
	if err := json.Unmarshal([]byte(replaced), &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}
