package transformer_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/pkg/transformer"
)

func TestDefaultRegistry_All21SnippetTypesRoundTrip(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	supported := reg.SupportedTypes()
	if len(supported) != 21 {
		t.Fatalf("SupportedTypes() len = %d, want 21 (%v)", len(supported), supported)
	}

	ctx := domain.RequestContext{
		TenantSlug:  "abeta-fresh",
		Platform:    "android",
		AppVersion:  "2.1.0",
		City:        "Bengaluru",
		Zone:        "koramangala-01",
		Segment:     "vip",
		WindowClass: "expanded",
	}

	widgets := map[string]domain.Widget{
		"w_header": {
			ID:          "w_header",
			SnippetType: snippets.SnippetTypeHeader,
			Active:      true,
			Props: map[string]any{
				"title": map[string]any{
					"text":       "Welcome to {{city}}",
					"typography": map[string]any{"token": string(theming.TypographyHeadlineSmall)},
				},
				"subtitle":     "Zone {{zone}} · {{segment}}",
				"trailing_tag": "EXPRESS",
			},
			Tracking: map[string]any{
				"event_name": "header_impression",
				"properties": map[string]any{"city": "{{city}}"},
			},
			Motion: map[string]any{
				"enter":          atoms.MotionTransitionFadeIn,
				"duration_token": atoms.MotionDurationFast,
			},
		},
		"w_banner": {
			ID:          "w_banner",
			SnippetType: snippets.SnippetTypeBanner,
			Active:      true,
			Props: map[string]any{
				"title": "8-Min Delivery in {{city}}",
				"body":  "App v{{app_version}} on {{platform}}",
				"icon":  "bolt",
				"tag":   "LIVE",
			},
		},
		"w_metric": {
			ID:          "w_metric",
			SnippetType: snippets.SnippetTypeMetricCard,
			Active:      true,
			Props: map[string]any{
				"label": "Orders",
				"value": "1,248",
				"delta": "+14%",
			},
		},
		"w_card": {
			ID:          "w_card",
			SnippetType: snippets.SnippetTypeCard,
			Active:      true,
			Props: map[string]any{
				"title":       "Curated Picks",
				"subtitle":    "For {{segment}} shoppers",
				"description": "Fresh bakery & pantry essentials",
				"badge":       "TOP",
			},
		},
		"w_action_row": {
			ID:          "w_action_row",
			SnippetType: snippets.SnippetTypeActionRow,
			Active:      true,
			Props: map[string]any{
				"title":         "Manage Addresses",
				"subtitle":      "{{zone}}, {{city}}",
				"leading_icon":  "location_on",
				"trailing_icon": "chevron_right",
			},
		},
		"w_form": {
			ID:          "w_form",
			SnippetType: snippets.SnippetTypeForm,
			Active:      true,
			Props: map[string]any{
				"form_id": "feedback_form",
				"title":   "Delivery Feedback",
				"fields": []map[string]any{
					{"field_id": "notes", "label": "Notes", "required": true},
				},
				"submit_button": map[string]any{"label": "Submit"},
			},
		},
		"w_empty": {
			ID:          "w_empty",
			SnippetType: snippets.SnippetTypeEmptyState,
			Active:      true,
			Props: map[string]any{
				"title":   "No Past Orders",
				"message": "Start exploring {{tenant}} in {{city}}",
				"icon":    "shopping_bag",
			},
		},
		"w_divider": {
			ID:          "w_divider",
			SnippetType: snippets.SnippetTypeDivider,
			Active:      true,
			Props: map[string]any{
				"thickness_dp": 1.0,
			},
		},
		"w_spacer": {
			ID:          "w_spacer",
			SnippetType: snippets.SnippetTypeSpacer,
			Active:      true,
			Props: map[string]any{
				"height_dp": 16,
			},
		},
		"w_product_1": {
			ID:          "w_product_1",
			Slug:        "organic_milk",
			SnippetType: snippets.SnippetTypeProductCard,
			Active:      true,
			Props: map[string]any{
				"product_id":       "sku_milk_1l",
				"title":            "Organic Farm Milk",
				"price":            "₹68",
				"unit_price_cents": 6800,
				"badge":            "10% OFF",
				"stepper": map[string]any{
					"field_id": "qty_milk",
					"label":    "ADD",
					"quantity": 1,
				},
			},
		},
		"w_product_2": {
			ID:          "w_product_2",
			Slug:        "sourdough_loaf",
			SnippetType: snippets.SnippetTypeProductCard,
			Active:      true,
			Props: map[string]any{
				"product_id":       "sku_bread",
				"title":            "Artisan Sourdough",
				"price":            "₹145",
				"unit_price_cents": 14500,
			},
		},
		"w_carousel": {
			ID:          "w_carousel",
			SnippetType: snippets.SnippetTypeCarousel,
			Active:      true,
			Props: map[string]any{
				"title":            "Trending in {{city}}",
				"item_width_dp":    180,
				"child_widget_ids": []string{"w_product_1"},
				"items": []snippets.SnippetEnvelope{
					snippets.NewMetricCardSnippet(
						"inline_eta",
						atoms.NewText("ETA", theming.TypographyBodySmall, theming.ColorTextSecondary),
						atoms.NewText("8 mins", theming.TypographyTitleMedium, theming.ColorTextPrimary),
					).Build(),
				},
			},
		},
		"w_grid": {
			ID:          "w_grid",
			SnippetType: snippets.SnippetTypeGrid,
			Active:      true,
			Props: map[string]any{
				"title":   "Quick Staples ({{window_class}})",
				"columns": 2,
				"columns_by_width": map[string]any{
					"compact":  2,
					"medium":   3,
					"expanded": 4,
				},
				"child_widget_ids": []string{"w_product_1", "w_product_2"},
			},
		},
		"w_tabs": {
			ID:          "w_tabs",
			SnippetType: snippets.SnippetTypeTabs,
			Active:      true,
			Props: map[string]any{
				"selected_index": 0,
				"tabs": []map[string]any{
					{
						"id":         "tab_dairy",
						"title":      "Dairy",
						"badge":      "HOT",
						"widget_ids": []string{"w_product_1"},
					},
					{
						"id":         "tab_bakery",
						"title":      "Bakery",
						"widget_ids": []string{"w_product_2"},
					},
				},
			},
		},
		"w_accordion": {
			ID:          "w_accordion",
			SnippetType: snippets.SnippetTypeAccordion,
			Active:      true,
			Props: map[string]any{
				"title":              "Delivery Guarantees",
				"subtitle":           "{{city}} Promise",
				"initially_expanded": true,
				"child_widget_ids":   []string{"w_action_row"},
			},
		},
		"w_swipe_deck": {
			ID:          "w_swipe_deck",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props: map[string]any{
				"deck_id":       "spark_feed",
				"empty_title":   "All caught up in {{city}}",
				"empty_message": "Come back later for more profiles",
				"cards": []map[string]any{
					{
						"card_id":       "card_1",
						"title":         "Meera, 27",
						"subtitle":      "Product Designer · {{city}}",
						"bio":           "Specialty coffee and weekend hikes",
						"distance_text": "2 km away",
						"badge":         "VERIFIED",
						"image":         "https://cdn.abeta.dev/profiles/meera.png",
						"tags":          []string{"Design", "{{segment}}"},
					},
				},
			},
		},
		"w_search_bar": {
			ID:          "w_search_bar",
			SnippetType: snippets.SnippetTypeSearchBar,
			Active:      true,
			Props: map[string]any{
				"placeholder":   "Search groceries in {{city}}...",
				"leading_icon":  "search",
				"trailing_icon": "mic",
				"submit_label":  "Go",
			},
		},
		"w_chip_group": {
			ID:          "w_chip_group",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props: map[string]any{
				"title":        "Filter in {{city}}",
				"multi_select": true,
				"chips": []map[string]any{
					{"id": "all", "label": "All", "selected": true, "badge": "24"},
					{"id": "fast", "label": "Express {{zone}}", "icon": "bolt", "badge": "8"},
				},
			},
		},
		"w_nudge_bar": {
			ID:          "w_nudge_bar",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    "Flash Drop in {{city}}",
				"subtitle": "Free delivery above $15",
				"icon":     "bolt",
				"badge":    "FLASH",
				"progress": 0.7,
				"countdown": map[string]any{
					"countdown_id":      "cd_flash",
					"remaining_seconds": 300,
					"label":             "Ends in",
				},
			},
		},
		"w_marquee_strip": {
			ID:          "w_marquee_strip",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props: map[string]any{
				"items": []map[string]any{
					{"id": "m1", "text": "8-Min Chilled Delivery in {{city}}", "icon": "bolt", "badge": "LIVE"},
					{"id": "m2", "text": "FRESH20 — 20% Off", "icon": "star"},
				},
			},
		},
		"w_order_tracker": {
			ID:          "w_order_tracker",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props: map[string]any{
				"order_code": "#BTS-4091",
				"headline":   "Arriving in {{city}} in 12 mins",
				"eta_badge":  "12 MINS",
				"rider_name": "Karan M.",
				"steps": []map[string]any{
					{"id": "s1", "label": "Confirmed", "icon": "check_circle", "completed": true},
					{"id": "s2", "label": "En Route", "icon": "local_shipping", "active": true},
				},
			},
		},
		"w_composed_block": {
			ID:          "w_composed_block",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"recipe_name": "loyalty_stamp_card",
				"root": map[string]any{
					"kind": snippets.ComposedNodeCard,
					"children": []map[string]any{
						{"kind": snippets.ComposedNodeText, "text": "Loyalty Rewards · {{city}}"},
						{"kind": snippets.ComposedNodeTag, "tag": "4/6 STAMPS"},
					},
				},
			},
		},
	}

	slots := []domain.LayoutSlot{
		{SlotID: "s_header", WidgetID: "w_header", Position: 0, Enabled: true},
		{SlotID: "s_banner", WidgetID: "w_banner", Position: 1, Enabled: true, Overrides: map[string]any{"tag": "VIP-{{segment}}"}},
		{SlotID: "s_metric", WidgetID: "w_metric", Position: 2, Enabled: true},
		{SlotID: "s_card", WidgetID: "w_card", Position: 3, Enabled: true},
		{SlotID: "s_action_row", WidgetID: "w_action_row", Position: 4, Enabled: true},
		{SlotID: "s_form", WidgetID: "w_form", Position: 5, Enabled: true},
		{SlotID: "s_empty", WidgetID: "w_empty", Position: 6, Enabled: true},
		{SlotID: "s_divider", WidgetID: "w_divider", Position: 7, Enabled: true},
		{SlotID: "s_spacer", WidgetID: "w_spacer", Position: 8, Enabled: true},
		{SlotID: "s_product", WidgetID: "w_product_1", Position: 9, Enabled: true},
		{SlotID: "s_carousel", WidgetID: "w_carousel", Position: 10, Enabled: true},
		{SlotID: "s_grid", WidgetID: "w_grid", Position: 11, Enabled: true},
		{SlotID: "s_tabs", WidgetID: "w_tabs", Position: 12, Enabled: true},
		{SlotID: "s_accordion", WidgetID: "w_accordion", Position: 13, Enabled: true},
		{SlotID: "s_swipe_deck", WidgetID: "w_swipe_deck", Position: 14, Enabled: true},
		{SlotID: "s_search_bar", WidgetID: "w_search_bar", Position: 15, Enabled: true},
		{SlotID: "s_chip_group", WidgetID: "w_chip_group", Position: 16, Enabled: true},
		{SlotID: "s_nudge_bar", WidgetID: "w_nudge_bar", Position: 17, Enabled: true},
		{SlotID: "s_marquee_strip", WidgetID: "w_marquee_strip", Position: 18, Enabled: true},
		{SlotID: "s_order_tracker", WidgetID: "w_order_tracker", Position: 19, Enabled: true},
		{SlotID: "s_composed_block", WidgetID: "w_composed_block", Position: 20, Enabled: true},
	}

	pageCfg := domain.LayoutPageConfig{
		Toolbar: &snippets.ToolbarSpec{
			Title:    &atoms.TextAtom{Text: "Abeta Fresh · {{city}}"},
			Subtitle: &atoms.TextAtom{Text: "Delivering to {{zone}}"},
		},
		FloatingCart: &snippets.FloatingCartSpec{
			Label:                      "View Cart ({{city}})",
			CTA:                        atoms.NewButton("btn_cart", "Checkout", "primary", nil),
			FreeDeliveryThresholdCents: atoms.IntPtr(1500),
			NudgePrefix:                "Add for {{city}}",
			UnlockedLabel:              "Free Delivery Unlocked in {{city}}!",
		},
		AdaptiveLayout: &snippets.AdaptivePageSpec{
			LayoutMode:          snippets.LayoutModeSplitPane,
			MaxWidthDp:          atoms.IntPtr(1200),
			SplitPaneEnabled:    true,
			SecondarySnippetIDs: []string{"s_metric"},
		},
		Pagination: &snippets.PaginationSpec{
			NextCursor: "cursor_2",
			HasMore:    true,
		},
		PullToRefresh: &actions.ActionTrigger{
			ID:   "ptr_refresh",
			Type: actions.ActionRefreshPage,
		},
		PageActions: []actions.ActionTrigger{
			actions.NewTrackEventAction("trk_load", actions.TrackEventPayload{
				EventName:  "store_open",
				Properties: map[string]any{"city": "{{city}}"},
			}),
		},
	}

	composer := transformer.NewPageComposer(reg)
	page, err := composer.Compose("store-home", "Abeta Fresh · {{city}}", pageCfg, slots, widgets, ctx)
	if err != nil {
		t.Fatalf("Compose() unexpected error: %v", err)
	}

	// Round-trip the entire PageResponse through JSON to verify 1:1 wire compatibility.
	rawJSON, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("json.Marshal(PageResponse) failed: %v", err)
	}
	var decoded snippets.PageResponse
	if err := json.Unmarshal(rawJSON, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(PageResponse) failed: %v", err)
	}

	if decoded.PageID != "store-home" || decoded.Title != "Abeta Fresh · Bengaluru" {
		t.Fatalf("unexpected page identity: id=%q title=%q", decoded.PageID, decoded.Title)
	}
	if decoded.Toolbar == nil || decoded.Toolbar.Title == nil || decoded.Toolbar.Title.Text != "Abeta Fresh · Bengaluru" {
		t.Fatalf("toolbar title not interpolated: %+v", decoded.Toolbar)
	}
	if decoded.FloatingCart == nil || decoded.FloatingCart.Label != "View Cart (Bengaluru)" {
		t.Fatalf("floating cart label not interpolated: %+v", decoded.FloatingCart)
	}
	if decoded.FloatingCart.FreeDeliveryThresholdCents == nil || *decoded.FloatingCart.FreeDeliveryThresholdCents != 1500 || decoded.FloatingCart.UnlockedLabel != "Free Delivery Unlocked in Bengaluru!" {
		t.Fatalf("floating cart free-delivery threshold nudge not preserved/interpolated: %+v", decoded.FloatingCart)
	}
	if decoded.Tracking == nil || decoded.Tracking.EventName != "page_view" {
		t.Fatalf("expected page_view tracking meta, got %+v", decoded.Tracking)
	}
	if len(decoded.Snippets) != 21 {
		t.Fatalf("expected 21 snippets, got %d", len(decoded.Snippets))
	}

	// Verify every single one of the 21 snippet types decodes cleanly into its KMP wire struct.
	seenTypes := make(map[string]bool, 21)
	for _, env := range decoded.Snippets {
		seenTypes[env.SnippetType] = true
		switch env.SnippetType {
		case snippets.SnippetTypeHeader:
			d, err := snippets.DecodeSnippetData[snippets.HeaderSnippetData](env)
			if err != nil || d.Title.Text != "Welcome to Bengaluru" {
				t.Errorf("header decode failed: err=%v data=%+v", err, d)
			}
			if env.Tracking == nil || env.Tracking.EventName != "header_impression" || env.Tracking.Properties["city"] != "Bengaluru" {
				t.Errorf("header tracking not decoded/interpolated: %+v", env.Tracking)
			}
			if env.Motion == nil || env.Motion.Enter != atoms.MotionTransitionFadeIn {
				t.Errorf("header motion not decoded: %+v", env.Motion)
			}
		case snippets.SnippetTypeBanner:
			d, err := snippets.DecodeSnippetData[snippets.BannerSnippetData](env)
			if err != nil || d.Tag == nil || d.Tag.Label != "VIP-vip" {
				t.Errorf("banner override+interpolation failed: err=%v data=%+v", err, d)
			}
		case snippets.SnippetTypeMetricCard:
			if _, err := snippets.DecodeSnippetData[snippets.MetricCardSnippetData](env); err != nil {
				t.Errorf("metric_card decode failed: %v", err)
			}
		case snippets.SnippetTypeCard:
			if _, err := snippets.DecodeSnippetData[snippets.CardSnippetData](env); err != nil {
				t.Errorf("card_v1 decode failed: %v", err)
			}
		case snippets.SnippetTypeActionRow:
			if _, err := snippets.DecodeSnippetData[snippets.ActionRowSnippetData](env); err != nil {
				t.Errorf("action_row decode failed: %v", err)
			}
		case snippets.SnippetTypeForm:
			d, err := snippets.DecodeSnippetData[snippets.FormSnippetData](env)
			if err != nil || d.FormID != "feedback_form" || len(d.Fields) != 1 {
				t.Errorf("form_v1 decode failed: err=%v data=%+v", err, d)
			}
		case snippets.SnippetTypeEmptyState:
			d, err := snippets.DecodeSnippetData[snippets.EmptyStateSnippetData](env)
			if err != nil || d.Message == nil || !strings.Contains(d.Message.Text, "abeta-fresh") {
				t.Errorf("empty_state decode failed: err=%v data=%+v", err, d)
			}
		case snippets.SnippetTypeDivider:
			d, err := snippets.DecodeSnippetData[snippets.DividerSnippetData](env)
			if err != nil || d.Divider.ThicknessDp == nil || *d.Divider.ThicknessDp != 1.0 {
				t.Errorf("divider decode failed: err=%v data=%+v", err, d)
			}
		case snippets.SnippetTypeSpacer:
			d, err := snippets.DecodeSnippetData[snippets.SpacerSnippetData](env)
			if err != nil || d.Spacer.HeightDp == nil || *d.Spacer.HeightDp != 16 {
				t.Errorf("spacer decode failed: err=%v data=%+v", err, d)
			}
		case snippets.SnippetTypeProductCard:
			d, err := snippets.DecodeSnippetData[snippets.ProductCardSnippetData](env)
			if err != nil || d.ProductID != "sku_milk_1l" || d.UnitPriceCents != 6800 {
				t.Errorf("product_card decode failed: err=%v data=%+v", err, d)
			}
		case snippets.SnippetTypeCarousel:
			d, err := snippets.DecodeSnippetData[snippets.CarouselSnippetData](env)
			if err != nil || len(d.Items) != 2 {
				t.Errorf("carousel decode failed (expected 1 inline + 1 child): err=%v items=%d", err, len(d.Items))
			} else if d.Items[1].SnippetParentID != "s_carousel" {
				t.Errorf("carousel child parent ID = %q, want s_carousel", d.Items[1].SnippetParentID)
			}
		case snippets.SnippetTypeGrid:
			d, err := snippets.DecodeSnippetData[snippets.GridSnippetData](env)
			if err != nil || len(d.Items) != 2 || d.ColumnsByWidth == nil || *d.ColumnsByWidth.Expanded != 4 {
				t.Errorf("grid decode failed: err=%v data=%+v", err, d)
			}
			// Verify auto-defaulted stepper on second product card.
			if len(d.Items) == 2 {
				p2, err := snippets.DecodeSnippetData[snippets.ProductCardSnippetData](d.Items[1])
				if err != nil || p2.Stepper.FieldID != "qty_sku_bread" || p2.Stepper.Label != "ADD" {
					t.Errorf("nested product_card stepper default failed: err=%v stepper=%+v", err, p2.Stepper)
				}
			}
		case snippets.SnippetTypeTabs:
			d, err := snippets.DecodeSnippetData[snippets.TabsSnippetData](env)
			if err != nil || len(d.Tabs) != 2 || len(d.Tabs[0].Snippets) != 1 || len(d.Tabs[1].Snippets) != 1 {
				t.Errorf("tabs decode failed: err=%v data=%+v", err, d)
			}
			if len(d.Tabs) > 0 && (d.Tabs[0].Badge == nil || d.Tabs[0].Badge.Label != "HOT") {
				t.Errorf("tab badge shorthand decode failed: %+v", d.Tabs[0].Badge)
			}
		case snippets.SnippetTypeAccordion:
			d, err := snippets.DecodeSnippetData[snippets.AccordionSnippetData](env)
			if err != nil || !d.InitiallyExpanded || len(d.Snippets) != 1 {
				t.Errorf("accordion decode failed: err=%v data=%+v", err, d)
			}
		case snippets.SnippetTypeSwipeDeck:
			d, err := snippets.DecodeSnippetData[snippets.SwipeDeckSnippetData](env)
			if err != nil || d.DeckID != "spark_feed" || len(d.Cards) != 1 {
				t.Errorf("swipe_deck decode failed: err=%v data=%+v", err, d)
			} else if d.Cards[0].Subtitle == nil || d.Cards[0].Subtitle.Text != "Product Designer · Bengaluru" {
				t.Errorf("swipe_deck card interpolation failed: %+v", d.Cards[0])
			} else if d.EmptyTitle == nil || d.EmptyTitle.Text != "All caught up in Bengaluru" {
				t.Errorf("swipe_deck empty_title promotion/interpolation failed: %+v", d.EmptyTitle)
			}
		case snippets.SnippetTypeSearchBar:
			d, err := snippets.DecodeSnippetData[snippets.SearchBarSnippetData](env)
			if err != nil || d.SearchID != "s_search_bar" || !d.Clearable || d.Placeholder != "Search groceries in Bengaluru..." {
				t.Errorf("search_bar decode failed: err=%v data=%+v", err, d)
			} else if d.LeadingIcon == nil || d.LeadingIcon.Name != "search" || d.TrailingIcon == nil || d.TrailingIcon.Name != "mic" {
				t.Errorf("search_bar icon promotion failed: %+v", d)
			}
		case snippets.SnippetTypeChipGroup:
			d, err := snippets.DecodeSnippetData[snippets.ChipGroupSnippetData](env)
			if err != nil || d.GroupID != "s_chip_group" || !d.MultiSelect || len(d.Chips) != 2 {
				t.Errorf("chip_group decode failed: err=%v data=%+v", err, d)
			} else if d.Title == nil || d.Title.Text != "Filter in Bengaluru" || d.Chips[1].Icon == nil || d.Chips[1].Icon.Name != "bolt" {
				t.Errorf("chip_group promotion/interpolation failed: %+v", d)
			}
		case snippets.SnippetTypeNudgeBar:
			d, err := snippets.DecodeSnippetData[snippets.NudgeBarSnippetData](env)
			if err != nil || d.NudgeID != "s_nudge_bar" || !d.Dismissible || d.Title.Text != "Flash Drop in Bengaluru" {
				t.Errorf("nudge_bar decode failed: err=%v data=%+v", err, d)
			} else if d.Countdown == nil || d.Countdown.RemainingSeconds != 300 || d.Badge == nil || d.Badge.Label != "FLASH" {
				t.Errorf("nudge_bar countdown/badge failed: %+v", d)
			}
		case snippets.SnippetTypeMarqueeStrip:
			d, err := snippets.DecodeSnippetData[snippets.MarqueeStripSnippetData](env)
			if err != nil || d.StripID != "s_marquee_strip" || d.VelocityDpPerSec != 48 || d.Separator != "•" || len(d.Items) != 2 {
				t.Errorf("marquee_strip decode failed: err=%v data=%+v", err, d)
			} else if d.Items[0].Text != "8-Min Chilled Delivery in Bengaluru" || d.Items[0].Badge == nil || d.Items[0].Badge.Label != "LIVE" {
				t.Errorf("marquee_strip item interpolation/promotion failed: %+v", d.Items[0])
			}
		case snippets.SnippetTypeOrderTracker:
			d, err := snippets.DecodeSnippetData[snippets.OrderTrackerSnippetData](env)
			if err != nil || d.TrackerID != "s_order_tracker" || d.OrderCode != "#BTS-4091" || d.Headline.Text != "Arriving in Bengaluru in 12 mins" {
				t.Errorf("order_tracker decode failed: err=%v data=%+v", err, d)
			} else if d.EtaBadge == nil || d.EtaBadge.Label != "12 MINS" || len(d.Steps) != 2 {
				t.Errorf("order_tracker eta/steps failed: %+v", d)
			}
		case snippets.SnippetTypeComposedBlock:
			d, err := snippets.DecodeSnippetData[snippets.ComposedBlockSnippetData](env)
			if err != nil || d.BlockID != "s_composed_block" || d.Root.Kind != snippets.ComposedNodeCard || len(d.Root.Children) != 2 {
				t.Errorf("composed_block decode failed: err=%v data=%+v", err, d)
			} else if d.Root.Children[0].Text == nil || d.Root.Children[0].Text.Text != "Loyalty Rewards · Bengaluru" {
				t.Errorf("composed_block child interpolation failed: %+v", d.Root.Children[0])
			}
		}
	}

	if len(seenTypes) != 21 {
		t.Fatalf("expected all 21 snippet types in page, got %d", len(seenTypes))
	}

	// Ensure caller pageCfg was not mutated by interpolation.
	if pageCfg.Toolbar.Title.Text != "Abeta Fresh · {{city}}" {
		t.Fatalf("caller pageCfg.Toolbar.Title was mutated: %q", pageCfg.Toolbar.Title.Text)
	}
}

