package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// ── Abeta Pay (Instant Payments & Smart Cards) ────────────────────────────────

const (
	payRecipientRegex = `^(@[a-zA-Z0-9._-]{2,}|[^@\s]+@[^@\s]+\.[^@\s]+)$`
	payAmountRegex    = `^\d+(\.\d{1,2})?$`
)

func buildPayWorkspaceSpec(now time.Time) seedWorkspaceSpec {
	const tid = "tenant_abeta_pay"
	return seedWorkspaceSpec{
		Tenant: domain.Tenant{
			ID:          tid,
			Slug:        "abeta-pay",
			Name:        "Abeta Pay — Instant Payments & Smart Cards",
			Description: "Zero-fee instant P2P wallet, Abeta Metal smart cards, biometric authorization, and real-time clearing",
			Active:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		TemplateIDPrefix: "pay",
		Widgets:          seedPayWidgets(tid, now),
		Rules:            seedPayRules(tid, now),
		SlottingRules:    seedPaySlottingRules(tid, now),
		Layouts:          seedPayLayouts(tid, now),
		SplashConfigs: []domain.SplashConfig{
			{
				ID: "splash_default_pay", TenantID: tid, Key: "default_splash",
				ImageURL: "https://cdn.abeta.dev/pay/splash-default.png", BackgroundColorHex: "#1D4ED8",
				DurationMs: 1300, IsDefault: true, Enabled: true, UpdatedAt: now,
			},
			{
				ID: "splash_metal_pay", TenantID: tid, Key: "metal_splash",
				ImageURL: "https://cdn.abeta.dev/pay/splash-metal.png", BackgroundColorHex: "#0F172A",
				DurationMs: 1500, RuleID: "rule_pay_metal_vip", IsDefault: false, Enabled: true, UpdatedAt: now,
			},
		},
		TokenOverrides: []domain.TokenOverride{
			{
				ID: "tok_default_pay", TenantID: tid, Key: "default_pay", ThemePreset: "abeta-pay",
				Tokens: map[string]string{
					string(theming.ColorBrandPrimary):   "#2563EB",
					string(theming.ColorBrandContainer): "#DBEAFE",
					string(theming.ColorAccentPrimary):  "#4F46E5",
					string(theming.ColorSuccess):        "#059669",
				},
				Enabled: true, UpdatedAt: now,
			},
			{
				ID: "tok_metal_pay", TenantID: tid, Key: "metal_obsidian", ThemePreset: "abeta-pay-metal",
				Tokens: map[string]string{
					string(theming.ColorBrandPrimary):   "#1E293B",
					string(theming.ColorBrandContainer): "#E2E8F0",
					string(theming.ColorAccentPrimary):  "#0EA5E9",
				},
				RuleID: "rule_pay_metal_vip", Enabled: true, UpdatedAt: now,
			},
		},
		ChangeLogID:      "chg_seed_abeta_pay",
		ChangeLogSummary: "Bootstrapped Abeta Pay payments & smart cards workspace with 4 published layouts, 21 templates, and molded extended-atom widgets",
	}
}

func seedPayWidgets(tenantID string, now time.Time) []domain.Widget {
	vipBannerLayout := seedCardLayout()
	vipBannerLayout.BgColor = atoms.ColorRef(theming.ColorBrandContainer)
	vipToast := actions.NewShowToastAction("toast_pay_metal_boost", actions.ShowToastPayload{
		Message: "Abeta Metal 5% instant cashback & 0% FX boost active!", DurationMs: 2200, Variant: "success",
	})
	vipCTA := atoms.NewButton("btn_pay_metal_boost", "Activate 5% Boost", "primary", &vipToast)
	vipCTA.Size = "sm"

	navTransfer := actions.NewNavigateAction("nav_pay_quick_send", actions.NavigatePayload{Route: "/screens/pay-transfer"})
	navCards := actions.NewNavigateAction("nav_pay_card_controls", actions.NavigatePayload{Route: "/screens/pay-cards"})
	navReceipt := actions.NewNavigateAction("nav_pay_latest_receipt", actions.NavigatePayload{Route: "/screens/pay-receipt"})

	shareAct := actions.NewShareAction("share_pay_receipt_3081", actions.SharePayload{
		Title: "Abeta Pay Receipt #PAY-3081",
		Text:  "Sent $125.00 to @maya.chen via Abeta Pay (Zero Fee · Ref #PAY-3081)",
		URL:   "https://abeta.dev/pay/receipt/PAY-3081",
	})

	widgets := []domain.Widget{
		buildSeedSearchBarWidget(
			tenantID, "w_pay_search_bar", "pay-search-bar", "Abeta Pay Contacts & Bills Search",
			"pay_search", "Search contacts, @handles & bills in {{city}}...", "pay-home", now,
		),
		seedPayPerksMarqueeWidget(tenantID, now),
		buildSeedChipGroupWidget(
			tenantID, "w_pay_quick_chips", "pay-quick-chips", "Abeta Pay Quick Actions Chips",
			"pay_filters", "pay-home",
			[]seedChipSpec{
				{id: "all", label: "Wallet", selected: true},
				{id: "send", label: "Send Money"},
				{id: "cards", label: "Smart Cards"},
				{id: "bills", label: "Bill Pay"},
				{id: "receipts", label: "Receipt #PAY-3081"},
			}, now,
		),
		seedPayCashbackNudgeWidget(tenantID, now),
		{
			ID: "w_pay_metal_vip_banner", TenantID: tenantID, Slug: "pay-metal-vip-banner",
			Name: "Abeta Metal VIP Concierge Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("Abeta Metal Concierge • {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Instant $25,000 P2P limits, 5% cashback & zero international FX markup.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("verified", 24, theming.ColorBrandPrimary)),
				Tag:    ptrTag(atoms.NewTag("METAL VIP", theming.ColorSuccessContainer)),
				CTA:    &vipCTA,
				Layout: vipBannerLayout,
			}),
			Tracking: map[string]any{"event_name": "slotting_banner_viewed", "properties": map[string]any{"vertical": "abeta-pay", "slot": "metal_vip"}},
			Tags:     []string{"banner", "vip", "slotting"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		seedPayMetalCardComposedWidget(tenantID, now),
		seedPayReceiveQrComposedWidget(tenantID, now),
		{
			ID: "w_pay_quick_send_row", TenantID: tenantID, Slug: "pay-quick-send-row",
			Name: "Quick Send P2P Action Row", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:        atoms.NewText("Quick Send P2P", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:     ptrText(atoms.NewText("Zero-fee instant transfer to @maya.chen, @arjun.rao or any handle", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon:  ptrIcon(atoms.NewIcon("bolt", 20, theming.ColorBrandPrimary)),
				TrailingTag:  ptrTag(atoms.NewTag("INSTANT", theming.ColorSuccessContainer)),
				TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 18, theming.ColorTextTertiary)),
				OnClick:      &navTransfer,
				Layout:       seedCardLayout(),
			}),
			Tags: []string{"action_row", "transfer"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_pay_card_controls_row", TenantID: tenantID, Slug: "pay-card-controls-row",
			Name: "Smart Card Controls Action Row", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:        atoms.NewText("Smart Card Controls", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:     ptrText(atoms.NewText("Freeze card, international spend, contactless & monthly limit", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon:  ptrIcon(atoms.NewIcon("settings", 20, theming.ColorBrandPrimary)),
				TrailingTag:  ptrTag(atoms.NewTag("ACTIVE", theming.ColorInfoContainer)),
				TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 18, theming.ColorTextTertiary)),
				OnClick:      &navCards,
				Layout:       seedCardLayout(),
			}),
			Tags: []string{"action_row", "cards"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_pay_recent_receipt_row", TenantID: tenantID, Slug: "pay-recent-receipt-row",
			Name: "Latest Settlement #PAY-3081 Row", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:        atoms.NewText("Latest Settlement #PAY-3081", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:     ptrText(atoms.NewText("$125.00 to @maya.chen · 4-stage live clearing tracker", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon:  ptrIcon(atoms.NewIcon("check_circle", 20, theming.ColorSuccess)),
				TrailingTag:  ptrTag(atoms.NewTag("SETTLED", theming.ColorSuccessContainer)),
				TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 18, theming.ColorTextTertiary)),
				OnClick:      &navReceipt,
				Layout:       seedCardLayout(),
			}),
			Tags: []string{"action_row", "receipt"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_pay_bill_power", slug: "pay-bill-power", name: "City Power & Grid Bill Card",
			title: "City Power & Grid", subtitle: "Auto-pay due in 3 days",
			details: "$68.40 · 5% cashback on Abeta Metal", badge: "BILL PAY",
			badgeColor: theming.ColorBrandContainer, ctaLabel: "Pay Now", route: "pay-transfer",
			tags: []string{"bill", "card"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_pay_bill_fiber", slug: "pay-bill-fiber", name: "HyperFiber Gigabit Bill Card",
			title: "HyperFiber Gigabit", subtitle: "Monthly fiber broadband",
			details: "$59.00 · Zero convenience fee", badge: "ZERO FEE",
			badgeColor: theming.ColorSuccessContainer, ctaLabel: "Pay Now", route: "pay-transfer",
			tags: []string{"bill", "card"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_pay_bill_mobile", slug: "pay-bill-mobile", name: "Abeta Wireless 5G Bill Card",
			title: "Abeta Wireless 5G", subtitle: "Unlimited 5G plan",
			details: "$45.00 · Instant top-up & receipt", badge: "INSTANT",
			badgeColor: theming.ColorInfoContainer, ctaLabel: "Top Up", route: "pay-transfer",
			tags: []string{"bill", "card"},
		}),
		{
			ID: "w_pay_bill_carousel", TenantID: tenantID, Slug: "pay-bill-carousel",
			Name: "Bill Pay & Utilities Carousel", TemplateKey: "tmpl_carousel", SnippetType: snippets.SnippetTypeCarousel,
			Props: map[string]any{
				"title":               structToMap(atoms.NewText("Bill Pay & Utilities", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
				"item_width_dp":       240,
				"item_width_by_width": map[string]any{"compact": 220, "medium": 260, "expanded": 280},
				"spacing_dp":          12,
				"child_widget_ids":    []string{"w_pay_bill_power", "w_pay_bill_fiber", "w_pay_bill_mobile"},
				"layout":              structToMap(seedHorizontalPaddingLayout()),
			},
			Tags: []string{"carousel", "bills"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		// ── pay-transfer widgets ──
		{
			ID: "w_pay_transfer_header", TenantID: tenantID, Slug: "pay-transfer-header",
			Name: "Instant P2P Transfer Header", TemplateKey: "tmpl_header_v1", SnippetType: snippets.SnippetTypeHeader,
			Props: structToMap(snippets.HeaderSnippetData{
				Title:       atoms.NewText("Instant P2P Transfer", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Available balance: $2,480.50 · Settles in < 2 seconds in {{city}}", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag: ptrTag(atoms.NewTag("ZERO FEE", theming.ColorSuccessContainer)),
				Layout:      seedCardLayout(),
			}),
			Tags: []string{"header", "transfer"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		buildSeedChipGroupWidget(
			tenantID, "w_pay_contact_chips", "pay-contact-chips", "Recent P2P Contacts Chips",
			"pay_recent_contacts", "pay-transfer",
			[]seedChipSpec{
				{id: "maya", label: "@maya.chen", selected: true},
				{id: "arjun", label: "@arjun.rao"},
				{id: "elena", label: "@elena.m"},
				{id: "noah", label: "@noah.park"},
			}, now,
		),
		seedPayTransferMoldedControlsWidget(tenantID, now),
		seedPayTransferFormWidget(tenantID, now),
		seedPayBiometricRowWidget(tenantID, now),
		// ── pay-cards widgets ──
		{
			ID: "w_pay_cards_header", TenantID: tenantID, Slug: "pay-cards-header",
			Name: "Abeta Metal Virtual Card Header", TemplateKey: "tmpl_header_v1", SnippetType: snippets.SnippetTypeHeader,
			Props: structToMap(snippets.HeaderSnippetData{
				Title:       atoms.NewText("Abeta Metal Virtual Card", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Instant real-time card security & spending controls", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag: ptrTag(atoms.NewTag("ACTIVE · •••• 4092", theming.ColorSuccessContainer)),
				Layout:      seedCardLayout(),
			}),
			Tags: []string{"header", "cards"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_pay_cards_spend_metric", TenantID: tenantID, Slug: "pay-cards-spend-metric",
			Name: "Monthly Card Spend Metric", TemplateKey: "tmpl_metric_card", SnippetType: snippets.SnippetTypeMetricCard,
			Props: structToMap(snippets.MetricCardSnippetData{
				Label: atoms.NewText("Monthly Card Spend", theming.TypographyLabelMedium, theming.ColorTextSecondary),
				Value: atoms.NewText("$1,420.00", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Delta: ptrTag(atoms.NewTag("28% OF LIMIT", theming.ColorSuccessContainer)),
				Icon:  ptrIcon(atoms.NewIcon("settings", 20, theming.ColorBrandPrimary)),
				Progress: &atoms.ProgressAtom{
					Kind: "linear", Progress: atoms.Float64Ptr(0.284), Label: "$1,420 / $5,000 monthly spend limit",
					Color: atoms.ColorRef(theming.ColorBrandPrimary),
				},
				Layout: seedCardLayout(),
			}),
			Tags: []string{"metric", "cards"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		seedPayCardsFormWidget(tenantID, now),
		// ── pay-receipt widgets ──
		seedPaySettlementTrackerWidget(tenantID, shareAct, now),
		seedPayReceiptLedgerComposedWidget(tenantID, now),
		{
			ID: "w_pay_receipt_share_row", TenantID: tenantID, Slug: "pay-receipt-share-row",
			Name: "Share Settlement Receipt Row", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:        atoms.NewText("Share Settlement Receipt #PAY-3081", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:     ptrText(atoms.NewText("Send verified zero-fee transfer confirmation via OS share sheet", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon:  ptrIcon(atoms.NewIcon("info", 20, theming.ColorBrandPrimary)),
				TrailingTag:  ptrTag(atoms.NewTag("SHARE", theming.ColorBrandContainer)),
				TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 18, theming.ColorTextTertiary)),
				OnClick:      &shareAct,
				Layout:       seedCardLayout(),
			}),
			Tags: []string{"action_row", "receipt", "share"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
	}

	return widgets
}

func seedPayPerksMarqueeWidget(tenantID string, now time.Time) domain.Widget {
	instantToast := actions.NewShowToastAction("tap_pay_instant", actions.ShowToastPayload{
		Message: "Instant P2P transfers settle in < 2s", DurationMs: 2200, Variant: "info",
	})
	cashbackToast := actions.NewShowToastAction("tap_pay_cashback", actions.ShowToastPayload{
		Message: "5% cashback active on Abeta Metal!", DurationMs: 2200, Variant: "success",
	})
	data := snippets.MarqueeStripSnippetData{
		StripID: "pay_live_perks", VelocityDpPerSec: 48, Separator: "•",
		Items: []snippets.MarqueeItem{
			{ID: "pay_mq_instant", Text: "Zero-fee instant P2P transfers in {{city}}", Icon: ptrIcon(atoms.NewIcon("bolt", 16, theming.ColorBrandPrimary)), Badge: ptrTag(atoms.NewTag("INSTANT", theming.ColorSuccessContainer)), OnClick: &instantToast},
			{ID: "pay_mq_cashback", Text: "5% cashback on Abeta Metal Card", Icon: ptrIcon(atoms.NewIcon("star", 16, theming.ColorSuccess)), Badge: ptrTag(atoms.NewTag("5% BACK", theming.ColorBrandContainer)), OnClick: &cashbackToast},
			{ID: "pay_mq_fx", Text: "0% international FX markup globally", Icon: ptrIcon(atoms.NewIcon("verified", 16, theming.ColorInfo)), Badge: ptrTag(atoms.NewTag("GLOBAL", theming.ColorInfoContainer))},
		},
		Layout: &atoms.LayoutConfig{Padding: atoms.SymmetricSpacing(theming.SpacingMD.Dp(), theming.SpacingXS.Dp()), FillMaxWidth: atoms.BoolPtr(true)},
	}
	return domain.Widget{
		ID: "w_pay_perks_marquee", TenantID: tenantID, Slug: "pay-perks-marquee",
		Name: "Abeta Pay Live Perks Ticker", TemplateKey: "tmpl_marquee_strip", SnippetType: snippets.SnippetTypeMarqueeStrip,
		Props:  structToMap(data),
		Motion: structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionFadeIn, DurationToken: atoms.MotionDurationFast, LoopEffect: atoms.MotionLoopMarquee}),
		Tags:   []string{"marquee_strip", "pay"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

func seedPayCashbackNudgeWidget(tenantID string, now time.Time) domain.Widget {
	badge := atoms.NewTag("5% CASHBACK", theming.ColorWarningContainer)
	badge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopWiggle}
	countdown := atoms.NewCountdown("pay_cashback_cd", 480, "Boost ends in")
	countdown.UrgentThresholdSeconds = 60
	countdown.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	navTransfer := actions.NewNavigateAction("nav_pay_nudge_transfer", actions.NavigatePayload{Route: "/screens/pay-transfer"})
	cta := atoms.NewButton("btn_pay_nudge_send", "Send Now", "primary", &navTransfer)
	cta.Size = "sm"
	dismissToast := actions.NewShowToastAction("dismiss_pay_nudge", actions.ShowToastPayload{
		Message: "Cashback reminder dismissed", DurationMs: 2200, Variant: "info",
	})

	data := snippets.NudgeBarSnippetData{
		NudgeID:     "pay_zero_fee_cashback",
		Title:       atoms.NewText("Zero-Fee 5% Cashback Window", theming.TypographyTitleSmall, theming.ColorTextPrimary),
		Subtitle:    ptrText(atoms.NewText("Send P2P or pay utilities before boost closes", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Icon:        ptrIcon(atoms.NewIcon("notifications", 20, theming.ColorBrandPrimary)),
		Badge:       &badge,
		Countdown:   &countdown,
		Progress:    atoms.Float64Ptr(0.72),
		CTA:         &cta,
		Dismissible: true,
		PointerEdge: snippets.PointerEdgeBottom,
		OnDismiss:   &dismissToast,
		Layout:      seedCardLayout(),
	}
	return domain.Widget{
		ID: "w_pay_cashback_nudge", TenantID: tenantID, Slug: "pay-cashback-nudge",
		Name: "5% Cashback Urgency Nudge Bar", TemplateKey: "tmpl_nudge_bar", SnippetType: snippets.SnippetTypeNudgeBar,
		Props:  structToMap(data),
		Motion: structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionSlideDown, DurationToken: atoms.MotionDurationFast}),
		Tags:   []string{"nudge_bar", "cashback", "pay"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

func seedPayRules(tenantID string, now time.Time) []domain.TargetingRule {
	return []domain.TargetingRule{
		{
			ID: "rule_pay_metal_vip", TenantID: tenantID, Key: "metal_or_vip",
			Name: "Abeta Metal & VIP Wallet Holders", Description: "Matches VIP or Metal cardholder accounts.",
			MatchMode: "all", Conditions: []domain.RuleCondition{{Field: "segment", Operator: "in", Value: []string{"vip", "metal"}}},
			Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "rule_pay_metro_clearing", TenantID: tenantID, Key: "metro_instant_clearing",
			Name: "Metro Instant Clearing Cities", Description: "Matches metro hubs with sub-2s real-time clearing.",
			MatchMode: "all", Conditions: []domain.RuleCondition{{Field: "city", Operator: "in", Value: []string{"Bengaluru", "Mumbai", "Delhi"}}},
			Priority: 80, Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedPaySlottingRules(tenantID string, now time.Time) []domain.SlottingRule {
	return []domain.SlottingRule{
		{
			ID: "slotrule_pay_metal_vip", TenantID: tenantID, LayoutSlug: "pay-home",
			Name: "Inject Abeta Metal Concierge Banner for VIP Holders", Action: domain.SlottingActionInsert,
			TargetPosition: 2, WidgetID: "w_pay_metal_vip_banner", RuleID: "rule_pay_metal_vip",
			Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedPayLayouts(tenantID string, now time.Time) []domain.Layout {
	navCards := actions.NewNavigateAction("nav_pay_cards_toolbar", actions.NavigatePayload{Route: "/screens/pay-cards"})
	cardsIcon := atoms.NewIcon("settings", 22, theming.ColorBrandPrimary)
	cardsIcon.ContentDescription = "Open Smart Card Controls"
	cardsIcon.OnClick = &navCards

	pullPay := actions.NewRefreshPageAction("pull_pay_home", actions.RefreshPagePayload{PageID: "pay-home", ShowLoader: true})

	return []domain.Layout{
		{
			ID: "lay_pay_home", TenantID: tenantID, Slug: "pay-home",
			Title: "Abeta Pay — Wallet & Smart Cards", Description: "Fintech wallet home with perks marquee, 5% cashback nudge, molded Metal card hub (extended atoms), P2P receive QR, and bill pay carousel.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Abeta Pay", "Wallet & Smart Cards · Zero-Fee P2P", "back_pay_home", cardsIcon),
				AdaptiveLayout: buildSeedCenteredAdaptive(1120),
				PullToRefresh:  &pullPay,
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_pay_search", WidgetID: "w_pay_search_bar", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_pay_marquee", WidgetID: "w_pay_perks_marquee", Position: 1, Enabled: true},
				{SlotID: "slot_pay_chips", WidgetID: "w_pay_quick_chips", Position: 2, Enabled: true},
				{SlotID: "slot_pay_nudge", WidgetID: "w_pay_cashback_nudge", Position: 3, Enabled: true},
				{SlotID: "slot_pay_metal_card", WidgetID: "w_pay_metal_card_block", Position: 4, Enabled: true},
				{SlotID: "slot_pay_receive_qr", WidgetID: "w_pay_receive_qr_block", Position: 5, Enabled: true},
				{SlotID: "slot_pay_quick_send", WidgetID: "w_pay_quick_send_row", Position: 6, Enabled: true},
				{SlotID: "slot_pay_card_controls", WidgetID: "w_pay_card_controls_row", Position: 7, Enabled: true},
				{SlotID: "slot_pay_recent_receipt", WidgetID: "w_pay_recent_receipt_row", Position: 8, Enabled: true},
				{SlotID: "slot_pay_bills", WidgetID: "w_pay_bill_carousel", Position: 9, Enabled: true},
			},
			Tags: []string{"vertical", "fintech", "payments"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_pay_transfer", TenantID: tenantID, Slug: "pay-transfer",
			Title: "Send Money — Instant P2P", Description: "Zero-fee P2P transfer with molded SegmentedControl, RadioGroup, Select, OtpInput, validated transfer form, and OS biometric authorization.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Send Money", "Abeta Pay · Zero-Fee Instant P2P", "back_pay_transfer"),
				AdaptiveLayout: buildSeedCenteredAdaptive(840),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_pay_tx_hdr", WidgetID: "w_pay_transfer_header", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_pay_tx_contacts", WidgetID: "w_pay_contact_chips", Position: 1, Enabled: true},
				{SlotID: "slot_pay_tx_controls", WidgetID: "w_pay_transfer_molded_controls", Position: 2, Enabled: true},
				{SlotID: "slot_pay_tx_form", WidgetID: "w_pay_transfer_form", Position: 3, Enabled: true},
				{SlotID: "slot_pay_tx_bio", WidgetID: "w_pay_biometric_row", Position: 4, Enabled: true},
			},
			Tags: []string{"vertical", "payments", "transfer"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_pay_cards", TenantID: tenantID, Slug: "pay-cards",
			Title: "Smart Card Controls", Description: "Real-time virtual card spend telemetry, freeze/FX/contactless switches, and monthly spend limit slider.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Smart Card Controls", "Abeta Metal · •••• 4092", "back_pay_cards"),
				AdaptiveLayout: buildSeedCenteredAdaptive(840),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_pay_cards_hdr", WidgetID: "w_pay_cards_header", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_pay_cards_metric", WidgetID: "w_pay_cards_spend_metric", Position: 1, Enabled: true},
				{SlotID: "slot_pay_cards_form", WidgetID: "w_pay_cards_controls_form", Position: 2, Enabled: true},
			},
			Tags: []string{"vertical", "payments", "cards"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_pay_receipt", TenantID: tenantID, Slug: "pay-receipt",
			Title: "Receipt #PAY-3081", Description: "Live 4-stage clearing tracker, verified KeyValue ledger with QR verification, celebration trigger, and OS receipt share.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Receipt #PAY-3081", "Abeta Pay · Instant Settlement", "back_pay_receipt"),
				AdaptiveLayout: buildSeedCenteredAdaptive(840),
				PageActions: []actions.ActionTrigger{
					actions.NewTriggerFeedbackAction("pay_receipt_celebration", actions.TriggerFeedbackPayload{
						Effect: actions.FeedbackCelebration, Title: "Transfer Sent!",
						Subtitle: "Instant settlement #PAY-3081 confirmed via Abeta Pay", DurationMs: 2400,
					}),
				},
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_pay_rcpt_tracker", WidgetID: "w_pay_settlement_tracker", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_pay_rcpt_ledger", WidgetID: "w_pay_receipt_ledger_block", Position: 1, Enabled: true},
				{SlotID: "slot_pay_rcpt_share", WidgetID: "w_pay_receipt_share_row", Position: 2, Enabled: true},
			},
			Tags: []string{"vertical", "payments", "receipt"}, CreatedAt: now, UpdatedAt: now,
		},
	}
}
