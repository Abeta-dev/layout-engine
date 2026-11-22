package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// seedTemplates returns all 21 canonical KMP v1 SnippetType templates for a tenant.
// idPrefix keeps template primary keys globally unique across tenants while keeping
// template_key canonical (e.g. "tmpl_composed_block") inside every workspace.
func seedTemplates(tenantID, idPrefix string, now time.Time) []domain.WidgetTemplate {
	sampleToast := actions.NewShowToastAction("tmpl_toast", actions.ShowToastPayload{
		Message:    "Sample action triggered",
		DurationMs: 1800,
		Variant:    "info",
	})
	sampleBtn := atoms.NewButton("btn_sample", "Action", "primary", &sampleToast)
	sampleStepper := atoms.NewStepper("qty_sample", "ADD", 0, &sampleToast, nil)
	sampleStepper.MinQuantity = 0
	sampleStepper.MaxQuantity = 10

	sampleCountdown := atoms.NewCountdown("sample_countdown", 300, "Ends in")
	sampleCountdown.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopHeartbeat}

	sampleNudgeBadge := atoms.NewTag("FLASH", theming.ColorErrorContainer)
	sampleNudgeBadge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	sampleEtaBadge := atoms.NewTag("12 MINS", theming.ColorSuccessContainer)
	sampleEtaBadge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	sampleComposedEnv := snippets.MoldLoyaltyStampRecipe(
		"sample_composed",
		"sample_loyalty",
		"Loyalty Punch Card",
		3,
		5,
		"2 more orders to unlock your reward",
		sampleToast,
	)

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
			desc:        "Expandable section wrapping FAQ rows, specifications, or preference forms.",
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
		{
			key:         "tmpl_search_bar",
			name:        "Interactive Search Bar",
			snippetType: snippets.SnippetTypeSearchBar,
			desc:        "Query input bar with leading search icon, clear control, and submit navigation or API trigger.",
			props: snippets.SearchBarSnippetData{
				SearchID:    "sample_search",
				Placeholder: "Search items in {{city}}...",
				LeadingIcon: ptrIcon(atoms.NewIcon("search", 18, theming.ColorTextTertiary)),
				SubmitLabel: "Search",
				Clearable:   true,
				OnSubmit:    &sampleToast,
			},
			tags: []string{"discovery", "search"},
		},
		{
			key:         "tmpl_chip_group",
			name:        "Filter Chip Group",
			snippetType: snippets.SnippetTypeChipGroup,
			desc:        "Horizontal single- or multi-select category filter chips with badge and click actions.",
			props: snippets.ChipGroupSnippetData{
				GroupID:     "sample_filters",
				MultiSelect: false,
				Chips: []snippets.FilterChipItem{
					{ID: "all", Label: "All", Selected: true, OnClick: &sampleToast},
					{ID: "featured", Label: "Featured", Selected: false, OnClick: &sampleToast},
				},
			},
			tags: []string{"discovery", "filter"},
		},
		{
			key:         "tmpl_swipe_deck",
			name:        "Interactive Swipe Card Deck",
			snippetType: snippets.SnippetTypeSwipeDeck,
			desc:        "Gesture-driven swipe discovery deck supporting Like, Pass, Super-Like, and Undo actions.",
			props: snippets.SwipeDeckSnippetData{
				DeckID:         "sample_deck",
				LikeLabel:      "LIKE",
				NopeLabel:      "PASS",
				SuperLikeLabel: "SPARK",
				Cards: []snippets.SwipeCardItem{
					{
						CardID:       "sample_profile_1",
						Title:        atoms.NewText("Curated Match", theming.TypographyTitleLarge, theming.ColorTextPrimary),
						Subtitle:     ptrText(atoms.NewText("Product Designer", theming.TypographyLabelLarge, theming.ColorTextBrand)),
						Bio:          ptrText(atoms.NewText("Coffee, design systems, and weekend hikes.", theming.TypographyBodyMedium, theming.ColorTextSecondary)),
						DistanceText: ptrText(atoms.NewText("2.5 km", theming.TypographyLabelMedium, theming.ColorTextSecondary)),
						Badge:        ptrTag(atoms.NewTag("VERIFIED", theming.ColorSuccessContainer)),
						OnLike:       &sampleToast,
						OnNope:       &sampleToast,
						OnSuperLike:  &sampleToast,
					},
				},
				OnUndo: &sampleToast,
			},
			tags: []string{"discovery", "swipe"},
		},
		{
			key:         "tmpl_nudge_bar",
			name:        "Urgency & Progress Nudge Bar",
			snippetType: snippets.SnippetTypeNudgeBar,
			desc:        "Dismissible urgency, threshold, or flash-drop nudge strip with animated badge, live countdown, and progress bar.",
			props: snippets.NudgeBarSnippetData{
				NudgeID:     "sample_nudge",
				Title:       atoms.NewText("Flash Drop Live in {{city}}", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Add $4.50 more to unlock complimentary express delivery.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:        ptrIcon(atoms.NewIcon("bolt", 20, theming.ColorWarning)),
				Badge:       &sampleNudgeBadge,
				Countdown:   &sampleCountdown,
				Progress:    atoms.Float64Ptr(0.68),
				CTA:         &sampleBtn,
				Dismissible: true,
				PointerEdge: snippets.PointerEdgeNone,
			},
			tags: []string{"nudge", "urgency", "promo"},
		},
		{
			key:         "tmpl_marquee_strip",
			name:        "Scrolling Offer Marquee Strip",
			snippetType: snippets.SnippetTypeMarqueeStrip,
			desc:        "Continuously scrolling horizontal announcement ticker with icons, badges, and tap actions.",
			props: snippets.MarqueeStripSnippetData{
				StripID:          "sample_marquee",
				VelocityDpPerSec: 48,
				Separator:        "•",
				Items: []snippets.MarqueeItem{
					{ID: "m1", Text: "8-Min Express in {{city}}", Icon: ptrIcon(atoms.NewIcon("bolt", 16, theming.ColorBrandPrimary)), Badge: ptrTag(atoms.NewTag("LIVE", theming.ColorSuccessContainer))},
					{ID: "m2", Text: "Free Delivery on orders above $15", Icon: ptrIcon(atoms.NewIcon("local_shipping", 16, theming.ColorSuccess))},
				},
			},
			tags: []string{"marquee", "ticker", "promo"},
		},
		{
			key:         "tmpl_order_tracker",
			name:        "Live Order Milestone Tracker",
			snippetType: snippets.SnippetTypeOrderTracker,
			desc:        "Animated 4-stage live order tracker with countdown ETA, pulsing active step, rider card, and support CTAs.",
			props: snippets.OrderTrackerSnippetData{
				TrackerID:    "sample_tracker",
				OrderCode:    "#AB-1024",
				Headline:     atoms.NewText("Rider arriving soon in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				EtaCountdown: &sampleCountdown,
				EtaBadge:     &sampleEtaBadge,
				RiderName:    "Karan M.",
				RiderCaption: "4.9 ★ · Electric bike · Thermal bag",
				Steps: []snippets.OrderTrackerStep{
					{ID: "confirmed", Label: "Order Confirmed", Caption: "Accepted by hub", Icon: ptrIcon(atoms.NewIcon("check_circle", 18, theming.ColorSuccess)), Completed: true},
					{ID: "packed", Label: "Packed Fresh", Caption: "Tamper-sealed", Icon: ptrIcon(atoms.NewIcon("inventory_2", 18, theming.ColorSuccess)), Completed: true},
					{ID: "en_route", Label: "Rider En Route", Caption: "1.2 km away", Icon: ptrIcon(atoms.NewIcon("local_shipping", 18, theming.ColorBrandPrimary)), Active: true},
					{ID: "delivered", Label: "Delivered", Caption: "Doorstep drop-off", Icon: ptrIcon(atoms.NewIcon("home", 18, theming.ColorTextTertiary))},
				},
				PrimaryAction: &sampleBtn,
			},
			tags: []string{"tracking", "order", "live"},
		},
		{
			key:         "tmpl_composed_block",
			name:        "Atomic Composed Block (Molding Engine)",
			snippetType: snippets.SnippetTypeComposedBlock,
			desc:        "Server-molded recursive atomic tree (depth <= 6) combining layout nodes and wire atoms without client app releases.",
			props:       sampleComposedEnv.Data,
			tags:        []string{"composed", "atomic", "molding"},
		},
	}

	out := make([]domain.WidgetTemplate, 0, len(specs))
	for _, sp := range specs {
		id := sp.key
		if idPrefix != "" {
			id = idPrefix + "_" + sp.key
		}
		out = append(out, domain.WidgetTemplate{
			ID:           id,
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
