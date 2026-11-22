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
