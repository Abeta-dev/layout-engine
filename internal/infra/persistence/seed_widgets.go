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

type seedProductSpec struct {
	widgetID       string
	slug           string
	productID      string
	title          string
	subtitle       string
	price          string
	originalPrice  string
	unitPriceCents int
	badge          string
	badgeColor     theming.ColorToken
	rating         string
	initialQty     int
	delayMs        int
}

func seedWidgets(tenantID string, now time.Time) []domain.Widget {
	widgets := make([]domain.Widget, 0, 20)
	widgets = append(widgets,
		seedHeaderWidget(tenantID, now),
		seedHeroBannerWidget(tenantID, now),
		seedVIPBannerWidget(tenantID, now),
		seedMorningBannerWidget(tenantID, now),
		seedETAMetricWidget(tenantID, now),
	)

	products := []seedProductSpec{
		{
			widgetID:       "w_prod_avocado",
			slug:           "prod-avocado",
			productID:      "avocados",
			title:          "Hass Avocados",
			subtitle:       "2 pcs · Ready to eat",
			price:          "$3.49",
			originalPrice:  "$4.49",
			unitPriceCents: 349,
			badge:          "BESTSELLER",
			badgeColor:     theming.ColorSuccessContainer,
			rating:         "4.9 ★",
			initialQty:     1,
			delayMs:        0,
		},
		{
			widgetID:       "w_prod_milk",
			slug:           "prod-milk",
			productID:      "whole_milk",
			title:          "A2 Farm Whole Milk",
			subtitle:       "1 L · Chilled glass bottle",
			price:          "$3.29",
			originalPrice:  "$3.79",
			unitPriceCents: 329,
			badge:          "CHILLED",
			badgeColor:     theming.ColorInfoContainer,
			rating:         "4.9 ★",
			initialQty:     0,
			delayMs:        60,
		},
		{
			widgetID:       "w_prod_sourdough",
			slug:           "prod-sourdough",
			productID:      "sourdough",
			title:          "Artisan Sourdough",
			subtitle:       "450 g · Stone-baked today",
			price:          "$4.99",
			originalPrice:  "$5.79",
			unitPriceCents: 499,
			badge:          "BAKED TODAY",
			badgeColor:     theming.ColorWarningContainer,
			rating:         "4.9 ★",
			initialQty:     1,
			delayMs:        120,
		},
		{
			widgetID:       "w_prod_espresso",
			slug:           "prod-espresso",
			productID:      "cold_brew",
			title:          "Oat Milk Espresso Brew",
			subtitle:       "250 ml · Single origin",
			price:          "$3.79",
			originalPrice:  "$4.29",
			unitPriceCents: 379,
			badge:          "BARISTA",
			badgeColor:     theming.ColorBrandContainer,
			rating:         "4.8 ★",
			initialQty:     0,
			delayMs:        180,
		},
		{
			widgetID:       "w_prod_mango",
			slug:           "prod-mango",
			productID:      "alphonso_mango",
			title:          "Alphonso Mango Box",
			subtitle:       "1 kg · Naturally ripened",
			price:          "$6.49",
			originalPrice:  "$7.99",
			unitPriceCents: 649,
			badge:          "FLASH DROP",
			badgeColor:     theming.ColorWarningContainer,
			rating:         "5.0 ★",
			initialQty:     0,
			delayMs:        0,
		},
		{
			widgetID:       "w_prod_berries",
			slug:           "prod-berries",
			productID:      "organic_berries",
			title:          "Organic Berry Medley",
			subtitle:       "250 g · Sweet morning harvest",
			price:          "$4.99",
			originalPrice:  "$5.99",
			unitPriceCents: 499,
			badge:          "ORGANIC",
			badgeColor:     theming.ColorBrandContainer,
			rating:         "4.8 ★",
			initialQty:     0,
			delayMs:        60,
		},
	}
	for _, spec := range products {
		widgets = append(widgets, buildSeedProductWidget(tenantID, now, spec))
	}

	widgets = append(widgets,
		seedFreshGridWidget(tenantID, now),
		seedBakeryGridWidget(tenantID, now),
		seedFlashCarouselWidget(tenantID, now),
		seedQuickActionsWidget(tenantID, now),
		seedCategoryTabsWidget(tenantID, now),
		seedDeliveryFAQWidget(tenantID, now),
		seedFeedbackFormWidget(tenantID, now),
		seedWatchStatusCardWidget(tenantID, now),
	)

	return widgets
}