func TestMergeAndInterpolateProps_DeepMergeAndImmutability(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"title": map[string]any{
			"text":  "Default {{city}}",
			"color": map[string]any{"token": "text.primary"},
		},
		"tags": []string{"{{segment}}", "{{unknown_token}}"},
	}
	overrides := map[string]any{
		"title": map[string]any{
			"text": "Express in {{city}} ({{zone}}) · {{platform}} v{{app_version}} · {{tenant}} · {{window_class}}",
		},
	}
	ctx := domain.RequestContext{
		City:        "Mumbai",
		Zone:        "bandra-02",
		Segment:     "vip",
		Platform:    "ios",
		AppVersion:  "3.0.1",
		TenantSlug:  "abeta-mart",
		WindowClass: "compact",
		Attributes:  map[string]string{"tier": "gold"},
	}

	merged := transformer.MergeAndInterpolateProps(base, overrides, ctx)
	titleMap := merged["title"].(map[string]any)
	wantText := "Express in Mumbai (bandra-02) · ios v3.0.1 · abeta-mart · compact"
	if titleMap["text"] != wantText {
		t.Fatalf("merged title.text = %q, want %q", titleMap["text"], wantText)
	}
	if colorMap := titleMap["color"].(map[string]any); colorMap["token"] != "text.primary" {
		t.Fatalf("deep merge lost sibling key color: %+v", colorMap)
	}
	tags := merged["tags"].([]string)
	if !slices.Equal(tags, []string{"vip", "{{unknown_token}}"}) {
		t.Fatalf("merged tags = %v, want [vip {{unknown_token}}]", tags)
	}

	// Verify original base map was not mutated.
	if base["title"].(map[string]any)["text"] != "Default {{city}}" {
		t.Fatalf("base props mutated in place: %+v", base)
	}
}

