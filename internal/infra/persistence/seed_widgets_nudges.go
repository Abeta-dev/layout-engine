package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

func seedFreshOfferMarqueeWidget(tenantID string, now time.Time) domain.Widget {
	promoToast := actions.NewShowToastAction("toast_fresh_marquee", actions.ShowToastPayload{
		Message:    "FRESH20 applied — extra 20% off chilled essentials in {{city}}!",
		DurationMs: 2200,
		Variant:    "success",
	})
	navFlash := actions.NewNavigateAction("nav_marquee_flash", actions.NavigatePayload{
		Route: "/screens/flash-sale",
	})

	data := snippets.MarqueeStripSnippetData{
		StripID:          "fresh_offer_marquee",
		VelocityDpPerSec: 48,
		Separator:        "•",
		Items: []snippets.MarqueeItem{
			{
				ID:      "mq_eta",
				Text:    "8-Min Chilled Delivery in {{city}}",
				Icon:    ptrIcon(atoms.NewIcon("bolt", 16, theming.ColorBrandPrimary)),
				Badge:   ptrTag(atoms.NewTag("LIVE", theming.ColorSuccessContainer)),
				OnClick: &navFlash,
			},
			{
				ID:      "mq_free_del",
				Text:    "Free Delivery on orders over $15",
				Icon:    ptrIcon(atoms.NewIcon("local_shipping", 16, theming.ColorSuccess)),
				OnClick: &promoToast,
			},
			{
				ID:      "mq_fresh20",
				Text:    "Use code FRESH20 for 20% off first basket",
				Icon:    ptrIcon(atoms.NewIcon("sparkles", 16, theming.ColorWarning)),
				Badge:   ptrTag(atoms.NewTag("PROMO", theming.ColorWarningContainer)),
				OnClick: &promoToast,
			},
			{
				ID:   "mq_cold_chain",
				Text: "Farm-to-door 2–4°C cold-chain freshness",
				Icon: ptrIcon(atoms.NewIcon("verified", 16, theming.ColorInfo)),
			},
		},
		Layout: &atoms.LayoutConfig{
			Padding:      atoms.SymmetricSpacing(theming.SpacingMD.Dp(), theming.SpacingXS.Dp()),
			FillMaxWidth: atoms.BoolPtr(true),
		},
	}

	return domain.Widget{
		ID:          "w_fresh_offer_marquee",
		TenantID:    tenantID,
		Slug:        "fresh-offer-marquee",
		Name:        "Abeta Fresh Offer Ticker",
		TemplateKey: "tmpl_marquee_strip",
		SnippetType: snippets.SnippetTypeMarqueeStrip,
		Props:       structToMap(data),
		Motion:      structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionFadeIn, DurationToken: atoms.MotionDurationFast, LoopEffect: atoms.MotionLoopMarquee}),
		Tracking: map[string]any{
			"event_name": "marquee_strip_viewed",
			"properties": map[string]any{"strip_id": "fresh_offer_marquee", "vertical": "abeta-fresh"},
		},
		Tags:      []string{"marquee_strip", "promo", "fresh"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func seedFreshFlashNudgeWidget(tenantID string, now time.Time) domain.Widget {
	navFlash := actions.NewNavigateAction("nav_fresh_flash_nudge", actions.NavigatePayload{
		Route: "/screens/flash-sale",
	})
	dismissTrack := actions.NewTrackEventAction("dismiss_fresh_flash_nudge", actions.TrackEventPayload{
		EventName:  "nudge_dismissed",
		Properties: map[string]any{"nudge_id": "fresh_flash_drop_nudge", "vertical": "abeta-fresh"},
	})

	badge := atoms.NewTag("FLASH DROP", theming.ColorErrorContainer)
	badge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	countdown := atoms.NewCountdown("fresh_flash_drop_timer", 540, "Ends in")
	countdown.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopHeartbeat}

	cta := atoms.NewButton("btn_fresh_flash_nudge", "Claim Drop", "primary", &navFlash)
	cta.Size = "sm"

	nudgeLayout := seedCardLayout()
	nudgeLayout.BgColor = atoms.ColorRef(theming.ColorWarningContainer)

	data := snippets.NudgeBarSnippetData{
		NudgeID:     "fresh_flash_drop_nudge",
		Title:       atoms.NewText("Flash Drop in {{city}} · Save up to 30%", theming.TypographyTitleSmall, theming.ColorTextPrimary),
		Subtitle:    ptrText(atoms.NewText("Alphonso Mango Box & organic berries stocked for the next 9 minutes.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Icon:        ptrIcon(atoms.NewIcon("bolt", 22, theming.ColorWarning)),
		Badge:       &badge,
		Countdown:   &countdown,
		Progress:    atoms.Float64Ptr(0.72),
		CTA:         &cta,
		Dismissible: true,
		PointerEdge: snippets.PointerEdgeNone,
		OnDismiss:   &dismissTrack,
		OnClick:     &navFlash,
		Layout:      nudgeLayout,
	}

	return domain.Widget{
		ID:          "w_fresh_flash_nudge",
		TenantID:    tenantID,
		Slug:        "fresh-flash-nudge",
		Name:        "Flash Drop Urgency Nudge Bar",
		TemplateKey: "tmpl_nudge_bar",
		SnippetType: snippets.SnippetTypeNudgeBar,
		Props:       structToMap(data),
		Motion:      structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionSlideDown, DurationToken: atoms.MotionDurationFast}),
		Tracking: map[string]any{
			"event_name": "nudge_bar_viewed",
			"properties": map[string]any{"nudge_id": "fresh_flash_drop_nudge", "vertical": "abeta-fresh"},
		},
		Tags:      []string{"nudge_bar", "urgency", "flash"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func seedFreshLoyaltyComposedWidget(tenantID string, now time.Time) domain.Widget {
	claimToast := actions.NewShowToastAction("toast_fresh_loyalty_claim", actions.ShowToastPayload{
		Message:    "You have 4/6 Abeta Fresh stamps — 2 more express orders unlock a free berry & sourdough box!",
		DurationMs: 2400,
		Variant:    "success",
	})
	moldedEnv := snippets.MoldLoyaltyStampRecipe(
		"w_fresh_loyalty_composed",
		"fresh_loyalty_stamps",
		"Abeta Fresh Punch Card · {{city}}",
		4,
		6,
		"Collect 2 more express orders for a complimentary berry & sourdough box",
		claimToast,
	)

	return domain.Widget{
		ID:          "w_fresh_loyalty_composed",
		TenantID:    tenantID,
		Slug:        "fresh-loyalty-composed",
		Name:        "Abeta Fresh Loyalty Stamp Card",
		TemplateKey: "tmpl_composed_block",
		SnippetType: snippets.SnippetTypeComposedBlock,
		Props:       structToMap(moldedEnv.Data),
		Motion:      structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionScaleIn, DurationToken: atoms.MotionDurationNormal, LoopEffect: atoms.MotionLoopGlowRing}),
		Tracking: map[string]any{
			"event_name": "composed_block_viewed",
			"properties": map[string]any{"recipe": "loyalty_stamp_card", "vertical": "abeta-fresh"},
		},
		Tags:      []string{"composed_block", "loyalty", "molded"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
