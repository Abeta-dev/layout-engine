package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

func seedFreshGridWidget(tenantID string, now time.Time) domain.Widget {
	props := map[string]any{
		"title":   structToMap(atoms.NewText("8-Min Fresh Essentials", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
		"columns": 2,
		"columns_by_width": map[string]any{
			"compact":  2,
			"medium":   3,
			"expanded": 4,
		},
		"spacing_dp":       12,
		"child_widget_ids": []string{"w_prod_avocado", "w_prod_milk", "w_prod_sourdough", "w_prod_espresso"},
		"layout": structToMap(atoms.LayoutConfig{
			Padding:      atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingSM.Dp()),
			FillMaxWidth: atoms.BoolPtr(true),
		}),
	}

	return domain.Widget{
		ID:          "w_fresh_grid",
		TenantID:    tenantID,
		Slug:        "fresh-grid",
		Name:        "Fresh Essentials Responsive Grid",
		TemplateKey: "tmpl_grid",
		SnippetType: snippets.SnippetTypeGrid,
		Props:       props,
		Tracking:    map[string]any{"event_name": "grid_viewed", "properties": map[string]any{"grid_id": "w_fresh_grid"}},
		Tags:        []string{"grid", "responsive", "commerce"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedBakeryGridWidget(tenantID string, now time.Time) domain.Widget {
	props := map[string]any{
		"title":   structToMap(atoms.NewText("Bakery & Espresso Bar", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
		"columns": 2,
		"columns_by_width": map[string]any{
			"compact":  2,
			"medium":   3,
			"expanded": 4,
		},
		"spacing_dp":       12,
		"child_widget_ids": []string{"w_prod_sourdough", "w_prod_espresso", "w_prod_milk"},
		"layout": structToMap(atoms.LayoutConfig{
			Padding:      atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingSM.Dp()),
			FillMaxWidth: atoms.BoolPtr(true),
		}),
	}

	return domain.Widget{
		ID:          "w_bakery_grid",
		TenantID:    tenantID,
		Slug:        "bakery-grid",
		Name:        "Bakery & Brew Responsive Grid",
		TemplateKey: "tmpl_grid",
		SnippetType: snippets.SnippetTypeGrid,
		Props:       props,
		Tags:        []string{"grid", "bakery", "commerce"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedFlashCarouselWidget(tenantID string, now time.Time) domain.Widget {
	props := map[string]any{
		"title":         structToMap(atoms.NewText("Lightning Flash Deals", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
		"item_width_dp": 220,
		"item_width_by_width": map[string]any{
			"compact":  220,
			"medium":   240,
			"expanded": 260,
		},
		"spacing_dp":       12,
		"child_widget_ids": []string{"w_prod_mango", "w_prod_berries", "w_prod_avocado"},
		"layout": structToMap(atoms.LayoutConfig{
			Padding:      atoms.SymmetricSpacing(theming.SpacingLG.Dp(), theming.SpacingSM.Dp()),
			FillMaxWidth: atoms.BoolPtr(true),
		}),
	}

	return domain.Widget{
		ID:          "w_flash_carousel",
		TenantID:    tenantID,
		Slug:        "flash-carousel",
		Name:        "Flash Deals Carousel",
		TemplateKey: "tmpl_carousel",
		SnippetType: snippets.SnippetTypeCarousel,
		Props:       props,
		Tracking:    map[string]any{"event_name": "carousel_viewed", "properties": map[string]any{"carousel_id": "w_flash_carousel"}},
		Tags:        []string{"carousel", "responsive", "flash"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedQuickActionsWidget(tenantID string, now time.Time) domain.Widget {
	navFlash := actions.NewNavigateAction("nav_quick_flash", actions.NavigatePayload{
		Route: "/screens/flash-sale",
	})
	data := snippets.ActionRowSnippetData{
		Title:        atoms.NewText("Flash Sale • Browse Catalog • Feedback Form • Watch Glance", theming.TypographyTitleSmall, theming.ColorTextPrimary),
		Subtitle:     ptrText(atoms.NewText("Jump across live Abeta Fresh Express screens", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		LeadingIcon:  ptrIcon(atoms.NewIcon("bolt", 22, theming.ColorBrandPrimary)),
		TrailingTag:  ptrTag(atoms.NewTag("EXPLORE", theming.ColorBrandContainer)),
		TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 20, theming.ColorTextTertiary)),
		OnClick:      &navFlash,
		Layout:       seedCardLayout(),
	}

	return domain.Widget{
		ID:          "w_quick_actions",
		TenantID:    tenantID,
		Slug:        "quick-actions",
		Name:        "Quick Screen Actions Row",
		TemplateKey: "tmpl_action_row",
		SnippetType: snippets.SnippetTypeActionRow,
		Props:       structToMap(data),
		Tags:        []string{"action_row", "navigation"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedCategoryTabsWidget(tenantID string, now time.Time) domain.Widget {
	props := map[string]any{
		"selected_index": 0,
		"tabs": []map[string]any{
			{
				"id":         "fresh_produce",
				"title":      "Fresh Produce",
				"badge":      structToMap(atoms.NewTag("FRESH", theming.ColorSuccessContainer)),
				"widget_ids": []string{"w_fresh_grid"},
			},
			{
				"id":         "bakery_brew",
				"title":      "Bakery & Brew",
				"badge":      structToMap(atoms.NewTag("WARM", theming.ColorWarningContainer)),
				"widget_ids": []string{"w_bakery_grid"},
			},
		},
		"layout": structToMap(atoms.LayoutConfig{
			Padding:      atoms.SymmetricSpacing(theming.SpacingMD.Dp(), theming.SpacingXS.Dp()),
			FillMaxWidth: atoms.BoolPtr(true),
		}),
	}

	return domain.Widget{
		ID:          "w_category_tabs",
		TenantID:    tenantID,
		Slug:        "category-tabs",
		Name:        "Store Category Tabs",
		TemplateKey: "tmpl_tabs",
		SnippetType: snippets.SnippetTypeTabs,
		Props:       props,
		Tags:        []string{"tabs", "navigation", "commerce"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedDeliveryFAQWidget(tenantID string, now time.Time) domain.Widget {
	slaToast := actions.NewShowToastAction("toast_sla_promise", actions.ShowToastPayload{
		Message:    "8-Minute Promise: Late orders automatically receive instant credit!",
		DurationMs: 2400,
		Variant:    "success",
	})
	coldToast := actions.NewShowToastAction("toast_cold_chain", actions.ShowToastPayload{
		Message:    "Cold-Chain Quality: Chilled 2–4°C insulated dispatch from dark store.",
		DurationMs: 2400,
		Variant:    "info",
	})

	promiseRow := snippets.NewActionRowSnippet(
		"faq_8min_promise",
		atoms.NewText("8-Minute Promise", theming.TypographyTitleSmall, theming.ColorTextPrimary),
	).
		WithParentID("w_delivery_faq").
		WithSubtitle(atoms.NewText("Hyperlocal dark stores dispatch in under 90 seconds from order confirmation.", theming.TypographyBodySmall, theming.ColorTextSecondary)).
		WithLeadingIcon(atoms.NewIcon("bolt", 22, theming.ColorBrandPrimary)).
		WithTrailingTag(atoms.NewTag("SLA", theming.ColorSuccessContainer)).
		WithOnClick(slaToast).
		WithLayout(*seedCardLayout()).
		Build()

	coldChainRow := snippets.NewActionRowSnippet(
		"faq_cold_chain",
		atoms.NewText("Cold-Chain Quality", theming.TypographyTitleSmall, theming.ColorTextPrimary),
	).
		WithParentID("w_delivery_faq").
		WithSubtitle(atoms.NewText("Tamper-evident chilled bags keep dairy, berries & produce crisp in {{city}}.", theming.TypographyBodySmall, theming.ColorTextSecondary)).
		WithLeadingIcon(atoms.NewIcon("verified", 22, theming.ColorSuccess)).
		WithTrailingTag(atoms.NewTag("VERIFIED", theming.ColorInfoContainer)).
		WithOnClick(coldToast).
		WithLayout(*seedCardLayout()).
		Build()

	data := snippets.AccordionSnippetData{
		Title:             atoms.NewText("8-Minute Promise & Cold-Chain Quality", theming.TypographyTitleMedium, theming.ColorTextPrimary),
		Subtitle:          ptrText(atoms.NewText("Delivery guarantee & freshness standards", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		InitiallyExpanded: true,
		Snippets:          []snippets.SnippetEnvelope{promiseRow, coldChainRow},
		Layout:            seedCardLayout(),
	}

	return domain.Widget{
		ID:          "w_delivery_faq",
		TenantID:    tenantID,
		Slug:        "delivery-faq",
		Name:        "Delivery Promise & Cold-Chain FAQ",
		TemplateKey: "tmpl_accordion",
		SnippetType: snippets.SnippetTypeAccordion,
		Props:       structToMap(data),
		Tags:        []string{"accordion", "faq", "trust"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedFeedbackFormWidget(tenantID string, now time.Time) domain.Widget {
	successToast := actions.NewShowToastAction("toast_feedback_saved", actions.ShowToastPayload{
		Message:    "Thank you! Your Abeta Fresh feedback was recorded.",
		DurationMs: 2400,
		Variant:    "success",
	})
	submitAct := actions.NewCallAPIAction("submit_order_feedback", actions.CallAPIPayload{
		Endpoint:  "/api/v1/actions/execute",
		Method:    "POST",
		Body:      map[string]any{"form_id": "express_order_feedback"},
		OnSuccess: &successToast,
	}).WithValidations(
		actions.ValidationRule{
			FieldID: "full_name",
			Rule:    "required",
			Message: "Please enter your name",
		},
		actions.ValidationRule{
			FieldID: "work_email",
			Rule:    "regex",
			Value:   `^[^@\s]+@[^@\s]+\.[^@\s]+$`,
			Message: "Enter a valid email address",
		},
	).WithSecondaryActions(
		actions.NewTriggerFeedbackAction("feedback_haptic", actions.TriggerFeedbackPayload{
			Effect: actions.FeedbackHapticTick,
		}),
	)

	savePref := actions.NewSaveStateAction("save_seal_verified", actions.SaveStatePayload{
		Key:   "cold_chain_seal_verified",
		Value: true,
		Scope: "session",
	})
	sealSwitch := atoms.NewSwitch("cold_chain_seal_verified", "Cold-Chain Seal Intact", true, &savePref)
	sealSwitch.Subtitle = "Confirm chilled bag arrived sealed"

	leaveDoorCheckbox := atoms.NewCheckbox("leave_at_door", "Allow contactless drop-off", true, &savePref)

	ratingSlider := atoms.NewSlider("delivery_rating", "Express Speed Rating", 5, 1, 5, &savePref)
	ratingSlider.Steps = 4
	ratingSlider.UnitLabel = "★"

	data := snippets.FormSnippetData{
		FormID:      "express_order_feedback",
		Title:       ptrText(atoms.NewText("Quick Order Feedback", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
		Description: ptrText(atoms.NewText("Help our dark-store team keep 8-minute deliveries spotless.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Fields: []atoms.TextFieldAtom{
			{
				FieldID:      "full_name",
				Label:        "Full Name",
				Placeholder:  "Alex Rivera",
				Required:     true,
				KeyboardType: "text",
			},
			{
				FieldID:      "work_email",
				Label:        "Email",
				Placeholder:  "alex@abeta.dev",
				Required:     false,
				Regex:        `^[^@\s]+@[^@\s]+\.[^@\s]+$`,
				KeyboardType: "email",
			},
			{
				FieldID:      "delivery_notes",
				Label:        "Delivery Notes",
				Placeholder:  "Super crisp produce and chilled milk!",
				KeyboardType: "text",
			},
		},
		Switches:     []atoms.SwitchAtom{sealSwitch},
		Checkboxes:   []atoms.CheckboxAtom{leaveDoorCheckbox},
		Sliders:      []atoms.SliderAtom{ratingSlider},
		SubmitButton: atoms.NewButton("btn_submit_feedback", "Submit Feedback", "primary", &submitAct),
		Layout:       seedCardLayout(),
	}

	return domain.Widget{
		ID:          "w_feedback_form",
		TenantID:    tenantID,
		Slug:        "feedback-form",
		Name:        "Express Order Feedback Form",
		TemplateKey: "tmpl_form_v1",
		SnippetType: snippets.SnippetTypeForm,
		Props:       structToMap(data),
		Tags:        []string{"form", "feedback"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func seedWatchStatusCardWidget(tenantID string, now time.Time) domain.Widget {
	trackAct := actions.NewBatchAction("watch_track_live", actions.BatchPayload{
		Actions: []actions.ActionTrigger{
			actions.NewTriggerFeedbackAction("watch_track_haptic", actions.TriggerFeedbackPayload{
				Effect: actions.FeedbackHapticTick,
			}),
			actions.NewShowToastAction("watch_track_toast", actions.ShowToastPayload{
				Message:    "Order #AF-409 is 400m away · Arriving in 6 mins",
				DurationMs: 2200,
				Variant:    "success",
			}),
		},
		StopOnError: true,
	})
	shareAct := actions.NewShareAction("watch_share_eta", actions.SharePayload{
		Title: "Abeta Fresh Delivery",
		Text:  "My Abeta Fresh Express order #AF-409 arrives in 6 minutes!",
		URL:   "https://app.abeta.dev/screens/store-home",
	})

	btnTrack := atoms.NewButton("btn_watch_track", "Track Live", "primary", &trackAct)
	btnTrack.Size = "sm"
	btnShare := atoms.NewButton("btn_watch_share", "Share ETA", "outline", &shareAct)
	btnShare.Size = "sm"

	liveBadge := atoms.NewTag("LIVE", theming.ColorSuccessContainer)
	liveBadge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	data := snippets.CardSnippetData{
		Title:           atoms.NewText("Order #AF-409 • Arriving in 6 min", theming.TypographyTitleSmall, theming.ColorTextPrimary),
		Subtitle:        ptrText(atoms.NewText("Abeta Fresh Hub · Rider en route", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Badge:           &liveBadge,
		PrimaryButton:   &btnTrack,
		SecondaryButton: &btnShare,
		Layout:          seedCompactCardLayout(),
	}

	return domain.Widget{
		ID:          "w_watch_status_card",
		TenantID:    tenantID,
		Slug:        "watch-status-card",
		Name:        "Watch Active Order Glance Card",
		TemplateKey: "tmpl_card_v1",
		SnippetType: snippets.SnippetTypeCard,
		Props:       structToMap(data),
		Tags:        []string{"watch", "card", "order"},
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
