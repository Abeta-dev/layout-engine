package persistence

import (
	"fmt"
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// seedTemplates returns all 17 canonical KMP v1 SnippetType templates for a tenant.
// idPrefix keeps template primary keys globally unique across tenants while keeping
// template_key canonical (e.g. "tmpl_search_bar") inside every workspace.
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

// ── Shared Vertical Widget & Chrome Builders ──────────────────────────────────

func seedSearchBarLayout() *atoms.LayoutConfig {
	return &atoms.LayoutConfig{
		Padding:      atoms.SymmetricSpacing(theming.SpacingMD.Dp(), 6),
		Margin:       atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingXS.Dp()),
		BgColor:      atoms.ColorRef(theming.ColorSurfacePrimary),
		CornerRadius: atoms.UniformRadius(theming.RadiusMD.Dp()),
		Border: &atoms.BorderSpec{
			WidthDp: 1,
			Color:   *atoms.ColorRef(theming.ColorBorderDefault),
		},
		FillMaxWidth: atoms.BoolPtr(true),
	}
}

func seedChipRowLayout() *atoms.LayoutConfig {
	return &atoms.LayoutConfig{
		Padding:      atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingXS.Dp()),
		FillMaxWidth: atoms.BoolPtr(true),
	}
}

func seedHorizontalPaddingLayout() *atoms.LayoutConfig {
	return &atoms.LayoutConfig{
		Padding:      atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingSM.Dp()),
		FillMaxWidth: atoms.BoolPtr(true),
	}
}