func seedCardLayout() *atoms.LayoutConfig {
	return &atoms.LayoutConfig{
		Padding:      atoms.UniformSpacing(theming.SpacingLG.Dp()),
		Margin:       atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingXS.Dp()),
		BgColor:      atoms.ColorRef(theming.ColorSurfacePrimary),
		CornerRadius: atoms.UniformRadius(theming.RadiusMD.Dp()),
		Border: &atoms.BorderSpec{
			WidthDp: 1,
			Color:   *atoms.ColorRef(theming.ColorBorderSubtle),
		},
		FillMaxWidth: atoms.BoolPtr(true),
	}
}

func seedCompactCardLayout() *atoms.LayoutConfig {
	return &atoms.LayoutConfig{
		Padding:      atoms.UniformSpacing(theming.SpacingMD.Dp()),
		BgColor:      atoms.ColorRef(theming.ColorSurfacePrimary),
		CornerRadius: atoms.UniformRadius(theming.RadiusMD.Dp()),
		Border: &atoms.BorderSpec{
			WidthDp: 1,
			Color:   *atoms.ColorRef(theming.ColorBorderSubtle),
		},
		FillMaxWidth: atoms.BoolPtr(true),
	}
}

func seedHeaderWidget(tenantID string, now time.Time) domain.Widget {
	navFlash := actions.NewNavigateAction("nav_flash_sale_hdr", actions.NavigatePayload{
		Route: "/screens/flash-sale",
	})
	flashBtn := atoms.NewButton("btn_hdr_flash", "Flash Sale", "outline", &navFlash)
	flashBtn.Size = "sm"

	liveTag := atoms.NewTag("LIVE", theming.ColorSuccessContainer)
	liveTag.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	data := snippets.HeaderSnippetData{
		Title:          atoms.NewText("8-Min Express • {{city}}", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
		Subtitle:       ptrText(atoms.NewText("Curated live by Abeta Layout Engine", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		LeadingIcon:    ptrIcon(atoms.NewIcon("bolt", 24, theming.ColorBrandPrimary)),
		TrailingTag:    &liveTag,
		TrailingAction: &flashBtn,
		Layout: &atoms.LayoutConfig{
			Padding:      atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingMD.Dp()),
			FillMaxWidth: atoms.BoolPtr(true),
		},
	}

	return domain.Widget{
		ID:          "w_store_header",
		TenantID:    tenantID,
		Slug:        "store-header",
		Name:        "Store Express Header",
		TemplateKey: "tmpl_header_v1",
		SnippetType: snippets.SnippetTypeHeader,
		Props:       structToMap(data),
		Tracking:    map[string]any{"event_name": "header_impression", "properties": map[string]any{"widget_id": "w_store_header"}},
		Tags:        []string{"header", "store"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedHeroBannerWidget(tenantID string, now time.Time) domain.Widget {
	navFlash := actions.NewNavigateAction("nav_hero_flash_sale", actions.NavigatePayload{
		Route: "/screens/flash-sale",
	})
	cta := atoms.NewButton("btn_hero_flash_sale", "Shop Flash Sale", "primary", &navFlash)
	cta.Size = "sm"

	etaTag := atoms.NewTag("8 MINS", theming.ColorSuccessContainer)
	etaTag.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	bannerLayout := seedCardLayout()
	bannerLayout.BgColor = atoms.ColorRef(theming.ColorBrandContainer)

	data := snippets.BannerSnippetData{
		Title:  atoms.NewText("Fresh Drop in {{city}}!", theming.TypographyTitleMedium, theming.ColorTextPrimary),
		Body:   ptrText(atoms.NewText("Chilled produce, stone-baked sourdough & espresso brews dispatched live from Abeta Fresh Hub.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Icon:   ptrIcon(atoms.NewIcon("bolt", 24, theming.ColorBrandPrimary)),
		Tag:    &etaTag,
		CTA:    &cta,
		Layout: bannerLayout,
	}

	return domain.Widget{
		ID:          "w_hero_banner",
		TenantID:    tenantID,
		Slug:        "hero-banner",
		Name:        "City Fresh Drop Hero Banner",
		TemplateKey: "tmpl_banner_v1",
		SnippetType: snippets.SnippetTypeBanner,
		Props:       structToMap(data),
		Motion:      structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionSlideUp, DurationToken: atoms.MotionDurationNormal}),
		Tracking:    map[string]any{"event_name": "hero_banner_viewed", "properties": map[string]any{"widget_id": "w_hero_banner"}},
		Tags:        []string{"banner", "hero", "promo"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedVIPBannerWidget(tenantID string, now time.Time) domain.Widget {
	applyVIP := actions.NewBatchAction("apply_vip_discount", actions.BatchPayload{
		Actions: []actions.ActionTrigger{
			actions.NewSaveStateAction("save_vip_coupon", actions.SaveStatePayload{
				Key:   "active_coupon",
				Value: "VIP20",
				Scope: "session",
			}),
			actions.NewTriggerFeedbackAction("vip_haptic", actions.TriggerFeedbackPayload{
				Effect: actions.FeedbackHapticTick,
			}),
			actions.NewShowToastAction("toast_vip_applied", actions.ShowToastPayload{
				Message:    "VIP20 unlocked — extra 20% off your express basket!",
				DurationMs: 2400,
				Variant:    "success",
			}),
		},
		StopOnError: true,
	})
	cta := atoms.NewButton("btn_vip_claim", "Claim VIP20", "primary", &applyVIP)
	cta.Size = "sm"

	vipLayout := seedCardLayout()
	vipLayout.BgColor = atoms.ColorRef(theming.ColorWarningContainer)

	data := snippets.BannerSnippetData{
		Title:  atoms.NewText("VIP Exclusive • Extra 20% Off in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
		Body:   ptrText(atoms.NewText("Priority dark-store packing + free express delivery for Abeta Fresh VIP members.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Icon:   ptrIcon(atoms.NewIcon("sparkles", 24, theming.ColorWarning)),
		Tag:    ptrTag(atoms.NewTag("VIP ONLY", theming.ColorSurfacePrimary)),
		CTA:    &cta,
		Layout: vipLayout,
	}

	return domain.Widget{
		ID:          "w_vip_flash_banner",
		TenantID:    tenantID,
		Slug:        "vip-flash-banner",
		Name:        "VIP Exclusive Flash Banner",
		TemplateKey: "tmpl_banner_v1",
		SnippetType: snippets.SnippetTypeBanner,
		Props:       structToMap(data),
		Motion:      structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionSpringBounce, DurationToken: atoms.MotionDurationFast}),
		Tracking:    map[string]any{"event_name": "vip_banner_viewed", "properties": map[string]any{"segment": "vip"}},
		Tags:        []string{"banner", "vip", "slotting"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedMorningBannerWidget(tenantID string, now time.Time) domain.Widget {
	navCatalog := actions.NewNavigateAction("nav_morning_catalog", actions.NavigatePayload{
		Route: "/screens/catalog",
	})
	cta := atoms.NewButton("btn_morning_shop", "Shop Breakfast", "secondary", &navCatalog)
	cta.Size = "sm"

	morningLayout := seedCardLayout()
	morningLayout.BgColor = atoms.ColorRef(theming.ColorInfoContainer)

	data := snippets.BannerSnippetData{
		Title:  atoms.NewText("Morning Breakfast Express • 6–11 AM", theming.TypographyTitleMedium, theming.ColorTextPrimary),
		Body:   ptrText(atoms.NewText("Warm stone-baked sourdough, chilled A2 milk & fresh berries before your first meeting.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Icon:   ptrIcon(atoms.NewIcon("wb_sunny", 24, theming.ColorInfo)),
		Tag:    ptrTag(atoms.NewTag("SCHEDULED", theming.ColorSurfacePrimary)),
		CTA:    &cta,
		Layout: morningLayout,
	}

	return domain.Widget{
		ID:          "w_morning_banner",
		TenantID:    tenantID,
		Slug:        "morning-banner",
		Name:        "Morning Breakfast Scheduled Banner",
		TemplateKey: "tmpl_banner_v1",
		SnippetType: snippets.SnippetTypeBanner,
		Props:       structToMap(data),
		Motion:      structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionSlideDown, DurationToken: atoms.MotionDurationNormal}),
		Tags:        []string{"banner", "scheduled", "slotting"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedETAMetricWidget(tenantID string, now time.Time) domain.Widget {
	toast := actions.NewShowToastAction("toast_eta_status", actions.ShowToastPayload{
		Message:    "Abeta Fresh dark store online · 8-minute cold-chain SLA active",
		DurationMs: 2200,
		Variant:    "info",
	})
	data := snippets.MetricCardSnippetData{
		Label: atoms.NewText("Live Dark-Store ETA", theming.TypographyLabelMedium, theming.ColorTextSecondary),
		Value: atoms.NewText("8 mins", theming.TypographyHeadlineMedium, theming.ColorTextPrimary),
		Delta: ptrTag(atoms.NewTag("-2 min", theming.ColorSuccessContainer)),
		Icon:  ptrIcon(atoms.NewIcon("local_shipping", 22, theming.ColorSuccess)),
		Progress: &atoms.ProgressAtom{
			Kind:     "linear",
			Progress: atoms.Float64Ptr(0.88),
			Label:    "88% riders ready in {{city}}",
			Color:    atoms.ColorRef(theming.ColorSuccess),
		},
		OnClick: &toast,
		Layout:  seedCardLayout(),
	}

	return domain.Widget{
		ID:          "w_eta_metric",
		TenantID:    tenantID,
		Slug:        "eta-metric",
		Name:        "Live Dark-Store ETA Metric",
		TemplateKey: "tmpl_metric_card",
		SnippetType: snippets.SnippetTypeMetricCard,
		Props:       structToMap(data),
		Tags:        []string{"metric", "eta", "watch"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func buildSeedProductWidget(tenantID string, now time.Time, spec seedProductSpec) domain.Widget {
	fieldID := "cart_qty_" + spec.productID
	incAct := actions.NewBatchAction("inc_"+spec.productID, actions.BatchPayload{
		Actions: []actions.ActionTrigger{
			actions.NewTriggerFeedbackAction("bounce_"+spec.productID, actions.TriggerFeedbackPayload{
				Effect:     actions.FeedbackCartBounce,
				TargetID:   "floating_cart_dock",
				DurationMs: 320,
			}),
			actions.NewTrackEventAction("add_to_cart_"+spec.productID, actions.TrackEventPayload{
				EventName: "add_to_cart",
				Properties: map[string]any{
					"product_id":       spec.productID,
					"field_id":         fieldID,
					"unit_price_cents": spec.unitPriceCents,
				},
			}),
		},
		StopOnError: true,
	})
	decAct := actions.NewTrackEventAction("dec_"+spec.productID, actions.TrackEventPayload{
		EventName: "cart_item_decremented",
		Properties: map[string]any{
			"product_id": spec.productID,
			"field_id":   fieldID,
		},
	})

	stepper := atoms.NewStepper(fieldID, "ADD", spec.initialQty, &incAct, &decAct)
	stepper.MinQuantity = 0
	stepper.MaxQuantity = 10

	previewToast := actions.NewShowToastAction("preview_"+spec.productID, actions.ShowToastPayload{
		Message:    fmt.Sprintf("%s (%s) · Delivered chilled in 8 mins", spec.title, spec.price),
		DurationMs: 2000,
		Variant:    "info",
	})

	data := snippets.ProductCardSnippetData{
		ProductID:      spec.productID,
		Title:          atoms.NewText(spec.title, theming.TypographyTitleSmall, theming.ColorTextPrimary),
		Subtitle:       ptrText(atoms.NewText(spec.subtitle, theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Price:          atoms.NewText(spec.price, theming.TypographyLabelLarge, theming.ColorTextPrimary),
		OriginalPrice:  ptrText(atoms.NewText(spec.originalPrice, theming.TypographyBodySmall, theming.ColorTextTertiary)),
		UnitPriceCents: spec.unitPriceCents,
		Badge:          ptrTag(atoms.NewTag(spec.badge, spec.badgeColor)),
		Image: &atoms.ImageAtom{
			URL:                fmt.Sprintf("https://cdn.abeta.dev/fresh/%s.png", spec.productID),
			ContentDescription: spec.title,
			HeightDp:           atoms.IntPtr(110),
			ScaleMode:          "crop",
			CornerRadius:       atoms.UniformRadius(theming.RadiusSM.Dp()),
		},
		Stepper: stepper,
		Rating:  ptrText(atoms.NewText(spec.rating, theming.TypographyLabelSmall, theming.ColorSuccess)),
		OnClick: &previewToast,
		Layout:  seedCompactCardLayout(),
	}

	return domain.Widget{
		ID:          spec.widgetID,
		TenantID:    tenantID,
		Slug:        spec.slug,
		Name:        spec.title,
		TemplateKey: "tmpl_product_card",
		SnippetType: snippets.SnippetTypeProductCard,
		Props:       structToMap(data),
		Motion: structToMap(atoms.MotionSpec{
			Enter:         atoms.MotionTransitionScaleIn,
			DurationToken: atoms.MotionDurationFast,
			StaggerMs:     60,
			DelayMs:       spec.delayMs,
		}),
		Tracking: map[string]any{
			"event_name": "product_impression",
			"properties": map[string]any{"product_id": spec.productID, "unit_price_cents": spec.unitPriceCents},
		},
		Tags:      []string{"product", "commerce"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
