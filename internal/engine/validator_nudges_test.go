package engine_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/engine"
)

func buildComposedTreeOfDepth(maxDepth int) map[string]any {
	leaf := map[string]any{
		"kind": snippets.ComposedNodeText,
		"text": map[string]any{"text": "leaf"},
	}
	cur := leaf
	for d := maxDepth - 1; d >= 0; d-- {
		cur = map[string]any{
			"kind":     snippets.ComposedNodeColumn,
			"children": []any{cur},
		}
	}
	return cur
}

func TestLayoutValidator_NudgesAndComposedSnippets_Valid(t *testing.T) {
	t.Parallel()

	validator := engine.NewLayoutValidator()

	allLeafKinds := []any{
		map[string]any{"kind": snippets.ComposedNodeText},
		map[string]any{"kind": snippets.ComposedNodeButton},
		map[string]any{"kind": snippets.ComposedNodeTag},
		map[string]any{"kind": snippets.ComposedNodeIcon},
		map[string]any{"kind": snippets.ComposedNodeImage},
		map[string]any{"kind": snippets.ComposedNodeAvatar},
		map[string]any{"kind": snippets.ComposedNodeProgress},
		map[string]any{"kind": snippets.ComposedNodeDivider},
		map[string]any{"kind": snippets.ComposedNodeSpacer},
		map[string]any{"kind": snippets.ComposedNodeStepper},
		map[string]any{"kind": snippets.ComposedNodeCountdown},
		map[string]any{"kind": snippets.ComposedNodeRollingCounter},
		map[string]any{"kind": snippets.ComposedNodeRadioGroup},
		map[string]any{"kind": snippets.ComposedNodeSelect},
		map[string]any{"kind": snippets.ComposedNodeOtpInput},
		map[string]any{"kind": snippets.ComposedNodeChip},
		map[string]any{"kind": snippets.ComposedNodeSegmentedControl},
		map[string]any{"kind": snippets.ComposedNodeRating},
		map[string]any{"kind": snippets.ComposedNodePrice},
		map[string]any{"kind": snippets.ComposedNodeBadge},
		map[string]any{"kind": snippets.ComposedNodeAlertBanner},
		map[string]any{"kind": snippets.ComposedNodeStatTile},
		map[string]any{"kind": snippets.ComposedNodeKeyValueRow},
		map[string]any{"kind": snippets.ComposedNodeQrCode},
		map[string]any{"kind": snippets.ComposedNodeTextField},
		map[string]any{"kind": snippets.ComposedNodeSwitch},
		map[string]any{"kind": snippets.ComposedNodeCheckbox},
		map[string]any{"kind": snippets.ComposedNodeSlider},
		map[string]any{"kind": snippets.ComposedNodeMediaPager},
	}

	if gotKinds := engine.SupportedComposedNodeKinds(); len(gotKinds) != 36 {
		t.Fatalf("expected 36 supported composed node kinds, got %d (%v)", len(gotKinds), gotKinds)
	}
	if !engine.IsValidComposedNodeKind(snippets.ComposedNodeMediaPager) ||
		!engine.IsValidComposedNodeKind(snippets.ComposedNodeScrollRow) ||
		!engine.IsValidComposedNodeKind(snippets.ComposedNodeGrid) ||
		engine.IsValidComposedNodeKind("nonexistent_atom") {
		t.Fatal("IsValidComposedNodeKind returned unexpected results")
	}
	if gotAligns := engine.SupportedBoxAlignments(); len(gotAligns) != 9 {
		t.Fatalf("expected 9 supported box_alignment values, got %d (%v)", len(gotAligns), gotAligns)
	}
	if !engine.IsValidBoxAlignment(snippets.BoxAlignTopStart) || engine.IsValidBoxAlignment("diagonal") {
		t.Fatal("IsValidBoxAlignment returned unexpected results")
	}

	widgets := map[string]domain.Widget{
		"w_nudge_str": {
			ID:          "w_nudge_str",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    "Add $4.50 for free delivery",
				"progress": 0.65,
			},
		},
		"w_nudge_bounds": {
			ID:          "w_nudge_bounds",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    atoms.TextAtom{Text: "Free delivery unlocked!"},
				"progress": atoms.Float64Ptr(1.0),
			},
		},
		"w_nudge_no_progress": {
			ID:          "w_nudge_no_progress",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title": map[string]any{"text": "Flash drop ending soon"},
			},
		},
		"w_marquee_mixed": {
			ID:          "w_marquee_mixed",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props: map[string]any{
				"items": []any{
					map[string]any{"id": "m1", "text": "8-Min Chilled Delivery"},
					snippets.MarqueeItem{ID: "m2", Text: "Free Delivery above $15"},
				},
			},
		},
		"w_marquee_typed": {
			ID:          "w_marquee_typed",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props: map[string]any{
				"items": []snippets.MarqueeItem{
					{ID: "m3", Text: "FRESH20 — 20% Off"},
				},
			},
		},
		"w_marquee_slice_map": {
			ID:          "w_marquee_slice_map",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props: map[string]any{
				"items": []map[string]any{
					{"id": "m4", "text": "Zero handling fees"},
				},
			},
		},
		"w_marquee_omitted": {
			ID:          "w_marquee_omitted",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props:       map[string]any{"strip_id": "empty_marquee"},
		},
		"w_tracker_map": {
			ID:          "w_tracker_map",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props: map[string]any{
				"headline": "Arriving in 14 mins",
				"steps": []any{
					map[string]any{"id": "s1", "label": "Confirmed", "completed": true},
					snippets.OrderTrackerStep{ID: "s2", Label: "Rider En Route", Active: true},
				},
			},
		},
		"w_tracker_typed": {
			ID:          "w_tracker_typed",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props: map[string]any{
				"headline": atoms.TextAtom{Text: "Order #BTS-4091 live"},
				"steps": []snippets.OrderTrackerStep{
					{ID: "s1", Label: "Prepared", Completed: true},
				},
			},
		},
		"w_tracker_slice_map": {
			ID:          "w_tracker_slice_map",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props: map[string]any{
				"headline": map[string]any{"text": "Kitchen preparing"},
				"steps": []map[string]any{
					{"id": "s1", "label": "Placed"},
				},
			},
		},
		"w_composed_all_kinds": {
			ID:          "w_composed_all_kinds",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": map[string]any{
					"kind": snippets.ComposedNodeCard,
					"children": []any{
						map[string]any{
							"kind": snippets.ComposedNodeColumn,
							"children": []any{
								map[string]any{"kind": snippets.ComposedNodeRow, "children": allLeafKinds[:6]},
								map[string]any{
									"kind": snippets.ComposedNodeBox,
									"children": []any{
										map[string]any{"kind": snippets.ComposedNodeImage, "box_alignment": snippets.BoxAlignCenter},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignTopStart},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignTopCenter},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignTopEnd},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignCenterStart},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignCenterEnd},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignBottomStart},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignBottomCenter},
										map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": snippets.BoxAlignBottomEnd},
										map[string]any{"kind": snippets.ComposedNodeText, "box_alignment": ""},
									},
								},
								map[string]any{"kind": snippets.ComposedNodeFlowRow, "children": allLeafKinds[6:18]},
								map[string]any{"kind": snippets.ComposedNodeScrollRow, "children": allLeafKinds[18:24]},
								map[string]any{"kind": snippets.ComposedNodeGrid, "columns": 2, "children": allLeafKinds[24:]},
							},
						},
					},
				},
			},
		},
		"w_composed_depth6_ok": {
			ID:          "w_composed_depth6_ok",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": buildComposedTreeOfDepth(6),
			},
		},
		"w_composed_typed": {
			ID:          "w_composed_typed",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": snippets.VStack(8,
					snippets.BoxNode(
						snippets.TextNode(atoms.TextAtom{Text: "Hello"}).WithBoxAlignment(snippets.BoxAlignTopStart),
					),
					snippets.ScrollRowNode(8, snippets.TextFieldNode(atoms.TextFieldAtom{FieldID: "q", Label: "Search"})),
					snippets.GridNode(2, 8, snippets.MediaPagerNode(atoms.NewMediaPager("mp1", nil, 160))),
				),
			},
		},
	}

	slots := make([]domain.LayoutSlot, 0, len(widgets))
	pos := 0
	for id := range widgets {
		slots = append(slots, domain.LayoutSlot{SlotID: "slot_" + id, WidgetID: id, Position: pos, Enabled: true})
		pos++
	}

	layout := domain.Layout{
		Slug:  "nudges-composed-valid",
		Title: "Valid Nudges & Composed",
		Slots: slots,
	}

	if issues := validator.ValidateLayout(layout, widgets); len(issues) != 0 {
		t.Fatalf("expected 0 validation issues, got: %+v", issues)
	}
}

