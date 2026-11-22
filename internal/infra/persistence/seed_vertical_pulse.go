package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// ── Abeta Pulse (Events, Movies & Dining Tickets) ─────────────────────────────

func buildPulseWorkspaceSpec(now time.Time) seedWorkspaceSpec {
	const tid = "tenant_abeta_pulse"
	return seedWorkspaceSpec{
		Tenant: domain.Tenant{
			ID:          tid,
			Slug:        "abeta-pulse",
			Name:        "Abeta Pulse — Events, Movies & Dining Tickets",
			Description: "Live concerts, IMAX premieres, stand-up comedy, and instant QR ticket booking",
			Active:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		TemplateIDPrefix: "pulse",
		Widgets:          seedPulseWidgets(tid, now),
		Rules: []domain.TargetingRule{
			{
				ID: "rule_pulse_weekend", TenantID: tid, Key: "weekend_spotlight",
				Name: "Weekend Event Spotlight (Fri–Sun UTC)", Description: "Promotes weekend live concerts & comedy showcases.",
				MatchMode: "all", Conditions: []domain.RuleCondition{},
				Schedule: &domain.ScheduleWindow{DaysOfWeek: []int{5, 6, 0}},
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		SlottingRules: []domain.SlottingRule{
			{
				ID: "slotrule_pulse_weekend_spotlight", TenantID: tid, LayoutSlug: "pulse-home",
				Name: "Inject Weekend Spotlight Banner", Action: domain.SlottingActionInsert,
				TargetPosition: 2, WidgetID: "w_pulse_weekend_banner", RuleID: "rule_pulse_weekend",
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		Layouts: seedPulseLayouts(tid, now),
		SplashConfigs: []domain.SplashConfig{
			{
				ID: "splash_default_pulse", TenantID: tid, Key: "default_splash",
				ImageURL: "https://cdn.abeta.dev/pulse/splash-default.png", BackgroundColorHex: "#312E81",
				DurationMs: 1400, IsDefault: true, Enabled: true, UpdatedAt: now,
			},
		},
		TokenOverrides: []domain.TokenOverride{
			{
				ID: "tok_default_pulse", TenantID: tid, Key: "default_pulse", ThemePreset: "abeta-pulse",
				Tokens: map[string]string{
					string(theming.ColorBrandPrimary):   "#4F46E5",
					string(theming.ColorBrandContainer): "#E0E7FF",
					string(theming.ColorAccentPrimary):  "#7C3AED",
				},
				Enabled: true, UpdatedAt: now,
			},
		},
		ChangeLogID:      "chg_seed_abeta_pulse",
		ChangeLogSummary: "Bootstrapped Abeta Pulse events & ticketing workspace with 2 published layouts and 21 templates",
	}
}

func seedPulseWidgets(tenantID string, now time.Time) []domain.Widget {
	navDetail := actions.NewNavigateAction("nav_pulse_spotlight", actions.NavigatePayload{Route: "/screens/pulse-event-detail"})
	spotCTA := atoms.NewButton("btn_pulse_spot", "Book Seats", "primary", &navDetail)
	spotCTA.Size = "sm"
	spotLayout := seedCardLayout()
	spotLayout.BgColor = atoms.ColorRef(theming.ColorBrandContainer)

	wkndLayout := seedCardLayout()
	wkndLayout.BgColor = atoms.ColorRef(theming.ColorWarningContainer)
	wkndCTA := atoms.NewButton("btn_pulse_wknd", "Reserve Pass", "primary", &navDetail)
	wkndCTA.Size = "sm"

	shareEvent := actions.NewShareAction("share_pulse_event", actions.SharePayload{
		Title: "Neon Horizon Live Symphony",
		Text:  "Join me at Neon Horizon Live Symphony on Abeta Pulse!",
		URL:   "https://app.abeta.dev/screens/pulse-event-detail",
	})
	shareBtn := atoms.NewButton("btn_pulse_hdr_share", "Share", "outline", &shareEvent)
	shareBtn.Size = "sm"

	bookTier := func(tierID, title, desc string) domain.Widget {
		act := actions.NewCallAPIAction("book_"+tierID, actions.CallAPIPayload{
			Endpoint: "/api/v1/pulse/bookings", Method: "POST",
			Body: map[string]any{"event_id": "neon_horizon_symphony", "tier": tierID, "quantity": 2},
		})
		btn := atoms.NewButton("btn_book_"+tierID, "Book Tickets", "primary", &act)
		btn.Size = "sm"
		return domain.Widget{
			ID: "w_pulse_" + tierID, TenantID: tenantID, Slug: "pulse-" + tierID,
			Name: title, TemplateKey: "tmpl_card_v1", SnippetType: snippets.SnippetTypeCard,
			Props: structToMap(snippets.CardSnippetData{
				Title:         atoms.NewText(title, theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Description:   ptrText(atoms.NewText(desc, theming.TypographyBodySmall, theming.ColorTextSecondary)),
				PrimaryButton: &btn,
				Layout:        seedCardLayout(),
			}),
			Tags: []string{"ticket-tier", "card"}, Active: true, CreatedAt: now, UpdatedAt: now,
		}
	}

	return []domain.Widget{
		buildSeedSearchBarWidget(
			tenantID, "w_pulse_search_bar", "pulse-search-bar", "Abeta Pulse Search Bar",
			"pulse_search", "Search concerts, movies, comedy in {{city}}...", "pulse-home", now,
		),
		buildSeedChipGroupWidget(
			tenantID, "w_pulse_category_chips", "pulse-category-chips", "Abeta Pulse Category Chips",
			"pulse_categories", "pulse-home",
			[]seedChipSpec{
				{id: "all", label: "All", selected: true},
				{id: "movies", label: "Movies"},
				{id: "concerts", label: "Concerts"},
				{id: "comedy", label: "Comedy"},
				{id: "dining", label: "Dining"},
			}, now,
		),
		{
			ID: "w_pulse_spotlight_banner", TenantID: tenantID, Slug: "pulse-spotlight-banner",
			Name: "Spotlight Concert Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("Neon Horizon Live Symphony · {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Sat 8:00 PM at Harbor Dome · Instant QR passes & acoustic floor seats.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("star", 24, theming.ColorBrandPrimary)),
				Tag:    ptrTag(atoms.NewTag("SPOTLIGHT", theming.ColorSuccessContainer)),
				CTA:    &spotCTA,
				Layout: spotLayout,
			}),
			Tracking: map[string]any{"event_name": "spotlight_viewed", "properties": map[string]any{"vertical": "abeta-pulse"}},
			Tags:     []string{"banner", "spotlight"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_pulse_weekend_banner", TenantID: tenantID, Slug: "pulse-weekend-banner",
			Name: "Weekend Priority Pass Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("Weekend Lineup Live in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("Limited Friday–Sunday releases across concerts, IMAX & comedy clubs.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("bolt", 24, theming.ColorWarning)),
				Tag:    ptrTag(atoms.NewTag("WEEKEND", theming.ColorSurfacePrimary)),
				CTA:    &wkndCTA,
				Layout: wkndLayout,
			}),
			Tags: []string{"banner", "weekend", "slotting"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_pulse_ev_symphony", slug: "pulse-ev-symphony", name: "Neon Horizon Concert Card",
			title: "Neon Horizon Live Symphony", subtitle: "Concerts · Sat, 8:00 PM · Harbor Dome",
			details: "From $30 · Instant QR · Selling fast", badge: "CONCERTS",
			badgeColor: theming.ColorBrandContainer, ctaLabel: "Book Tickets", route: "pulse-event-detail",
			tags: []string{"event", "concerts"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_pulse_ev_cinema", slug: "pulse-ev-cinema", name: "Orbital Drift IMAX Card",
			title: "Orbital Drift — IMAX 70mm", subtitle: "Movies · Today, 7:15 PM · Starlight",
			details: "From $16 · Dolby Atmos · Plush recliners", badge: "IMAX",
			badgeColor: theming.ColorInfoContainer, ctaLabel: "Select Seats", route: "pulse-event-detail",
			tags: []string{"event", "movies"},
		}),
		buildSeedListingCardWidget(tenantID, now, seedListingCardSpec{
			widgetID: "w_pulse_ev_comedy", slug: "pulse-ev-comedy", name: "Stand-Up Showcase Card",
			title: "Late-Night Stand-Up Showcase", subtitle: "Comedy & Dining · Fri, 9:00 PM · Lantern",
			details: "From $22 · Includes 1 craft beverage", badge: "COMEDY",
			badgeColor: theming.ColorWarningContainer, ctaLabel: "Reserve Table", route: "pulse-event-detail",
			tags: []string{"event", "comedy"},
		}),
		{
			ID: "w_pulse_detail_header", TenantID: tenantID, Slug: "pulse-detail-header",
			Name: "Event Detail Header", TemplateKey: "tmpl_header_v1", SnippetType: snippets.SnippetTypeHeader,
			Props: structToMap(snippets.HeaderSnippetData{
				Title:          atoms.NewText("Neon Horizon Live Symphony", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:       ptrText(atoms.NewText("Sat 8:00 PM · Harbor Dome, {{city}}", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag:    ptrTag(atoms.NewTag("SELLING FAST", theming.ColorWarningContainer)),
				TrailingAction: &shareBtn,
				Layout:         seedCardLayout(),
			}),
			Tags: []string{"header", "event"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		bookTier("tier_vip", "Front Row VIP · $95", "Lounge access, signed poster & pre-show soundcheck"),
		bookTier("tier_gold", "Arena Gold · $55", "Center acoustic sweet spot + fast-track entry"),
		bookTier("tier_balcony", "Balcony Acoustic · $30", "Panoramic stage view & instant QR pass"),
		{
			ID: "w_pulse_share_row", TenantID: tenantID, Slug: "pulse-share-row",
			Name: "Share Event Row", TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:        atoms.NewText("Share Event with Friends", theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:     ptrText(atoms.NewText("Send Harbor Dome seating & schedule link", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon:  ptrIcon(atoms.NewIcon("info", 20, theming.ColorBrandPrimary)),
				TrailingTag:  ptrTag(atoms.NewTag("SHARE", theming.ColorBrandContainer)),
				TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 18, theming.ColorTextTertiary)),
				OnClick:      &shareEvent,
				Layout:       seedCardLayout(),
			}),
			Tags: []string{"action_row", "share"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedPulseLayouts(tenantID string, now time.Time) []domain.Layout {
	bookGold := actions.NewCallAPIAction("pulse_book_gold", actions.CallAPIPayload{
		Endpoint: "/api/v1/pulse/bookings", Method: "POST",
		Body: map[string]any{"event_id": "neon_horizon_symphony", "tier": "arena_gold", "quantity": 2},
	})
	bookBtn := atoms.NewButton("btn_pulse_bottom_book", "Book 2 Gold · $110", "primary", &bookGold)
	bookBtn.Layout = &atoms.LayoutConfig{FillMaxWidth: atoms.BoolPtr(true)}
	pullPulse := actions.NewRefreshPageAction("pull_pulse_home", actions.RefreshPagePayload{PageID: "pulse-home", ShowLoader: true})

	return []domain.Layout{
		{
			ID: "lay_pulse_home", TenantID: tenantID, Slug: "pulse-home",
			Title: "Abeta Pulse — Live Events", Description: "Events, movies, comedy, and dining ticketing hub.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Abeta Pulse", "{{city}} · Events, Movies & Dining", "back_pulse_home"),
				AdaptiveLayout: buildSeedCenteredAdaptive(1120),
				PullToRefresh:  &pullPulse,
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_pulse_search", WidgetID: "w_pulse_search_bar", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_pulse_chips", WidgetID: "w_pulse_category_chips", Position: 1, Enabled: true},
				{SlotID: "slot_pulse_spotlight", WidgetID: "w_pulse_spotlight_banner", Position: 2, Enabled: true},
				{SlotID: "slot_pulse_symphony", WidgetID: "w_pulse_ev_symphony", Position: 3, Enabled: true},
				{SlotID: "slot_pulse_cinema", WidgetID: "w_pulse_ev_cinema", Position: 4, Enabled: true},
				{SlotID: "slot_pulse_comedy", WidgetID: "w_pulse_ev_comedy", Position: 5, Enabled: true},
			},
			Tags: []string{"vertical", "events"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_pulse_event_detail", TenantID: tenantID, Slug: "pulse-event-detail",
			Title: "Neon Horizon Live Symphony", Description: "Ticket tier selection, venue sharing, and instant QR booking.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar: buildSeedVerticalToolbar("Event Tickets", "Abeta Pulse · Instant QR Pass", "back_pulse_detail"),
				BottomBar: &snippets.BottomBarSpec{
					PrimaryButton: &bookBtn,
					HelperText:    ptrText(atoms.NewText("Instant QR passes · Refundable up to 24h", theming.TypographyLabelSmall, theming.ColorTextSecondary)),
				},
				AdaptiveLayout: buildSeedCenteredAdaptive(920),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_pulse_hdr", WidgetID: "w_pulse_detail_header", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_pulse_vip", WidgetID: "w_pulse_tier_vip", Position: 1, Enabled: true},
				{SlotID: "slot_pulse_gold", WidgetID: "w_pulse_tier_gold", Position: 2, Enabled: true},
				{SlotID: "slot_pulse_balcony", WidgetID: "w_pulse_tier_balcony", Position: 3, Enabled: true},
				{SlotID: "slot_pulse_share", WidgetID: "w_pulse_share_row", Position: 4, Enabled: true},
			},
			Tags: []string{"vertical", "event-detail"}, CreatedAt: now, UpdatedAt: now,
		},
	}
}