func TestRegistry_CustomOCPTransformerAndCycleGuard(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	reg.Register(transformer.WidgetTransformerFunc{
		Type: "countdown_banner_v2",
		Fn: func(input transformer.TransformInput) (snippets.SnippetEnvelope, error) {
			return snippets.SnippetEnvelope{
				Data: map[string]any{
					"headline":    input.Props["headline"],
					"secondsLeft": input.Props["seconds_left"],
				},
			}, nil
		},
	})

	if len(reg.SupportedTypes()) != 22 {
		t.Fatalf("expected 22 supported types after custom registration, got %d", len(reg.SupportedTypes()))
	}

	env, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "custom_countdown", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_countdown",
			SnippetType: "countdown_banner_v2",
			Active:      true,
			Props: map[string]any{
				"headline":     "Sale ends soon in {{city}}!",
				"seconds_left": 180,
			},
			Tracking: map[string]any{"event_name": "countdown_shown"},
		},
		Context: domain.RequestContext{City: "Pune"},
	})
	if err != nil {
		t.Fatalf("custom Transform() failed: %v", err)
	}
	if env.SnippetID != "custom_countdown" || env.SnippetType != "countdown_banner_v2" {
		t.Fatalf("unexpected custom envelope identity: %+v", env)
	}
	if env.Tracking == nil || env.Tracking.EventName != "countdown_shown" {
		t.Fatalf("expected automatic tracking enrichment on custom transformer, got %+v", env.Tracking)
	}
	dataMap := env.Data.(map[string]any)
	if dataMap["headline"] != "Sale ends soon in Pune!" {
		t.Fatalf("unexpected custom headline: %v", dataMap["headline"])
	}

	// Verify cycle guard returns error instead of infinite recursion.
	cyclicWidgets := map[string]domain.Widget{
		"c1": {ID: "c1", SnippetType: snippets.SnippetTypeGrid, Active: true, Props: map[string]any{"child_widget_ids": []string{"c2"}}},
		"c2": {ID: "c2", SnippetType: snippets.SnippetTypeCarousel, Active: true, Props: map[string]any{"child_widget_ids": []string{"c1"}}},
	}
	_, err = reg.Transform(transformer.TransformInput{
		Slot:    domain.LayoutSlot{SlotID: "root_grid", WidgetID: "c1", Enabled: true},
		Widget:  cyclicWidgets["c1"],
		Widgets: cyclicWidgets,
	})
	if err == nil || !strings.Contains(err.Error(), "cyclic widget reference") {
		t.Fatalf("expected cyclic reference error, got %v", err)
	}
}

