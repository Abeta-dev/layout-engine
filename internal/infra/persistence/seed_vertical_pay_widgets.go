package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// seedPayMetalCardComposedWidget molds an Abeta Metal hub card using the v0.5.0
// extended atoms: AlertBanner, StatTile, Badge, Chip, Rating, and Price.
func seedPayMetalCardComposedWidget(tenantID string, now time.Time) domain.Widget {
	navSend := actions.NewNavigateAction("nav_metal_send", actions.NavigatePayload{Route: "/screens/pay-transfer"})
	sendBtn := atoms.NewButton("btn_metal_send", "Send Money", "primary", &navSend)
	sendBtn.Size = "sm"

	navCards := actions.NewNavigateAction("nav_metal_controls", actions.NavigatePayload{Route: "/screens/pay-cards"})
	controlsBtn := atoms.NewButton("btn_metal_controls", "Card Controls", "outline", &navCards)
	controlsBtn.Size = "sm"

	alert := atoms.NewAlertBanner("pay_alert_clearing", "Instant Real-Time Clearing Active in {{city}}", "0% transfer fee on all P2P handles & utilities.", atoms.VariantSuccess)
	alert.Icon = ptrIcon(atoms.NewIcon("verified", 18, theming.ColorSuccess))

	balStat := atoms.NewStatTile("pay_stat_bal", "Available Wallet", "$2,480.50")
	balStat.DeltaLabel = "+5.0% APY"
	balStat.DeltaPositive = atoms.BoolPtr(true)
	balStat.Icon = ptrIcon(atoms.NewIcon("star", 18, theming.ColorBrandPrimary))

	cbStat := atoms.NewStatTile("pay_stat_cb", "Monthly Cashback", "$124.80")
	cbStat.DeltaLabel = "+18% MoM"
	cbStat.DeltaPositive = atoms.BoolPtr(true)
	cbStat.Icon = ptrIcon(atoms.NewIcon("bolt", 18, theming.ColorSuccess))

	perkBadge := atoms.NewBadge("pay_perks_badge", 3, 99, atoms.VariantBrand)
	perkBadge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	chipZeroFX := atoms.NewChip("chip_zero_fx", "0% FX Markup", true, nil)
	chipZeroFX.LeadingIcon = ptrIcon(atoms.NewIcon("verified", 14, theming.ColorBrandPrimary))
	chipNfc := atoms.NewChip("chip_nfc_tap", "Tap-to-Pay NFC", true, nil)
	chipNfc.LeadingIcon = ptrIcon(atoms.NewIcon("bolt", 14, theming.ColorSuccess))

	trustRating := atoms.NewRating("pay_trust_rating", 4.9, 5)
	trustRating.ReviewCount = atoms.IntPtr(4120)
	trustRating.Label = "Member Rating"

	feePrice := atoms.NewPrice("pay_fee_price", "$0.00", "$2.50")
	feePrice.DiscountLabel = "ZERO FEE"
	feePrice.UnitSuffix = "/ transfer"

	fullWidth := atoms.LayoutConfig{FillMaxWidth: atoms.BoolPtr(true)}
	root := snippets.CardNode(
		*seedCardLayout(),
		snippets.VStack(
			theming.SpacingSM.Dp(),
			snippets.HStack(
				theming.SpacingSM.Dp(),
				snippets.TagNode(atoms.NewTag("ABETA METAL · •••• 4092", theming.ColorBrandContainer)),
				snippets.BadgeNode(perkBadge),
			).WithArrangement(snippets.ComposedAlignSpaceBetween).WithAlignment("", snippets.ComposedAlignCenter).WithLayout(fullWidth),
			snippets.AlertBannerNode(alert),
			snippets.HStack(
				theming.SpacingSM.Dp(),
				snippets.StatTileNode(balStat).WithWeight(1),
				snippets.StatTileNode(cbStat).WithWeight(1),
			).WithLayout(fullWidth),
			snippets.FlowRowNode(
				theming.SpacingXS.Dp(),
				snippets.ChipNode(chipZeroFX),
				snippets.ChipNode(chipNfc),
			),
			snippets.HStack(
				theming.SpacingSM.Dp(),
				snippets.RatingNode(trustRating),
				snippets.PriceNode(feePrice),
			).WithArrangement(snippets.ComposedAlignSpaceBetween).WithAlignment("", snippets.ComposedAlignCenter).WithLayout(fullWidth),
			snippets.DividerNode(atoms.DividerAtom{ThicknessDp: atoms.Float64Ptr(1), Color: atoms.ColorRef(theming.ColorDivider)}),
			snippets.HStack(
				theming.SpacingSM.Dp(),
				snippets.ButtonNode(sendBtn),
				snippets.ButtonNode(controlsBtn),
			).WithArrangement(snippets.ComposedAlignSpaceBetween).WithLayout(fullWidth),
		).WithLayout(fullWidth),
	).WithNodeID("pay_metal_card_root").WithMotion(atoms.MotionSpec{LoopEffect: atoms.MotionLoopGlowRing})

	data := snippets.ComposedBlockSnippetData{
		BlockID: "pay_metal_virtual_card", RecipeName: "metal_virtual_card_hub", Root: root,
	}
	return domain.Widget{
		ID: "w_pay_metal_card_block", TenantID: tenantID, Slug: "pay-metal-card-block",
		Name: "Abeta Metal Hub (Molded Extended Atoms)", TemplateKey: "tmpl_composed_block", SnippetType: snippets.SnippetTypeComposedBlock,
		Props:    structToMap(data),
		Motion:   structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionSlideUp, DurationToken: atoms.MotionDurationNormal}),
		Tracking: map[string]any{"event_name": "composed_block_viewed", "properties": map[string]any{"recipe": "metal_virtual_card_hub", "vertical": "abeta-pay"}},
		Tags:     []string{"composed_block", "metal", "extended_atoms"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

// seedPayReceiveQrComposedWidget molds a scannable P2P QR Receive card using QrCode + KeyValueRow.
func seedPayReceiveQrComposedWidget(tenantID string, now time.Time) domain.Widget {
	qr := atoms.NewQrCode("pay_receive_qr", "abeta://pay/@alex.rivera?city={{city}}", "@alex.rivera", 156)
	qr.Caption = "Scan with any Abeta Pay app for instant 0% fee settlement"

	kvHandle := atoms.NewKeyValueRow("kv_pay_handle", "Pay Handle", "@alex.rivera")
	kvHandle.ShowDivider = true
	kvRail := atoms.NewKeyValueRow("kv_pay_rail", "Settlement Rail", "Instant (< 2s)")
	kvRail.ValueColor = atoms.ColorRef(theming.ColorSuccess)
	kvRail.Emphasized = true

	fullWidth := atoms.LayoutConfig{FillMaxWidth: atoms.BoolPtr(true)}
	root := snippets.CardNode(
		*seedCardLayout(),
		snippets.VStack(
			theming.SpacingSM.Dp(),
			snippets.TextNode(atoms.NewText("Receive Instant Payments", theming.TypographyTitleSmall, theming.ColorTextPrimary)),
			snippets.BoxNode(snippets.QrCodeNode(qr)).WithContentAlignment(snippets.ComposedAlignCenter).WithLayout(fullWidth),
			snippets.KeyValueRowNode(kvHandle),
			snippets.KeyValueRowNode(kvRail),
		).WithLayout(fullWidth),
	).WithNodeID("pay_receive_qr_root")

	return domain.Widget{
		ID: "w_pay_receive_qr_block", TenantID: tenantID, Slug: "pay-receive-qr-block",
		Name: "P2P Receive QR Code Card", TemplateKey: "tmpl_composed_block", SnippetType: snippets.SnippetTypeComposedBlock,
		Props:    structToMap(snippets.ComposedBlockSnippetData{BlockID: "pay_receive_qr_card", RecipeName: "pay_receive_qr_card", Root: root}),
		Tracking: map[string]any{"event_name": "composed_block_viewed", "properties": map[string]any{"recipe": "pay_receive_qr_card", "vertical": "abeta-pay"}},
		Tags:     []string{"composed_block", "qr_code", "extended_atoms"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

// seedPayTransferMoldedControlsWidget molds transfer selection controls using
// SegmentedControl, RadioGroup, Select, OtpInput, and Price.
func seedPayTransferMoldedControlsWidget(tenantID string, now time.Time) domain.Widget {
	seg := atoms.NewSegmentedControl(
		"pay_transfer_mode", "instant",
		atoms.OptionItem{ID: "instant", Label: "Instant (<2s)"},
		atoms.OptionItem{ID: "scheduled", Label: "Scheduled"},
		atoms.OptionItem{ID: "split", Label: "Split Bill"},
	)

	rg := atoms.NewRadioGroup(
		"pay_funding_source", "Funding Source", "metal_4092",
		atoms.OptionItem{ID: "metal_4092", Label: "Abeta Metal •••• 4092", Subtitle: "5% instant cashback eligible", Badge: "5% BACK"},
		atoms.OptionItem{ID: "vault_usd", Label: "USD Cash Vault", Subtitle: "Available: $2,480.50 · Zero fee"},
		atoms.OptionItem{ID: "checking_8821", Label: "Linked Checking •••• 8821", Subtitle: "Instant ACH clearing"},
	)

	sel := atoms.NewSelect(
		"pay_currency", "Settlement Currency", "Choose currency", "usd",
		atoms.OptionItem{ID: "usd", Label: "USD — US Dollar ($)"},
		atoms.OptionItem{ID: "eur", Label: "EUR — Euro (€)", Badge: "0% FX"},
		atoms.OptionItem{ID: "gbp", Label: "GBP — British Pound (£)", Badge: "0% FX"},
		atoms.OptionItem{ID: "inr", Label: "INR — Indian Rupee (₹)", Badge: "0% FX"},
	)
	sel.HelperText = "Guaranteed mid-market rate with zero FX markup"

	pinComplete := actions.NewShowToastAction("toast_pay_pin_verified", actions.ShowToastPayload{
		Message: "6-digit transfer PIN verified!", DurationMs: 1800, Variant: "success",
	})
	otp := atoms.NewOtpInput("pay_security_pin", "6-Digit Transfer PIN", 6, &pinComplete)
	otp.Masked = true
	otp.HelperText = "Required for instant transfers over $100"

	summaryPrice := atoms.NewPrice("pay_tx_summary_price", "$125.00", "$127.50")
	summaryPrice.DiscountLabel = "ZERO FEE"

	fullWidth := atoms.LayoutConfig{FillMaxWidth: atoms.BoolPtr(true)}
	root := snippets.CardNode(
		*seedCardLayout(),
		snippets.VStack(
			theming.SpacingMD.Dp(),
			snippets.SegmentedControlNode(seg),
			snippets.RadioGroupNode(rg),
			snippets.SelectNode(sel),
			snippets.OtpInputNode(otp),
			snippets.DividerNode(atoms.DividerAtom{ThicknessDp: atoms.Float64Ptr(1), Color: atoms.ColorRef(theming.ColorDivider)}),
			snippets.HStack(
				theming.SpacingSM.Dp(),
				snippets.TextNode(atoms.NewText("Total Debited", theming.TypographyLabelLarge, theming.ColorTextSecondary)),
				snippets.PriceNode(summaryPrice),
			).WithArrangement(snippets.ComposedAlignSpaceBetween).WithAlignment("", snippets.ComposedAlignCenter).WithLayout(fullWidth),
		).WithLayout(fullWidth),
	).WithNodeID("pay_transfer_controls_root")

	return domain.Widget{
		ID: "w_pay_transfer_molded_controls", TenantID: tenantID, Slug: "pay-transfer-molded-controls",
		Name: "P2P Transfer Controls (Segmented, Radio, Select, OTP)", TemplateKey: "tmpl_composed_block", SnippetType: snippets.SnippetTypeComposedBlock,
		Props:    structToMap(snippets.ComposedBlockSnippetData{BlockID: "pay_transfer_controls", RecipeName: "p2p_transfer_controls", Root: root}),
		Tracking: map[string]any{"event_name": "composed_block_viewed", "properties": map[string]any{"recipe": "p2p_transfer_controls", "vertical": "abeta-pay"}},
		Tags:     []string{"composed_block", "transfer", "extended_atoms"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

func seedPayTransferFormWidget(tenantID string, now time.Time) domain.Widget {
	submitAct := actions.NewCallAPIAction("pay_execute_transfer", actions.CallAPIPayload{
		Endpoint: "/api/v1/pay/transfer",
		Method:   "POST",
		Body:     map[string]any{"vertical": "abeta-pay", "recipient": "@maya.chen", "amount_cents": 12500, "currency": "USD"},
	}).WithValidations(
		actions.ValidationRule{FieldID: "pay_recipient", Rule: "required", Message: "Recipient handle or email is required"},
		actions.ValidationRule{FieldID: "pay_recipient", Rule: "regex", Value: payRecipientRegex, Message: "Enter a valid @handle or email"},
		actions.ValidationRule{FieldID: "pay_amount", Rule: "required", Message: "Transfer amount is required"},
		actions.ValidationRule{FieldID: "pay_amount", Rule: "regex", Value: payAmountRegex, Message: "Enter a valid USD amount (e.g. 125.00)"},
	)
	submitBtn := atoms.NewButton("btn_pay_send_now", "Send $125.00 Instantly", "primary", &submitAct)

	data := snippets.FormSnippetData{
		FormID:      "pay_transfer_form",
		Title:       ptrText(atoms.NewText("Transfer Details", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
		Description: ptrText(atoms.NewText("Validated client & server before instant settlement.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Fields: []atoms.TextFieldAtom{
			{FieldID: "pay_recipient", Label: "Recipient (@handle or email)", Placeholder: "@maya.chen", InitialValue: "@maya.chen", HelperText: "Instant Abeta Pay handle or verified email", Required: true, Regex: payRecipientRegex, KeyboardType: "text"},
			{FieldID: "pay_amount", Label: "Amount (USD)", Placeholder: "125.00", InitialValue: "125.00", HelperText: "Zero fee up to $2,500 per instant transfer", Required: true, Regex: payAmountRegex, KeyboardType: "number"},
			{FieldID: "pay_note", Label: "Memo (Optional)", Placeholder: "Dinner split & coffee", InitialValue: "Dinner split", KeyboardType: "text"},
		},
		SubmitButton: submitBtn,
		Layout:       seedCardLayout(),
	}
	return domain.Widget{
		ID: "w_pay_transfer_form", TenantID: tenantID, Slug: "pay-transfer-form",
		Name: "Instant P2P Transfer Form", TemplateKey: "tmpl_form_v1", SnippetType: snippets.SnippetTypeForm,
		Props: structToMap(data), Tags: []string{"form", "transfer"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

func seedPayBiometricRowWidget(tenantID string, now time.Time) domain.Widget {
	onAuth := actions.NewCallAPIAction("pay_bio_transfer", actions.CallAPIPayload{
		Endpoint: "/api/v1/pay/transfer", Method: "POST",
		Body: map[string]any{"vertical": "abeta-pay", "recipient": "@maya.chen", "amount_cents": 12500},
	})
	onFailed := actions.NewShowToastAction("pay_biometric_fallback_toast", actions.ShowToastPayload{
		Message: "Biometric unavailable on this host — use 6-digit PIN above", DurationMs: 2200, Variant: "warning",
	})
	bioAct := actions.NewBiometricAuthAction("pay_biometric_authorize", actions.BiometricAuthPayload{
		Reason: "Authorize instant $125.00 P2P transfer to @maya.chen", Title: "Abeta Pay Biometric Auth",
		AllowDeviceCredential: atoms.BoolPtr(true), OnAuthenticated: &onAuth, OnFailed: &onFailed,
	})
	return domain.Widget{
		ID: "w_pay_biometric_row", TenantID: tenantID, Slug: "pay-biometric-row",
		Name: "Biometric Transfer Authorization Row", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
		Props: structToMap(snippets.ActionRowSnippetData{
			Title:        atoms.NewText("Authorize with Face ID / Touch ID", theming.TypographyTitleSmall, theming.ColorTextPrimary),
			Subtitle:     ptrText(atoms.NewText("Confirm $125.00 instant transfer via OS biometric prompt", theming.TypographyBodySmall, theming.ColorTextSecondary)),
			LeadingIcon:  ptrIcon(atoms.NewIcon("verified", 20, theming.ColorBrandPrimary)),
			TrailingTag:  ptrTag(atoms.NewTag("BIOMETRIC", theming.ColorBrandContainer)),
			TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 18, theming.ColorTextTertiary)),
			OnClick:      &bioAct,
			Layout:       seedCardLayout(),
		}),
		Tags: []string{"action_row", "biometric"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

func seedPayCardsFormWidget(tenantID string, now time.Time) domain.Widget {
	saveFreeze := actions.NewSaveStateAction("save_card_freeze", actions.SaveStatePayload{Key: "card_freeze", Value: false, Scope: "session"})
	saveIntl := actions.NewSaveStateAction("save_card_intl", actions.SaveStatePayload{Key: "card_international", Value: true, Scope: "session"})
	saveNfc := actions.NewSaveStateAction("save_card_nfc", actions.SaveStatePayload{Key: "card_contactless", Value: true, Scope: "session"})
	saveLimit := actions.NewSaveStateAction("save_card_limit", actions.SaveStatePayload{Key: "card_spend_limit", Value: 5000, Scope: "session"})
	saveAll := actions.NewSaveStateAction("save_pay_card_controls", actions.SaveStatePayload{
		Key: "pay_card_controls_saved", Value: true, Scope: "session",
	}).WithSecondaryActions(actions.NewShowToastAction("toast_pay_cards_saved", actions.ShowToastPayload{
		Message: "Smart Card Controls saved & synced!", DurationMs: 2200, Variant: "success",
	}))

	freezeSw := atoms.NewSwitch("card_freeze", "Freeze Card Temporarily", false, &saveFreeze)
	freezeSw.Subtitle = "Instantly block all new card transactions & subscriptions"
	intlSw := atoms.NewSwitch("card_international", "International & FX Spend", true, &saveIntl)
	intlSw.Subtitle = "Zero FX markup across 140+ global currencies"
	nfcSw := atoms.NewSwitch("card_contactless", "Contactless & Tap-to-Pay", true, &saveNfc)
	nfcSw.Subtitle = "Enable NFC tap-to-pay on phones & wearables"
	limitSl := atoms.NewSlider("card_spend_limit", "Monthly Spend Limit", 5000, 500, 15000, &saveLimit)
	limitSl.Steps = 29
	limitSl.UnitLabel = "USD"

	data := snippets.FormSnippetData{
		FormID:       "pay_card_controls_form",
		Title:        ptrText(atoms.NewText("Card Security & Spend Limits", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
		Description:  ptrText(atoms.NewText("Real-time switches & monthly spend cap for •••• 4092.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		Fields:       []atoms.TextFieldAtom{},
		Switches:     []atoms.SwitchAtom{freezeSw, intlSw, nfcSw},
		Sliders:      []atoms.SliderAtom{limitSl},
		SubmitButton: atoms.NewButton("btn_save_pay_cards", "Save Card Controls", "primary", &saveAll),
		Layout:       seedCardLayout(),
	}
	return domain.Widget{
		ID: "w_pay_cards_controls_form", TenantID: tenantID, Slug: "pay-cards-controls-form",
		Name: "Smart Card Security Controls Form", TemplateKey: "tmpl_form_v1", SnippetType: snippets.SnippetTypeForm,
		Props: structToMap(data), Tags: []string{"form", "cards"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

func seedPaySettlementTrackerWidget(tenantID string, shareAct actions.ActionTrigger, now time.Time) domain.Widget {
	etaCountdown := atoms.NewCountdown("pay_settlement_cd", 12, "Ledger sync")
	etaCountdown.UrgentThresholdSeconds = 5
	etaCountdown.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}
	etaBadge := atoms.NewTag("INSTANT", theming.ColorSuccessContainer)
	etaBadge.Motion = &atoms.MotionSpec{LoopEffect: atoms.MotionLoopPulse}

	primaryBtn := atoms.NewButton("btn_pay_share_receipt", "Share Receipt", "primary", &shareAct)
	primaryBtn.Size = "sm"
	navHome := actions.NewNavigateAction("nav_pay_receipt_home", actions.NavigatePayload{Route: "/screens/pay-home"})
	secondaryBtn := atoms.NewButton("btn_pay_back_wallet", "Back to Wallet", "outline", &navHome)
	secondaryBtn.Size = "sm"

	data := snippets.OrderTrackerSnippetData{
		TrackerID: "pay_settlement_3081", OrderCode: "#PAY-3081",
		Headline:     atoms.NewText("$125.00 sent to @maya.chen · Instant Settlement", theming.TypographyTitleMedium, theming.ColorTextPrimary),
		EtaCountdown: &etaCountdown, EtaBadge: &etaBadge,
		RiderName:    "Abeta Real-Time Clearing",
		RiderCaption: "Ref #PAY-3081 · Zero fee · +$6.25 (5%) cashback credited",
		Steps: []snippets.OrderTrackerStep{
			{ID: "step_initiated", Label: "Initiated", Caption: "09:42:01", Icon: ptrIcon(atoms.NewIcon("check_circle", 18, theming.ColorSuccess)), Completed: true},
			{ID: "step_verified", Label: "Biometric Verified", Caption: "09:42:02", Icon: ptrIcon(atoms.NewIcon("check_circle", 18, theming.ColorSuccess)), Completed: true},
			{ID: "step_clearing", Label: "Instant Clearing", Caption: "09:42:03", Icon: ptrIcon(atoms.NewIcon("check_circle", 18, theming.ColorSuccess)), Completed: true},
			{ID: "step_settled", Label: "Settled", Caption: "Confirmed", Icon: ptrIcon(atoms.NewIcon("star", 18, theming.ColorBrandPrimary)), Active: true},
		},
		PrimaryAction: &primaryBtn, SecondaryAction: &secondaryBtn, Layout: seedCardLayout(),
	}
	return domain.Widget{
		ID: "w_pay_settlement_tracker", TenantID: tenantID, Slug: "pay-settlement-tracker",
		Name: "4-Stage Live Settlement Tracker #PAY-3081", TemplateKey: "tmpl_order_tracker", SnippetType: snippets.SnippetTypeOrderTracker,
		Props:    structToMap(data),
		Motion:   structToMap(atoms.MotionSpec{Enter: atoms.MotionTransitionSlideUp, DurationToken: atoms.MotionDurationNormal}),
		Tracking: map[string]any{"event_name": "order_tracker_viewed", "properties": map[string]any{"order_code": "#PAY-3081", "vertical": "abeta-pay"}},
		Tags:     []string{"order_tracker", "receipt", "pay"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}

func seedPayReceiptLedgerComposedWidget(tenantID string, now time.Time) domain.Widget {
	alert := atoms.NewAlertBanner("rcpt_alert_ok", "Settlement Confirmed", "Delivered to @maya.chen in 1.4 seconds via Abeta Instant Rail.", atoms.VariantSuccess)
	alert.Icon = ptrIcon(atoms.NewIcon("check_circle", 18, theming.ColorSuccess))

	kv1 := atoms.NewKeyValueRow("kv_rcpt_to", "Recipient", "@maya.chen")
	kv1.ShowDivider = true
	kv2 := atoms.NewKeyValueRow("kv_rcpt_amt", "Amount Sent", "$125.00 USD")
	kv2.Emphasized = true
	kv2.ShowDivider = true
	kv3 := atoms.NewKeyValueRow("kv_rcpt_fee", "Network Fee", "$0.00")
	kv3.ValueColor = atoms.ColorRef(theming.ColorSuccess)
	kv3.ShowDivider = true
	kv4 := atoms.NewKeyValueRow("kv_rcpt_cb", "5% Metal Cashback", "+$6.25")
	kv4.ValueColor = atoms.ColorRef(theming.ColorBrandPrimary)
	kv4.Emphasized = true

	qr := atoms.NewQrCode("rcpt_verify_qr", "https://abeta.dev/pay/receipt/PAY-3081", "Ref #PAY-3081", 120)
	qr.Caption = "Scan to verify signed clearing receipt"

	fullWidth := atoms.LayoutConfig{FillMaxWidth: atoms.BoolPtr(true)}
	root := snippets.CardNode(
		*seedCardLayout(),
		snippets.VStack(
			theming.SpacingSM.Dp(),
			snippets.AlertBannerNode(alert),
			snippets.KeyValueRowNode(kv1),
			snippets.KeyValueRowNode(kv2),
			snippets.KeyValueRowNode(kv3),
			snippets.KeyValueRowNode(kv4),
			snippets.BoxNode(snippets.QrCodeNode(qr)).WithContentAlignment(snippets.ComposedAlignCenter).WithLayout(fullWidth),
		).WithLayout(fullWidth),
	).WithNodeID("pay_receipt_ledger_root")

	return domain.Widget{
		ID: "w_pay_receipt_ledger_block", TenantID: tenantID, Slug: "pay-receipt-ledger-block",
		Name: "Settlement Receipt Ledger (KeyValue + Alert + QR)", TemplateKey: "tmpl_composed_block", SnippetType: snippets.SnippetTypeComposedBlock,
		Props:    structToMap(snippets.ComposedBlockSnippetData{BlockID: "pay_receipt_ledger", RecipeName: "settlement_receipt_ledger", Root: root}),
		Tracking: map[string]any{"event_name": "composed_block_viewed", "properties": map[string]any{"recipe": "settlement_receipt_ledger", "vertical": "abeta-pay"}},
		Tags:     []string{"composed_block", "receipt", "extended_atoms"}, Active: true, CreatedAt: now, UpdatedAt: now,
	}
}
