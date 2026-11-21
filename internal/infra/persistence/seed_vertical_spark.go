package persistence

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// ── Abeta Spark (Swipe Discovery & Matches) ───────────────────────────────────

func buildSparkWorkspaceSpec(now time.Time) seedWorkspaceSpec {
	const tid = "tenant_abeta_spark"
	boostStart, boostEnd := 13, 18
	return seedWorkspaceSpec{
		Tenant: domain.Tenant{
			ID:          tid,
			Slug:        "abeta-spark",
			Name:        "Abeta Spark — Swipe Discovery & Matches",
			Description: "Curated interactive swipe discovery deck, instant mutual sparks, and match threads",
			Active:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		TemplateIDPrefix: "spark",
		Widgets:          seedSparkWidgets(tid, now),
		Rules: []domain.TargetingRule{
			{
				ID: "rule_spark_evening_boost", TenantID: tid, Key: "evening_boost_window",
				Name: "Evening Discovery Boost (13:00–18:00 UTC)", Description: "Highlights active evening profiles nearby.",
				MatchMode: "all", Conditions: []domain.RuleCondition{},
				Schedule: &domain.ScheduleWindow{StartHourUTC: &boostStart, EndHourUTC: &boostEnd},
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		SlottingRules: []domain.SlottingRule{
			{
				ID: "slotrule_spark_evening_boost", TenantID: tid, LayoutSlug: "spark-deck",
				Name: "Inject Evening Boost Banner", Action: domain.SlottingActionInsert,
				TargetPosition: 2, WidgetID: "w_spark_boost_banner", RuleID: "rule_spark_evening_boost",
				Priority: 100, Enabled: true, CreatedAt: now, UpdatedAt: now,
			},
		},
		Layouts: seedSparkLayouts(tid, now),
		SplashConfigs: []domain.SplashConfig{
			{
				ID: "splash_default_spark", TenantID: tid, Key: "default_splash",
				ImageURL: "https://cdn.abeta.dev/spark/splash-default.png", BackgroundColorHex: "#4338CA",
				DurationMs: 1300, IsDefault: true, Enabled: true, UpdatedAt: now,
			},
		},
		TokenOverrides: []domain.TokenOverride{
			{
				ID: "tok_default_spark", TenantID: tid, Key: "default_spark", ThemePreset: "abeta-spark",
				Tokens: map[string]string{
					string(theming.ColorBrandPrimary):   "#6366F1",
					string(theming.ColorBrandContainer): "#EEF2FF",
					string(theming.ColorAccentPrimary):  "#EC4899",
				},
				Enabled: true, UpdatedAt: now,
			},
		},
		ChangeLogID:      "chg_seed_abeta_spark",
		ChangeLogSummary: "Bootstrapped Abeta Spark swipe discovery workspace with 2 published layouts and 17 templates",
	}
}

func sparkSwipeAction(profileID, decision string) actions.ActionTrigger {
	return actions.NewCallAPIAction("swipe_"+profileID+"_"+decision, actions.CallAPIPayload{
		Endpoint: "/api/v1/spark/swipe",
		Method:   "POST",
		Body:     map[string]any{"profile_id": profileID, "action": decision},
	})
}

func buildSparkCard(id, nameAge, role, bio, dist, badge string, badgeColor theming.ColorToken, interests []string) snippets.SwipeCardItem {
	tags := make([]atoms.TagAtom, 0, len(interests))
	for _, label := range interests {
		tags = append(tags, atoms.NewTag(label, theming.ColorSurfaceSecondary))
	}
	onLike := sparkSwipeAction(id, "like")
	onNope := sparkSwipeAction(id, "nope")
	onSuper := sparkSwipeAction(id, "super_like")
	return snippets.SwipeCardItem{
		CardID:       id,
		Title:        atoms.NewText(nameAge, theming.TypographyTitleLarge, theming.ColorTextPrimary),
		Subtitle:     ptrText(atoms.NewText(role, theming.TypographyLabelLarge, theming.ColorTextBrand)),
		Bio:          ptrText(atoms.NewText(bio, theming.TypographyBodyMedium, theming.ColorTextSecondary)),
		DistanceText: ptrText(atoms.NewText(dist, theming.TypographyLabelMedium, theming.ColorTextSecondary)),
		Badge:        ptrTag(atoms.NewTag(badge, badgeColor)),
		Image: &atoms.ImageAtom{
			URL: "https://cdn.abeta.dev/spark/" + id + ".png", ContentDescription: nameAge,
			HeightDp: atoms.IntPtr(220), ScaleMode: "crop", CornerRadius: atoms.UniformRadius(theming.RadiusMD.Dp()),
		},
		Tags:        tags,
		OnLike:      &onLike,
		OnNope:      &onNope,
		OnSuperLike: &onSuper,
	}
}

func seedSparkWidgets(tenantID string, now time.Time) []domain.Widget {
	navMatches := actions.NewNavigateAction("nav_spark_matches", actions.NavigatePayload{Route: "/screens/spark-matches"})
	matchesBtn := atoms.NewButton("btn_spark_matches", "Matches (3)", "outline", &navMatches)
	matchesBtn.Size = "sm"

	deckLike := sparkSwipeAction("deck", "like")
	deckNope := sparkSwipeAction("deck", "nope")
	deckSuper := sparkSwipeAction("deck", "super_like")
	deckUndo := actions.NewShowToastAction("spark_undo_card", actions.ShowToastPayload{
		Message: "Rewound previous discovery card", DurationMs: 1800, Variant: "info",
	})

	deckData := snippets.SwipeDeckSnippetData{
		DeckID:         "spark_discovery_deck",
		LikeLabel:      "LIKE",
		NopeLabel:      "PASS",
		SuperLikeLabel: "SPARK",
		Cards: []snippets.SwipeCardItem{
			buildSparkCard("spark_maya", "Maya Chen, 28", "Product Designer · Studio K", "Trail runner, espresso nerd, and weekend ceramicist.", "2.4 km", "VERIFIED", theming.ColorSuccessContainer, []string{"Design", "Coffee", "Hiking"}),
			buildSparkCard("spark_arjun", "Arjun Rao, 30", "Acoustic Architect · Soundcraft", "Building modular synths and exploring indie vinyl shops.", "3.1 km", "ACTIVE", theming.ColorInfoContainer, []string{"Vinyl", "Synths", "Cycling"}),
			buildSparkCard("spark_elena", "Elena Morales, 27", "Culinary Photographer", "Always hunting neighborhood street food and rooftop jazz.", "4.8 km", "VERIFIED", theming.ColorWarningContainer, []string{"Photo", "Jazz", "Food"}),
			buildSparkCard("spark_noah", "Noah Park, 29", "Marine Robotics Engineer", "Sailing on Sundays, simmering ramen from scratch.", "5.5 km", "NEW", theming.ColorBrandContainer, []string{"Sailing", "Ramen", "Robotics"}),
		},
		EmptyTitle:   ptrText(atoms.NewText("You've seen everyone nearby", theming.TypographyTitleMedium, theming.ColorTextPrimary)),
		EmptyMessage: ptrText(atoms.NewText("Undo the last card or message your mutual sparks.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
		OnLike:       &deckLike,
		OnNope:       &deckNope,
		OnSuperLike:  &deckSuper,
		OnUndo:       &deckUndo,
		Layout:       seedHorizontalPaddingLayout(),
	}

	boostLayout := seedCardLayout()
	boostLayout.BgColor = atoms.ColorRef(theming.ColorBrandContainer)

	matchRow := func(id, slug, title, subtitle string) domain.Widget {
		toast := actions.NewShowToastAction("chat_"+id, actions.ShowToastPayload{
			Message: "Opened Spark thread with " + title, DurationMs: 2000, Variant: "success",
		})
		return domain.Widget{
			ID: id, TenantID: tenantID, Slug: slug, Name: title,
			TemplateKey: "tmpl_action_row", SnippetType: snippets.SnippetTypeActionRow,
			Props: structToMap(snippets.ActionRowSnippetData{
				Title:        atoms.NewText(title, theming.TypographyTitleSmall, theming.ColorTextPrimary),
				Subtitle:     ptrText(atoms.NewText(subtitle, theming.TypographyBodySmall, theming.ColorTextSecondary)),
				LeadingIcon:  ptrIcon(atoms.NewIcon("person", 20, theming.ColorBrandPrimary)),
				TrailingTag:  ptrTag(atoms.NewTag("SPARK", theming.ColorSuccessContainer)),
				TrailingIcon: ptrIcon(atoms.NewIcon("chevron_right", 18, theming.ColorTextTertiary)),
				OnClick:      &toast,
				Layout:       seedCardLayout(),
			}),
			Tags: []string{"match", "action_row"}, Active: true, CreatedAt: now, UpdatedAt: now,
		}
	}

	backDeck := actions.NewNavigateAction("nav_back_spark_deck", actions.NavigatePayload{Route: "/screens/spark-deck"})
	backDeckBtn := atoms.NewButton("btn_spark_back_deck", "Back to Deck", "primary", &backDeck)

	return []domain.Widget{
		{
			ID: "w_spark_header", TenantID: tenantID, Slug: "spark-header",
			Name: "Abeta Spark Discovery Header", TemplateKey: "tmpl_header_v1", SnippetType: snippets.SnippetTypeHeader,
			Props: structToMap(snippets.HeaderSnippetData{
				Title:          atoms.NewText("Abeta Spark · {{city}}", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:       ptrText(atoms.NewText("Swipe right to like, up for Super Spark", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag:    ptrTag(atoms.NewTag("4 NEW", theming.ColorSuccessContainer)),
				TrailingAction: &matchesBtn,
				Layout:         seedCardLayout(),
			}),
			Tags: []string{"header", "discovery"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		buildSeedChipGroupWidget(
			tenantID, "w_spark_filter_chips", "spark-filter-chips", "Abeta Spark Filter Chips",
			"spark_filters", "spark-deck",
			[]seedChipSpec{
				{id: "nearby", label: "Nearby", selected: true},
				{id: "new", label: "New"},
				{id: "verified", label: "Verified"},
			}, now,
		),
		{
			ID: "w_spark_swipe_deck", TenantID: tenantID, Slug: "spark-swipe-deck",
			Name: "Curated Discovery Swipe Deck", TemplateKey: "tmpl_swipe_deck", SnippetType: snippets.SnippetTypeSwipeDeck,
			Props:    structToMap(deckData),
			Tracking: map[string]any{"event_name": "swipe_deck_loaded", "properties": map[string]any{"deck_id": "spark_discovery_deck"}},
			Tags:     []string{"swipe_deck", "discovery"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_spark_boost_banner", TenantID: tenantID, Slug: "spark-boost-banner",
			Name: "Evening Discovery Boost Banner", TemplateKey: "tmpl_banner_v1", SnippetType: snippets.SnippetTypeBanner,
			Props: structToMap(snippets.BannerSnippetData{
				Title:  atoms.NewText("Evening Boost Live in {{city}}", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Body:   ptrText(atoms.NewText("3× profile visibility active during peak evening hours.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				Icon:   ptrIcon(atoms.NewIcon("bolt", 22, theming.ColorBrandPrimary)),
				Tag:    ptrTag(atoms.NewTag("BOOST", theming.ColorSuccessContainer)),
				Layout: boostLayout,
			}),
			Tags: []string{"banner", "boost", "slotting"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "w_spark_matches_header", TenantID: tenantID, Slug: "spark-matches-header",
			Name: "Mutual Sparks Header", TemplateKey: "tmpl_header_v1", SnippetType: snippets.SnippetTypeHeader,
			Props: structToMap(snippets.HeaderSnippetData{
				Title:       atoms.NewText("Mutual Sparks", theming.TypographyHeadlineSmall, theming.ColorTextPrimary),
				Subtitle:    ptrText(atoms.NewText("Reply within 24 hours to keep your spark alive", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				TrailingTag: ptrTag(atoms.NewTag("3 ACTIVE", theming.ColorSuccessContainer)),
				Layout:      seedCardLayout(),
			}),
			Tags: []string{"header", "matches"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
		matchRow("w_spark_match_maya", "spark-match-maya", "Maya Chen · Matched 10m ago", "Asked: Best pour-over spot near the park?"),
		matchRow("w_spark_match_arjun", "spark-match-arjun", "Arjun Rao · Super-Liked You", "Shared a curated synth & vinyl playlist"),
		matchRow("w_spark_match_elena", "spark-match-elena", "Elena Morales · Matched yesterday", "Recommended a rooftop jazz showcase"),
		{
			ID: "w_spark_back_card", TenantID: tenantID, Slug: "spark-back-card",
			Name: "Return to Swipe Deck Card", TemplateKey: "tmpl_card_v1", SnippetType: snippets.SnippetTypeCard,
			Props: structToMap(snippets.CardSnippetData{
				Title:         atoms.NewText("Keep Discovering", theming.TypographyTitleMedium, theming.ColorTextPrimary),
				Description:   ptrText(atoms.NewText("Jump back into your curated Spark discovery deck.", theming.TypographyBodySmall, theming.ColorTextSecondary)),
				PrimaryButton: &backDeckBtn,
				Layout:        seedCardLayout(),
			}),
			Tags: []string{"card", "navigation"}, Active: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedSparkLayouts(tenantID string, now time.Time) []domain.Layout {
	return []domain.Layout{
		{
			ID: "lay_spark_deck", TenantID: tenantID, Slug: "spark-deck",
			Title: "Abeta Spark — Discovery Deck", Description: "Interactive swipe card deck with category filter chips and evening boost slotting.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Abeta Spark", "Curated Discovery · {{city}}", "back_spark_deck"),
				AdaptiveLayout: buildSeedCenteredAdaptive(720),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_spark_hdr", WidgetID: "w_spark_header", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_spark_chips", WidgetID: "w_spark_filter_chips", Position: 1, Enabled: true},
				{SlotID: "slot_spark_deck", WidgetID: "w_spark_swipe_deck", Position: 2, Enabled: true},
			},
			Tags: []string{"vertical", "swipe-discovery"}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "lay_spark_matches", TenantID: tenantID, Slug: "spark-matches",
			Title: "Abeta Spark — Your Matches", Description: "Mutual sparks list with conversation action rows.",
			Status: domain.LayoutStatusPublished, PublishedVersion: 1, DraftVersion: 1,
			PageConfig: domain.LayoutPageConfig{
				Toolbar:        buildSeedVerticalToolbar("Your Matches", "Abeta Spark · 3 Mutual Connections", "back_spark_matches"),
				AdaptiveLayout: buildSeedCenteredAdaptive(720),
			},
			Slots: []domain.LayoutSlot{
				{SlotID: "slot_matches_hdr", WidgetID: "w_spark_matches_header", Position: 0, Pinned: true, Enabled: true},
				{SlotID: "slot_match_maya", WidgetID: "w_spark_match_maya", Position: 1, Enabled: true},
				{SlotID: "slot_match_arjun", WidgetID: "w_spark_match_arjun", Position: 2, Enabled: true},
				{SlotID: "slot_match_elena", WidgetID: "w_spark_match_elena", Position: 3, Enabled: true},
				{SlotID: "slot_match_back", WidgetID: "w_spark_back_card", Position: 4, Enabled: true},
			},
			Tags: []string{"vertical", "matches"}, CreatedAt: now, UpdatedAt: now,
		},
	}
}