func TestPageComposer_ErrorAndSkipBehavior(t *testing.T) {
	t.Parallel()

	composer := transformer.NewPageComposer(nil)

	// Case 1: Configured enabled slots but none resolve -> error.
	_, err := composer.Compose(
		"broken-page",
		"Broken",
		domain.LayoutPageConfig{},
		[]domain.LayoutSlot{
			{SlotID: "s1", WidgetID: "missing_w", Enabled: true},
			{SlotID: "s2", WidgetID: "inactive_w", Enabled: true},
		},
		map[string]domain.Widget{
			"inactive_w": {ID: "inactive_w", SnippetType: snippets.SnippetTypeSpacer, Active: false},
		},
		domain.RequestContext{},
	)
	if err == nil {
		t.Fatal("expected error when enabled slots resolve to zero active widgets")
	}

	// Case 2: Mixed disabled/missing/inactive and one valid active widget -> succeeds with 1 snippet.
	page, err := composer.Compose(
		"partial-page",
		"Partial",
		domain.LayoutPageConfig{},
		[]domain.LayoutSlot{
			{SlotID: "s_disabled", WidgetID: "w_ok", Enabled: false},
			{SlotID: "s_missing", WidgetID: "missing", Enabled: true},
			{SlotID: "s_ok", WidgetID: "w_ok", Enabled: true},
		},
		map[string]domain.Widget{
			"w_ok": {
				ID:          "w_ok",
				SnippetType: snippets.SnippetTypeHeader,
				Active:      true,
				Props:       map[string]any{"title": "Active Section"},
			},
		},
		domain.RequestContext{},
	)
	if err != nil {
		t.Fatalf("expected partial composition to succeed, got %v", err)
	}
	if len(page.Snippets) != 1 || page.Snippets[0].SnippetID != "s_ok" {
		t.Fatalf("unexpected resolved snippets: %+v", page.Snippets)
	}

	// Case 3: Zero enabled slots -> valid empty page.
	emptyPage, err := composer.Compose("empty", "Empty", domain.LayoutPageConfig{}, nil, nil, domain.RequestContext{})
	if err != nil {
		t.Fatalf("expected empty slots to produce valid empty page, got %v", err)
	}
	if emptyPage.Snippets == nil || len(emptyPage.Snippets) != 0 {
		t.Fatalf("expected non-nil empty snippets slice, got %#v", emptyPage.Snippets)
	}

	// Case 4: Unsupported snippet type returns error.
	_, err = composer.Compose(
		"bad-type",
		"Bad",
		domain.LayoutPageConfig{},
		[]domain.LayoutSlot{{SlotID: "s1", WidgetID: "w_bad", Enabled: true}},
		map[string]domain.Widget{"w_bad": {ID: "w_bad", SnippetType: "unknown_v99", Active: true}},
		domain.RequestContext{},
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported snippet type") {
		t.Fatalf("expected unsupported snippet type error, got %v", err)
	}
}