func buildSeedSearchBarWidget(
	tenantID, widgetID, slug, name, searchID, placeholder, routeBase string,
	now time.Time,
) domain.Widget {
	submitAct := actions.NewNavigateAction("submit_"+searchID, actions.NavigatePayload{
		Route: fmt.Sprintf("/screens/%s?q={{query}}", routeBase),
	})
	clearAct := actions.NewNavigateAction("clear_"+searchID, actions.NavigatePayload{
		Route: "/screens/" + routeBase,
	})
	data := snippets.SearchBarSnippetData{
		SearchID:    searchID,
		Placeholder: placeholder,
		LeadingIcon: ptrIcon(atoms.NewIcon("search", 18, theming.ColorTextTertiary)),
		SubmitLabel: "Search",
		Clearable:   true,
		OnSubmit:    &submitAct,
		OnClear:     &clearAct,
		Layout:      seedSearchBarLayout(),
	}
	return domain.Widget{
		ID:          widgetID,
		TenantID:    tenantID,
		Slug:        slug,
		Name:        name,
		TemplateKey: "tmpl_search_bar",
		SnippetType: snippets.SnippetTypeSearchBar,
		Props:       structToMap(data),
		Tracking: map[string]any{
			"event_name": "search_bar_viewed",
			"properties": map[string]any{"search_id": searchID, "screen": routeBase},
		},
		Tags:      []string{"search_bar", "discovery"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type seedChipSpec struct {
	id       string
	label    string
	selected bool
}

func buildSeedChipGroupWidget(
	tenantID, widgetID, slug, name, groupID, routeBase string,
	chips []seedChipSpec,
	now time.Time,
) domain.Widget {
	items := make([]snippets.FilterChipItem, 0, len(chips))
	for _, c := range chips {
		clickAct := actions.NewNavigateAction("chip_"+groupID+"_"+c.id, actions.NavigatePayload{
			Route: fmt.Sprintf("/screens/%s?filter=%s", routeBase, c.id),
		})
		items = append(items, snippets.FilterChipItem{
			ID:       c.id,
			Label:    c.label,
			Selected: c.selected,
			OnClick:  &clickAct,
		})
	}
	data := snippets.ChipGroupSnippetData{
		GroupID:     groupID,
		MultiSelect: false,
		Chips:       items,
		Layout:      seedChipRowLayout(),
	}
	return domain.Widget{
		ID:          widgetID,
		TenantID:    tenantID,
		Slug:        slug,
		Name:        name,
		TemplateKey: "tmpl_chip_group",
		SnippetType: snippets.SnippetTypeChipGroup,
		Props:       structToMap(data),
		Tracking: map[string]any{
			"event_name": "chip_group_viewed",
			"properties": map[string]any{"group_id": groupID, "screen": routeBase},
		},
		Tags:      []string{"chip_group", "filter"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type seedListingCardSpec struct {
	widgetID   string
	slug       string
	name       string
	title      string
	subtitle   string
	details    string
	badge      string
	badgeColor theming.ColorToken
	ctaLabel   string
	route      string
	tags       []string
}

func buildSeedListingCardWidget(tenantID string, now time.Time, sp seedListingCardSpec) domain.Widget {
	navAct := actions.NewNavigateAction("nav_"+sp.widgetID, actions.NavigatePayload{
		Route: "/screens/" + sp.route,
	})
	cta := atoms.NewButton("btn_"+sp.widgetID, sp.ctaLabel, "primary", &navAct)
	cta.Size = "sm"
	data := snippets.CardSnippetData{
		Title:         atoms.NewText(sp.title, theming.TypographyTitleMedium, theming.ColorTextPrimary),
		Subtitle:      ptrText(atoms.NewText(sp.subtitle, theming.TypographyBodySmall, theming.ColorTextBrand)),
		Description:   ptrText(atoms.NewText(sp.details, theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Badge:         ptrTag(atoms.NewTag(sp.badge, sp.badgeColor)),
		PrimaryButton: &cta,
		OnClick:       &navAct,
		Layout:        seedCardLayout(),
	}
	return domain.Widget{
		ID:          sp.widgetID,
		TenantID:    tenantID,
		Slug:        sp.slug,
		Name:        sp.name,
		TemplateKey: "tmpl_card_v1",
		SnippetType: snippets.SnippetTypeCard,
		Props:       structToMap(data),
		Tracking: map[string]any{
			"event_name": "listing_card_viewed",
			"properties": map[string]any{"widget_id": sp.widgetID, "target_route": sp.route},
		},
		Tags:      sp.tags,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildSeedFloatingCart(label, eta, ctaID, ctaLabel, endpoint string, body map[string]any) *snippets.FloatingCartSpec {
	orderAct := actions.NewCallAPIAction(ctaID, actions.CallAPIPayload{
		Endpoint: endpoint,
		Method:   "POST",
		Body:     body,
	})
	etaBadge := atoms.NewTag(eta, theming.ColorSuccessContainer)
	etaBadge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}
	cta := atoms.NewButton("btn_"+ctaID, ctaLabel, "primary", &orderAct)
	cta.Motion = &atoms.MotionSpec{Enter: atoms.MotionTransitionSpringBounce, DurationToken: atoms.MotionDurationFast}
	return &snippets.FloatingCartSpec{
		Label:    label,
		EtaBadge: &etaBadge,
		CTA:      cta,
		Motion: &atoms.MotionSpec{
			Enter:         atoms.MotionTransitionSlideUp,
			Exit:          atoms.MotionTransitionSlideDown,
			DurationToken: atoms.MotionDurationNormal,
		},
	}
}

func buildSeedVerticalToolbar(title, subtitle, backActionID string, trailing ...atoms.IconAtom) *snippets.ToolbarSpec {
	t := atoms.NewText(title, theming.TypographyTitleLarge, theming.ColorTextPrimary)
	s := atoms.NewText(subtitle, theming.TypographyBodySmall, theming.ColorTextSecondary)
	return &snippets.ToolbarSpec{
		Title:           &t,
		Subtitle:        &s,
		NavigationIcon:  buildSeedBackIcon(backActionID),
		TrailingActions: trailing,
		BgColor:         atoms.ColorRef(theming.ColorSurfacePrimary),
	}
}

func buildSeedCenteredAdaptive(maxWidth int) *snippets.AdaptivePageSpec {
	return &snippets.AdaptivePageSpec{
		LayoutMode: snippets.LayoutModeCentered,
		MaxWidthDp: atoms.IntPtr(maxWidth),
	}
}
