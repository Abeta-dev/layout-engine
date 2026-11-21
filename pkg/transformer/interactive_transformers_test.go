package transformer_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/pkg/transformer"
)

func TestTransformSwipeDeck_PromotionDefaultsAndRoundTrip(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{
		TenantSlug: "abeta-spark",
		City:       "Bengaluru",
		Zone:       "indiranagar",
		Segment:    "vip",
	}

	origProps := map[string]any{
		"like_label":       "SPARK",
		"nope_label":       "PASS",
		"super_like_label": "STAR",
		"empty_title":      "No more profiles in {{city}}",
		"empty_message":    "Expand your {{zone}} radius to see more.",
		"cards": []any{
			map[string]any{
				"card_id":       "p_1",
				"title":         "Aarav, 28",
				"subtitle":      "Staff Engineer · {{city}}",
				"bio":           "Trail running and pour-over coffee.",
				"distance_text": "1.8 km away",
				"badge":         "VERIFIED",
				"image":         "https://cdn.abeta.dev/spark/aarav.jpg",
				"tags": []any{
					"Running",
					"{{segment}}",
					map[string]any{"label": "Coffee"},
				},
			},
			map[string]any{
				"card_id": "p_2",
				"title":   map[string]any{"text": "Diya, 26"},
				"tags":    []string{"Art", "Music"},
			},
		},
	}

	env, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "slot_spark_deck", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_spark",
			Slug:        "spark_deck_widget",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props:       origProps,
			Tracking: map[string]any{
				"event_name": "deck_impression",
				"properties": map[string]any{"city": "{{city}}", "count": 2},
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(swipe_deck) failed: %v", err)
	}

	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(swipe_deck) failed: %v", err)
	}
	var roundTripped snippets.SnippetEnvelope
	if err := json.Unmarshal(raw, &roundTripped); err != nil {
		t.Fatalf("json.Unmarshal(swipe_deck) failed: %v", err)
	}

	data, err := snippets.DecodeSnippetData[snippets.SwipeDeckSnippetData](roundTripped)
	if err != nil {
		t.Fatalf("DecodeSnippetData(swipe_deck) failed: %v", err)
	}
	if data.DeckID != "slot_spark_deck" {
		t.Fatalf("DeckID = %q, want slot_spark_deck", data.DeckID)
	}
	if data.EmptyTitle == nil || data.EmptyTitle.Text != "No more profiles in Bengaluru" {
		t.Fatalf("EmptyTitle = %+v, want interpolated Bengaluru title", data.EmptyTitle)
	}
	if data.EmptyMessage == nil || data.EmptyMessage.Text != "Expand your indiranagar radius to see more." {
		t.Fatalf("EmptyMessage = %+v, want interpolated indiranagar message", data.EmptyMessage)
	}
	if len(data.Cards) != 2 {
		t.Fatalf("len(Cards) = %d, want 2", len(data.Cards))
	}

	c0 := data.Cards[0]
	if c0.Title.Text != "Aarav, 28" || c0.Subtitle == nil || c0.Subtitle.Text != "Staff Engineer · Bengaluru" {
		t.Fatalf("unexpected card[0] title/subtitle: %+v", c0)
	}
	if c0.Bio == nil || c0.Bio.Text != "Trail running and pour-over coffee." {
		t.Fatalf("unexpected card[0] bio: %+v", c0.Bio)
	}
	if c0.DistanceText == nil || c0.DistanceText.Text != "1.8 km away" {
		t.Fatalf("unexpected card[0] distance_text: %+v", c0.DistanceText)
	}
	if c0.Badge == nil || c0.Badge.Label != "VERIFIED" {
		t.Fatalf("unexpected card[0] badge: %+v", c0.Badge)
	}
	if c0.Image == nil || c0.Image.URL != "https://cdn.abeta.dev/spark/aarav.jpg" {
		t.Fatalf("unexpected card[0] image: %+v", c0.Image)
	}
	if len(c0.Tags) != 3 || c0.Tags[0].Label != "Running" || c0.Tags[1].Label != "vip" || c0.Tags[2].Label != "Coffee" {
		t.Fatalf("unexpected card[0] tags: %+v", c0.Tags)
	}

	c1 := data.Cards[1]
	if c1.Title.Text != "Diya, 26" || len(c1.Tags) != 2 || c1.Tags[0].Label != "Art" || c1.Tags[1].Label != "Music" {
		t.Fatalf("unexpected card[1]: %+v", c1)
	}

	// Verify Tracking.Properties emits string values per KMP wire contract.
	if roundTripped.Tracking == nil || roundTripped.Tracking.Properties["count"] != "2" || roundTripped.Tracking.Properties["city"] != "Bengaluru" {
		t.Fatalf("expected stringified tracking properties, got %+v", roundTripped.Tracking)
	}

	// Verify empty deck produces non-nil JSON array for cards (`"cards":[]`).
	emptyEnv, err := reg.Transform(transformer.TransformInput{
		Widget: domain.Widget{
			ID:          "w_empty_deck",
			Slug:        "empty_deck",
			SnippetType: snippets.SnippetTypeSwipeDeck,
			Active:      true,
			Props:       map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("Transform(empty swipe_deck) failed: %v", err)
	}
	emptyJSON, err := json.Marshal(emptyEnv)
	if err != nil {
		t.Fatalf("json.Marshal(empty swipe_deck) failed: %v", err)
	}
	if !strings.Contains(string(emptyJSON), `"cards":[]`) {
		t.Fatalf("expected non-nil empty cards array in JSON, got: %s", string(emptyJSON))
	}

	// Verify original props were not mutated.
	origCards := origProps["cards"].([]any)
	origCard0 := origCards[0].(map[string]any)
	if _, isStr := origCard0["title"].(string); !isStr {
		t.Fatalf("original props card[0].title was mutated: %#v", origCard0["title"])
	}
}

func TestTransformSearchBar_DefaultsClearableAndPromotion(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{City: "Mumbai"}

	// Case 1: Omitted search_id, placeholder, clearable -> defaults applied.
	envDefault, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "hero_search", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_search_default",
			SnippetType: snippets.SnippetTypeSearchBar,
			Active:      true,
			Props: map[string]any{
				"leading_icon":  "search",
				"trailing_icon": "tune",
				"query":         "Fresh {{city}}",
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(search_bar default) failed: %v", err)
	}

	raw, err := json.Marshal(envDefault)
	if err != nil {
		t.Fatalf("json.Marshal(search_bar) failed: %v", err)
	}
	var decodedEnv snippets.SnippetEnvelope
	if err := json.Unmarshal(raw, &decodedEnv); err != nil {
		t.Fatalf("json.Unmarshal(search_bar) failed: %v", err)
	}
	data, err := snippets.DecodeSnippetData[snippets.SearchBarSnippetData](decodedEnv)
	if err != nil {
		t.Fatalf("DecodeSnippetData(search_bar) failed: %v", err)
	}
	if data.SearchID != "hero_search" {
		t.Fatalf("SearchID = %q, want hero_search", data.SearchID)
	}
	if data.Placeholder != "Search..." {
		t.Fatalf("Placeholder = %q, want Search...", data.Placeholder)
	}
	if !data.Clearable {
		t.Fatal("Clearable = false, want true when omitted")
	}
	if data.Query != "Fresh Mumbai" {
		t.Fatalf("Query = %q, want Fresh Mumbai", data.Query)
	}
	if data.LeadingIcon == nil || data.LeadingIcon.Name != "search" || data.TrailingIcon == nil || data.TrailingIcon.Name != "tune" {
		t.Fatalf("unexpected promoted icons: leading=%+v trailing=%+v", data.LeadingIcon, data.TrailingIcon)
	}

	// Case 2: Explicit clearable=false and custom search_id/placeholder preserved.
	envExplicit, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "slot_ignored", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_search_explicit",
			SnippetType: snippets.SnippetTypeSearchBar,
			Active:      true,
			Props: map[string]any{
				"search_id":    "trade_rfq_search",
				"placeholder":  "Search suppliers in {{city}}",
				"clearable":    false,
				"submit_label": "Find",
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(search_bar explicit) failed: %v", err)
	}
	dataExplicit, err := snippets.DecodeSnippetData[snippets.SearchBarSnippetData](envExplicit)
	if err != nil {
		t.Fatalf("DecodeSnippetData(search_bar explicit) failed: %v", err)
	}
	if dataExplicit.SearchID != "trade_rfq_search" || dataExplicit.Clearable || dataExplicit.Placeholder != "Search suppliers in Mumbai" || dataExplicit.SubmitLabel != "Find" {
		t.Fatalf("unexpected explicit search_bar data: %+v", dataExplicit)
	}
}

func TestTransformChipGroup_PromotionDefaultsAndRoundTrip(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{City: "Delhi", Segment: "pro"}

	env, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "slot_filters", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_chips",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props: map[string]any{
				"title":        "Browse {{city}} ({{segment}})",
				"multi_select": true,
				"chips": []any{
					map[string]any{"id": "c_all", "label": "All", "selected": true, "badge": "18"},
					map[string]any{"id": "c_city", "label": "Near {{city}}", "icon": "location_on", "badge": "5"},
				},
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(chip_group) failed: %v", err)
	}

	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(chip_group) failed: %v", err)
	}
	var roundTripped snippets.SnippetEnvelope
	if err := json.Unmarshal(raw, &roundTripped); err != nil {
		t.Fatalf("json.Unmarshal(chip_group) failed: %v", err)
	}
	data, err := snippets.DecodeSnippetData[snippets.ChipGroupSnippetData](roundTripped)
	if err != nil {
		t.Fatalf("DecodeSnippetData(chip_group) failed: %v", err)
	}
	if data.GroupID != "slot_filters" {
		t.Fatalf("GroupID = %q, want slot_filters", data.GroupID)
	}
	if data.Title == nil || data.Title.Text != "Browse Delhi (pro)" {
		t.Fatalf("Title = %+v, want Browse Delhi (pro)", data.Title)
	}
	if !data.MultiSelect || len(data.Chips) != 2 {
		t.Fatalf("unexpected chip_group metadata: %+v", data)
	}
	if data.Chips[0].ID != "c_all" || !data.Chips[0].Selected || data.Chips[0].Badge != "18" {
		t.Fatalf("unexpected chips[0]: %+v", data.Chips[0])
	}
	if data.Chips[1].Label != "Near Delhi" || data.Chips[1].Icon == nil || data.Chips[1].Icon.Name != "location_on" {
		t.Fatalf("unexpected chips[1]: %+v", data.Chips[1])
	}

	// Verify empty chip group produces non-nil JSON array for chips (`"chips":[]`).
	emptyEnv, err := reg.Transform(transformer.TransformInput{
		Widget: domain.Widget{
			ID:          "w_empty_chips",
			Slug:        "empty_chips",
			SnippetType: snippets.SnippetTypeChipGroup,
			Active:      true,
			Props:       map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("Transform(empty chip_group) failed: %v", err)
	}
	emptyJSON, err := json.Marshal(emptyEnv)
	if err != nil {
		t.Fatalf("json.Marshal(empty chip_group) failed: %v", err)
	}
	if !strings.Contains(string(emptyJSON), `"chips":[]`) {
		t.Fatalf("expected non-nil empty chips array in JSON, got: %s", string(emptyJSON))
	}
}