func TestTransformer_DefaultsAndContainerAliases(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{
		UserID:     "u_9",
		Attributes: map[string]string{"promo": "MEGA"},
	}

	// Verify attribute interpolation, accordion `items` alias, tabs object badge + clamped index,
	// form default FormID/Fields, and image shorthand promotion.
	widgets := map[string]domain.Widget{
		"w_child": {
			ID:          "w_child",
			SnippetType: snippets.SnippetTypeCard,
			Active:      true,
			Props: map[string]any{
				"title": "Deal {{attr.promo}} for {{user_id}}",
				"media": "https://cdn.abeta.dev/card.png",
			},
		},
		"w_form_min": {
			ID:          "w_form_min",
			Slug:        "minimal_form",
			SnippetType: snippets.SnippetTypeForm,
			Active:      true,
			Props:       map[string]any{"submit_button": map[string]any{"label": "Save"}},
		},
		"w_tabs_alias": {
			ID:          "w_tabs_alias",
			SnippetType: snippets.SnippetTypeTabs,
			Active:      true,
			Props: map[string]any{
				"selected_index": 99,
				"tabs": []any{
					map[string]any{
						"id":               "t_one",
						"title":            map[string]any{"text": "Tab One"},
						"badge":            map[string]any{"label": "NEW"},
						"child_widget_ids": []any{"w_child"},
					},
				},
			},
		},
		"w_acc_alias": {
			ID:          "w_acc_alias",
			SnippetType: snippets.SnippetTypeAccordion,
			Active:      true,
			Props: map[string]any{
				"title": "FAQ",
				"items": []snippets.SnippetEnvelope{
					{SnippetID: "faq_1", SnippetType: snippets.SnippetTypeDivider, Data: snippets.DividerSnippetData{}},
				},
			},
		},
	}

	formEnv, err := reg.Transform(transformer.TransformInput{Widget: widgets["w_form_min"], Context: ctx})
	if err != nil {
		t.Fatalf("form transform failed: %v", err)
	}
	formData, err := snippets.DecodeSnippetData[snippets.FormSnippetData](formEnv)
	if err != nil || formData.FormID != "minimal_form" || formData.Fields == nil {
		t.Fatalf("unexpected minimal form output: err=%v data=%+v", err, formData)
	}

	tabsEnv, err := reg.Transform(transformer.TransformInput{
		Slot:    domain.LayoutSlot{SlotID: "tabs_root", WidgetID: "w_tabs_alias", Enabled: true},
		Widget:  widgets["w_tabs_alias"],
		Widgets: widgets,
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("tabs transform failed: %v", err)
	}
	tabsData, err := snippets.DecodeSnippetData[snippets.TabsSnippetData](tabsEnv)
	if err != nil || tabsData.SelectedIndex != 0 || len(tabsData.Tabs) != 1 || tabsData.Tabs[0].Title != "Tab One" {
		t.Fatalf("unexpected tabs output: err=%v data=%+v", err, tabsData)
	}
	if tabsData.Tabs[0].Badge == nil || tabsData.Tabs[0].Badge.Label != "NEW" {
		t.Fatalf("unexpected tabs object badge: %+v", tabsData.Tabs[0].Badge)
	}
	cardData, err := snippets.DecodeSnippetData[snippets.CardSnippetData](tabsData.Tabs[0].Snippets[0])
	if err != nil || cardData.Title.Text != "Deal MEGA for u_9" || cardData.Media == nil || cardData.Media.URL != "https://cdn.abeta.dev/card.png" {
		t.Fatalf("unexpected nested card output: err=%v data=%+v", err, cardData)
	}

	accEnv, err := reg.Transform(transformer.TransformInput{
		Slot:    domain.LayoutSlot{SlotID: "acc_root", WidgetID: "w_acc_alias", Enabled: true},
		Widget:  widgets["w_acc_alias"],
		Widgets: widgets,
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("accordion transform failed: %v", err)
	}
	accData, err := snippets.DecodeSnippetData[snippets.AccordionSnippetData](accEnv)
	if err != nil || len(accData.Snippets) != 1 || accData.Snippets[0].SnippetID != "faq_1" {
		t.Fatalf("unexpected accordion alias output: err=%v data=%+v", err, accData)
	}
}