func TestLayoutValidator_NudgesAndComposedSnippets_Invalid(t *testing.T) {
	t.Parallel()

	validator := engine.NewLayoutValidator()

	widgets := map[string]domain.Widget{
		"w_nudge_missing_title_bad_prog": {
			ID:          "w_nudge_missing_title_bad_prog",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    "   ",
				"progress": 1.25,
			},
		},
		"w_nudge_neg_prog": {
			ID:          "w_nudge_neg_prog",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    "Valid title",
				"progress": -0.1,
			},
		},
		"w_nudge_nan_prog": {
			ID:          "w_nudge_nan_prog",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    "Valid title",
				"progress": math.NaN(),
			},
		},
		"w_nudge_str_prog": {
			ID:          "w_nudge_str_prog",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    "Valid title",
				"progress": "half",
			},
		},
		"w_marquee_not_slice": {
			ID:          "w_marquee_not_slice",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props:       map[string]any{"items": "not-a-slice"},
		},
		"w_marquee_bad_items": {
			ID:          "w_marquee_bad_items",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props: map[string]any{
				"items": []any{
					42,
					map[string]any{"id": "m1", "text": "   "},
					snippets.MarqueeItem{ID: "m2", Text: ""},
				},
			},
		},
		"w_tracker_missing_headline_bad_steps": {
			ID:          "w_tracker_missing_headline_bad_steps",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props: map[string]any{
				"headline": "",
				"steps":    "not-a-slice",
			},
		},
		"w_tracker_bad_step_entries": {
			ID:          "w_tracker_bad_step_entries",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props: map[string]any{
				"headline": "Live Order",
				"steps": []any{
					false,
					map[string]any{"id": "s1", "label": ""},
					snippets.OrderTrackerStep{ID: "s2", Label: "  "},
				},
			},
		},
		"w_composed_missing_root": {
			ID:          "w_composed_missing_root",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props:       map[string]any{"block_id": "b1"},
		},
		"w_composed_bad_root_kind_and_children": {
			ID:          "w_composed_bad_root_kind_and_children",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": map[string]any{
					"kind":     "unknown_widget",
					"children": "not-a-slice",
				},
			},
		},
		"w_composed_depth7_overflow": {
			ID:          "w_composed_depth7_overflow",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": buildComposedTreeOfDepth(7),
			},
		},
		"w_composed_empty_kind": {
			ID:          "w_composed_empty_kind",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": map[string]any{
					"kind": "",
				},
			},
		},
		"w_nudge_json_number_ok": {
			ID:          "w_nudge_json_number_ok",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"title":    "JSON number progress",
				"progress": json.Number("0.5"),
			},
		},
		"w_composed_bad_box_align": {
			ID:          "w_composed_bad_box_align",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": map[string]any{
					"kind": snippets.ComposedNodeBox,
					"children": []any{
						map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": "diagonal_corner"},
						map[string]any{"kind": snippets.ComposedNodeTag, "box_alignment": 42},
					},
				},
			},
		},
		"w_composed_typed_bad_box_align": {
			ID:          "w_composed_typed_bad_box_align",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"root": snippets.BoxNode(
					snippets.TagNode(atoms.TagAtom{Label: "BAD"}).WithBoxAlignment("top_left"),
				),
			},
		},
	}

	layout := domain.Layout{
		Slug:  "invalid-nudges-composed",
		Title: "Invalid Nudges & Composed",
		Slots: []domain.LayoutSlot{
			{SlotID: "s0", WidgetID: "w_nudge_missing_title_bad_prog", Position: 0, Enabled: true},
			{SlotID: "s1", WidgetID: "w_nudge_neg_prog", Position: 1, Enabled: true},
			{SlotID: "s2", WidgetID: "w_nudge_nan_prog", Position: 2, Enabled: true},
			{SlotID: "s3", WidgetID: "w_nudge_str_prog", Position: 3, Enabled: true},
			{SlotID: "s4", WidgetID: "w_marquee_not_slice", Position: 4, Enabled: true},
			{SlotID: "s5", WidgetID: "w_marquee_bad_items", Position: 5, Enabled: true},
			{SlotID: "s6", WidgetID: "w_tracker_missing_headline_bad_steps", Position: 6, Enabled: true},
			{SlotID: "s7", WidgetID: "w_tracker_bad_step_entries", Position: 7, Enabled: true},
			{SlotID: "s8", WidgetID: "w_composed_missing_root", Position: 8, Enabled: true},
			{SlotID: "s9", WidgetID: "w_composed_bad_root_kind_and_children", Position: 9, Enabled: true},
			{SlotID: "s10", WidgetID: "w_composed_depth7_overflow", Position: 10, Enabled: true},
			{SlotID: "s11", WidgetID: "w_composed_empty_kind", Position: 11, Enabled: true},
			{SlotID: "s12", WidgetID: "w_nudge_json_number_ok", Position: 12, Enabled: true},
			{SlotID: "s13", WidgetID: "w_composed_bad_box_align", Position: 13, Enabled: true},
			{SlotID: "s14", WidgetID: "w_composed_typed_bad_box_align", Position: 14, Enabled: true},
		},
	}

	issues := validator.ValidateLayout(layout, widgets)

	byPathCode := make(map[string]string, len(issues))
	for _, iss := range issues {
		byPathCode[iss.Path] = iss.Code
	}

	wantPathCodes := map[string]string{
		"/slots/0/props/title":         "required",
		"/slots/0/props/progress":      "invalid_progress",
		"/slots/1/props/progress":      "invalid_progress",
		"/slots/2/props/progress":      "invalid_progress",
		"/slots/3/props/progress":      "invalid_progress",
		"/slots/4/props/items":         "invalid_marquee_items",
		"/slots/5/props/items/0":       "invalid_marquee_item",
		"/slots/5/props/items/1/text":  "required",
		"/slots/5/props/items/2/text":  "required",
		"/slots/6/props/headline":      "required",
		"/slots/6/props/steps":         "invalid_order_steps",
		"/slots/7/props/steps/0":       "invalid_order_step",
		"/slots/7/props/steps/1/label": "required",
		"/slots/7/props/steps/2/label": "required",
		"/slots/8/props/root":          "required",
		"/slots/9/props/root/kind":     "invalid_composed_node_kind",
		"/slots/9/props/root/children": "invalid_composed_children",
		"/slots/10/props/root/children/0/children/0/children/0/children/0/children/0/children/0/children/0": "max_depth_exceeded",
		"/slots/11/props/root/kind":                     "required",
		"/slots/13/props/root/children/0/box_alignment": "invalid_box_alignment",
		"/slots/13/props/root/children/1/box_alignment": "invalid_box_alignment",
		"/slots/14/props/root/children/0/box_alignment": "invalid_box_alignment",
	}

	for path, wantCode := range wantPathCodes {
		if gotCode, ok := byPathCode[path]; !ok || gotCode != wantCode {
			t.Errorf("issue at %q = %q (exists=%v), want %q (all issues: %+v)", path, gotCode, ok, wantCode, issues)
		}
	}
	if _, hasSlot12 := byPathCode["/slots/12/props/progress"]; hasSlot12 {
		t.Errorf("expected slot 12 json.Number progress 0.5 to be valid, got issues: %+v", issues)
	}
}
