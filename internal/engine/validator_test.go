package engine_test

import (
	"slices"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/engine"
)

func TestSupportedSnippetTypes_All21CanonicalTypes(t *testing.T) {
	t.Parallel()

	want := []string{
		snippets.SnippetTypeAccordion,
		snippets.SnippetTypeActionRow,
		snippets.SnippetTypeBanner,
		snippets.SnippetTypeCard,
		snippets.SnippetTypeCarousel,
		snippets.SnippetTypeChipGroup,
		snippets.SnippetTypeComposedBlock,
		snippets.SnippetTypeDivider,
		snippets.SnippetTypeEmptyState,
		snippets.SnippetTypeForm,
		snippets.SnippetTypeGrid,
		snippets.SnippetTypeHeader,
		snippets.SnippetTypeMarqueeStrip,
		snippets.SnippetTypeMetricCard,
		snippets.SnippetTypeNudgeBar,
		snippets.SnippetTypeOrderTracker,
		snippets.SnippetTypeProductCard,
		snippets.SnippetTypeSearchBar,
		snippets.SnippetTypeSpacer,
		snippets.SnippetTypeSwipeDeck,
		snippets.SnippetTypeTabs,
	}

	got := engine.SupportedSnippetTypes()
	if len(got) != 21 {
		t.Fatalf("SupportedSnippetTypes() len = %d, want 21 (%v)", len(got), got)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("SupportedSnippetTypes() = %v, want %v", got, want)
	}

	for _, st := range want {
		if !engine.IsValidSnippetType(st) {
			t.Errorf("IsValidSnippetType(%q) = false, want true", st)
		}
	}
	if engine.IsValidSnippetType("legacy_unknown") {
		t.Error("IsValidSnippetType(legacy_unknown) = true, want false")
	}
}

func TestLayoutValidator_InteractiveSnippets_Valid(t *testing.T) {
	t.Parallel()

	validator := engine.NewLayoutValidator()

	widgets := map[string]domain.Widget{
		"w_swipe_map": {
			ID:          "w_swipe_map",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props: map[string]any{
				"deck_id": "spark_feed",
				"cards": []any{
					map[string]any{
						"card_id": "c1",
						"title":   "Meera, 27",
					},
					map[string]any{
						"card_id": "c2",
						"title":   map[string]any{"text": "Rohan, 29"},
					},
					snippets.SwipeCardItem{
						CardID: "c3",
						Title:  atoms.TextAtom{Text: "Ananya, 26"},
					},
				},
			},
		},
		"w_swipe_typed": {
			ID:          "w_swipe_typed",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props: map[string]any{
				"cards": []snippets.SwipeCardItem{
					{CardID: "c_typed", Title: atoms.TextAtom{Text: "Kabir, 30"}},
				},
			},
		},
		"w_swipe_slice_map": {
			ID:          "w_swipe_slice_map",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props: map[string]any{
				"cards": []map[string]any{
					{"card_id": "c_sm", "title": atoms.TextAtom{Text: "Dev, 28"}},
				},
			},
		},
		"w_search_valid": {
			ID:          "w_search_valid",
			SnippetType: snippets.SnippetTypeSearchBar,
			Active:      true,
			Props: map[string]any{
				"search_id":   "catalog_search",
				"placeholder": "Search produce...",
			},
		},
		"w_search_minimal": {
			ID:          "w_search_minimal",
			SnippetType: snippets.SnippetTypeSearchBar,
			Active:      true,
			Props:       map[string]any{},
		},
		"w_chips_map": {
			ID:          "w_chips_map",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props: map[string]any{
				"group_id": "filters",
				"chips": []any{
					map[string]any{"id": "all", "label": "All"},
					snippets.FilterChipItem{ID: "fast", Label: "Under 15m", Selected: true},
				},
			},
		},
		"w_chips_typed": {
			ID:          "w_chips_typed",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props: map[string]any{
				"chips": []snippets.FilterChipItem{
					{ID: "veg", Label: "Vegetarian"},
				},
			},
		},
		"w_chips_slice_map": {
			ID:          "w_chips_slice_map",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props: map[string]any{
				"chips": []map[string]any{
					{"id": "top", "label": "Top Rated"},
				},
			},
		},
	}

	layout := domain.Layout{
		Slug:  "discovery-hub",
		Title: "Discovery Hub",
		Slots: []domain.LayoutSlot{
			{SlotID: "s1", WidgetID: "w_swipe_map", Position: 0, Enabled: true},
			{SlotID: "s2", WidgetID: "w_swipe_typed", Position: 1, Enabled: true},
			{SlotID: "s3", WidgetID: "w_swipe_slice_map", Position: 2, Enabled: true},
			{SlotID: "s4", WidgetID: "w_search_valid", Position: 3, Enabled: true},
			{SlotID: "s5", WidgetID: "w_search_minimal", Position: 4, Enabled: true},
			{SlotID: "s6", WidgetID: "w_chips_map", Position: 5, Enabled: true},
			{SlotID: "s7", WidgetID: "w_chips_typed", Position: 6, Enabled: true},
			{SlotID: "s8", WidgetID: "w_chips_slice_map", Position: 7, Enabled: true},
		},
	}

	if issues := validator.ValidateLayout(layout, widgets); len(issues) != 0 {
		t.Fatalf("expected 0 validation issues for valid interactive snippets, got: %+v", issues)
	}
}

func TestLayoutValidator_InteractiveSnippets_Invalid(t *testing.T) {
	t.Parallel()

	validator := engine.NewLayoutValidator()

	widgets := map[string]domain.Widget{
		"w_swipe_not_slice": {
			ID:          "w_swipe_not_slice",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props:       map[string]any{"cards": "not-a-list"},
		},
		"w_swipe_bad_entries": {
			ID:          "w_swipe_bad_entries",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props: map[string]any{
				"cards": []any{
					"not-an-object",
					map[string]any{"card_id": "", "title": ""},
					map[string]any{"card_id": "c_ok", "title": map[string]any{"text": "  "}},
					snippets.SwipeCardItem{CardID: "", Title: atoms.TextAtom{Text: ""}},
				},
			},
		},
		"w_swipe_typed_bad": {
			ID:          "w_swipe_typed_bad",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props: map[string]any{
				"cards": []snippets.SwipeCardItem{
					{CardID: "  ", Title: atoms.TextAtom{Text: ""}},
				},
			},
		},
		"w_search_bad_id": {
			ID:          "w_search_bad_id",
			SnippetType: snippets.SnippetTypeSearchBar,
			Active:      true,
			Props:       map[string]any{"search_id": 42},
		},
		"w_search_override_bad": {
			ID:          "w_search_override_bad",
			SnippetType: snippets.SnippetTypeSearchBar,
			Active:      true,
			Props:       map[string]any{"search_id": "ok_id"},
		},
		"w_chips_not_slice": {
			ID:          "w_chips_not_slice",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props:       map[string]any{"chips": map[string]any{"id": "x"}},
		},
		"w_chips_bad_entries": {
			ID:          "w_chips_bad_entries",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props: map[string]any{
				"chips": []any{
					123,
					map[string]any{"id": "", "label": ""},
					snippets.FilterChipItem{ID: "", Label: "  "},
				},
			},
		},
		"w_chips_typed_bad": {
			ID:          "w_chips_typed_bad",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props: map[string]any{
				"chips": []snippets.FilterChipItem{
					{ID: "", Label: ""},
				},
			},
		},
	}

	layout := domain.Layout{
		Slug:  "invalid-interactive",
		Title: "Invalid Interactive Snippets",
		Slots: []domain.LayoutSlot{
			{SlotID: "s0", WidgetID: "w_swipe_not_slice", Position: 0, Enabled: true},
			{SlotID: "s1", WidgetID: "w_swipe_bad_entries", Position: 1, Enabled: true},
			{SlotID: "s2", WidgetID: "w_swipe_typed_bad", Position: 2, Enabled: true},
			{SlotID: "s3", WidgetID: "w_search_bad_id", Position: 3, Enabled: true},
			{SlotID: "s4", WidgetID: "w_search_override_bad", Position: 4, Enabled: true, Overrides: map[string]any{"search_id": true}},
			{SlotID: "s5", WidgetID: "w_chips_not_slice", Position: 5, Enabled: true},
			{SlotID: "s6", WidgetID: "w_chips_bad_entries", Position: 6, Enabled: true},
			{SlotID: "s7", WidgetID: "w_chips_typed_bad", Position: 7, Enabled: true},
		},
	}

	issues := validator.ValidateLayout(layout, widgets)

	byPathCode := make(map[string]string, len(issues))
	for _, iss := range issues {
		byPathCode[iss.Path] = iss.Code
	}

	wantPathCodes := map[string]string{
		"/slots/0/props/cards":           "invalid_swipe_cards",
		"/slots/1/props/cards/0":         "invalid_swipe_card",
		"/slots/1/props/cards/1/card_id": "required",
		"/slots/1/props/cards/1/title":   "required",
		"/slots/1/props/cards/2/title":   "required",
		"/slots/1/props/cards/3/card_id": "required",
		"/slots/1/props/cards/3/title":   "required",
		"/slots/2/props/cards/0/card_id": "required",
		"/slots/2/props/cards/0/title":   "required",
		"/slots/3/props/search_id":       "invalid_search_id",
		"/slots/4/props/search_id":       "invalid_search_id",
		"/slots/5/props/chips":           "invalid_chips",
		"/slots/6/props/chips/0":         "invalid_chip",
		"/slots/6/props/chips/1/id":      "required",
		"/slots/6/props/chips/1/label":   "required",
		"/slots/6/props/chips/2/id":      "required",
		"/slots/6/props/chips/2/label":   "required",
		"/slots/7/props/chips/0/id":      "required",
		"/slots/7/props/chips/0/label":   "required",
	}

	for path, wantCode := range wantPathCodes {
		if gotCode, ok := byPathCode[path]; !ok || gotCode != wantCode {
			t.Errorf("issue at %q = %q (exists=%v), want %q (all issues: %+v)", path, gotCode, ok, wantCode, issues)
		}
	}
}
