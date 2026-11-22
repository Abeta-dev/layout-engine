package transformer_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/pkg/transformer"
)

func TestTransformNudgeBar_DefaultsShorthandsAndInterpolation(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{City: "Bengaluru", Segment: "vip"}

	origProps := map[string]any{
		"title":    "Flash Drop in {{city}}",
		"subtitle": "Exclusive for {{segment}} baskets",
		"icon":     "bolt",
		"badge":    "FLASH DROP",
		"progress": 0.75,
		"countdown": map[string]any{
			"countdown_id":             "flash_cd",
			"remaining_seconds":        420,
			"label":                    "Ends in",
			"urgent_threshold_seconds": 60,
		},
	}

	env, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "slot_nudge_flash", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_nudge_flash",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props:       origProps,
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(nudge_bar) failed: %v", err)
	}

	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(nudge_bar) failed: %v", err)
	}
	var decoded snippets.SnippetEnvelope
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(nudge_bar) failed: %v", err)
	}

	data, err := snippets.DecodeSnippetData[snippets.NudgeBarSnippetData](decoded)
	if err != nil {
		t.Fatalf("DecodeSnippetData(nudge_bar) failed: %v", err)
	}
	if data.NudgeID != "slot_nudge_flash" {
		t.Fatalf("NudgeID = %q, want slot_nudge_flash", data.NudgeID)
	}
	if !data.Dismissible {
		t.Fatal("Dismissible = false, want true when omitted")
	}
	if data.Title.Text != "Flash Drop in Bengaluru" || data.Subtitle == nil || data.Subtitle.Text != "Exclusive for vip baskets" {
		t.Fatalf("unexpected nudge title/subtitle: %+v", data)
	}
	if data.Icon == nil || data.Icon.Name != "bolt" || data.Badge == nil || data.Badge.Label != "FLASH DROP" {
		t.Fatalf("unexpected promoted icon/badge: icon=%+v badge=%+v", data.Icon, data.Badge)
	}
	if data.Progress == nil || *data.Progress != 0.75 || data.Countdown == nil || data.Countdown.RemainingSeconds != 420 {
		t.Fatalf("unexpected progress/countdown: %+v", data)
	}

	// Verify explicit dismissible=false and explicit nudge_id preserved.
	envNonDismiss, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "ignored_slot", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_nudge_locked",
			SnippetType: snippets.SnippetTypeNudgeBar,
			Active:      true,
			Props: map[string]any{
				"nudge_id":    "locked_nudge",
				"title":       "Locked Nudge",
				"dismissible": false,
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(nudge_bar dismissible=false) failed: %v", err)
	}
	dataLocked, err := snippets.DecodeSnippetData[snippets.NudgeBarSnippetData](envNonDismiss)
	if err != nil {
		t.Fatalf("DecodeSnippetData(locked nudge_bar) failed: %v", err)
	}
	if dataLocked.NudgeID != "locked_nudge" || dataLocked.Dismissible {
		t.Fatalf("expected explicit nudge_id and dismissible=false, got %+v", dataLocked)
	}

	// Ensure original props map was not mutated.
	if _, isStr := origProps["title"].(string); !isStr {
		t.Fatalf("origProps title was mutated: %#v", origProps["title"])
	}
}

func TestTransformMarqueeStrip_DefaultsPromotionAndEmptyArray(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{City: "Mumbai"}

	env, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "slot_marquee", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_marquee",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props: map[string]any{
				"items": []any{
					map[string]any{"id": "m1", "text": "8-Min Express in {{city}}", "icon": "bolt", "badge": "LIVE"},
					map[string]any{"id": "m2", "text": "Free Delivery over $15", "icon": "local_shipping"},
				},
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(marquee_strip) failed: %v", err)
	}

	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(marquee_strip) failed: %v", err)
	}
	var decoded snippets.SnippetEnvelope
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(marquee_strip) failed: %v", err)
	}

	data, err := snippets.DecodeSnippetData[snippets.MarqueeStripSnippetData](decoded)
	if err != nil {
		t.Fatalf("DecodeSnippetData(marquee_strip) failed: %v", err)
	}
	if data.StripID != "slot_marquee" || data.VelocityDpPerSec != 48 || data.Separator != "•" {
		t.Fatalf("unexpected marquee defaults: %+v", data)
	}
	if len(data.Items) != 2 || data.Items[0].Text != "8-Min Express in Mumbai" {
		t.Fatalf("unexpected marquee items: %+v", data.Items)
	}
	if data.Items[0].Icon == nil || data.Items[0].Icon.Name != "bolt" || data.Items[0].Badge == nil || data.Items[0].Badge.Label != "LIVE" {
		t.Fatalf("marquee item[0] icon/badge not promoted: %+v", data.Items[0])
	}

	// Empty marquee emits non-nil JSON array ("items":[]).
	emptyEnv, err := reg.Transform(transformer.TransformInput{
		Widget: domain.Widget{
			ID:          "w_marquee_empty",
			Slug:        "empty_marquee",
			SnippetType: snippets.SnippetTypeMarqueeStrip,
			Active:      true,
			Props:       map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("Transform(empty marquee_strip) failed: %v", err)
	}
	emptyJSON, err := json.Marshal(emptyEnv)
	if err != nil {
		t.Fatalf("json.Marshal(empty marquee_strip) failed: %v", err)
	}
	if !strings.Contains(string(emptyJSON), `"items":[]`) {
		t.Fatalf("expected non-nil empty items array, got: %s", string(emptyJSON))
	}
}

func TestTransformOrderTracker_DefaultsPromotionAndEmptySteps(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{City: "Pune"}

	env, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "slot_tracker", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_tracker",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props: map[string]any{
				"order_code":    "#BTS-4091",
				"headline":      "Rider approaching {{city}} hub",
				"eta_badge":     "14 MINS",
				"rider_name":    "Karan M.",
				"rider_caption": "Electric bike · 4.9 ★",
				"steps": []any{
					map[string]any{"id": "s1", "label": "Confirmed", "icon": "check_circle", "completed": true},
					map[string]any{"id": "s2", "label": "En Route in {{city}}", "icon": "local_shipping", "active": true},
				},
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(order_tracker) failed: %v", err)
	}

	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(order_tracker) failed: %v", err)
	}
	var decoded snippets.SnippetEnvelope
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(order_tracker) failed: %v", err)
	}

	data, err := snippets.DecodeSnippetData[snippets.OrderTrackerSnippetData](decoded)
	if err != nil {
		t.Fatalf("DecodeSnippetData(order_tracker) failed: %v", err)
	}
	if data.TrackerID != "slot_tracker" || data.OrderCode != "#BTS-4091" {
		t.Fatalf("unexpected tracker identity: %+v", data)
	}
	if data.Headline.Text != "Rider approaching Pune hub" || data.EtaBadge == nil || data.EtaBadge.Label != "14 MINS" {
		t.Fatalf("unexpected headline/eta_badge: %+v", data)
	}
	if len(data.Steps) != 2 || data.Steps[0].Icon == nil || data.Steps[0].Icon.Name != "check_circle" || data.Steps[1].Label != "En Route in Pune" {
		t.Fatalf("unexpected tracker steps: %+v", data.Steps)
	}

	// Empty steps slice serializes as "steps":[].
	emptyEnv, err := reg.Transform(transformer.TransformInput{
		Widget: domain.Widget{
			ID:          "w_tracker_empty",
			Slug:        "empty_tracker",
			SnippetType: snippets.SnippetTypeOrderTracker,
			Active:      true,
			Props:       map[string]any{"headline": "Preparing order"},
		},
	})
	if err != nil {
		t.Fatalf("Transform(empty order_tracker) failed: %v", err)
	}
	emptyJSON, err := json.Marshal(emptyEnv)
	if err != nil {
		t.Fatalf("json.Marshal(empty order_tracker) failed: %v", err)
	}
	if !strings.Contains(string(emptyJSON), `"steps":[]`) {
		t.Fatalf("expected non-nil empty steps array, got: %s", string(emptyJSON))
	}
}

func TestTransformComposedBlock_RecursivePromotionAndNonNilContainerChildren(t *testing.T) {
	t.Parallel()

	reg := transformer.NewDefaultRegistry()
	ctx := domain.RequestContext{City: "Hyderabad", Segment: "vip"}

	env, err := reg.Transform(transformer.TransformInput{
		Slot: domain.LayoutSlot{SlotID: "slot_loyalty_block", Enabled: true},
		Widget: domain.Widget{
			ID:          "w_loyalty",
			SnippetType: snippets.SnippetTypeComposedBlock,
			Active:      true,
			Props: map[string]any{
				"recipe_name": "loyalty_stamp_card",
				"root": map[string]any{
					"kind": snippets.ComposedNodeCard,
					"children": []any{
						map[string]any{
							"kind":       snippets.ComposedNodeColumn,
							"spacing_dp": 8,
							"children": []any{
								map[string]any{"kind": snippets.ComposedNodeText, "text": "Loyalty Card · {{city}}"},
								map[string]any{
									"kind": snippets.ComposedNodeRow,
									"children": []any{
										map[string]any{"kind": snippets.ComposedNodeTag, "tag": "TIER {{segment}}"},
										map[string]any{"kind": snippets.ComposedNodeIcon, "icon": "check_circle"},
										map[string]any{"kind": snippets.ComposedNodeImage, "image": "https://cdn.abeta.dev/fresh/stamp.png"},
									},
								},
								map[string]any{"kind": snippets.ComposedNodeFlowRow},
							},
						},
					},
				},
			},
			Motion: map[string]any{
				"enter":       atoms.MotionTransitionScaleIn,
				"loop_effect": atoms.MotionLoopGlowRing,
			},
		},
		Context: ctx,
	})
	if err != nil {
		t.Fatalf("Transform(composed_block) failed: %v", err)
	}

	// Verify direct Go struct has non-nil Children on empty container FlowRow.
	directData, ok := env.Data.(snippets.ComposedBlockSnippetData)
	if !ok {
		t.Fatalf("expected ComposedBlockSnippetData, got %T", env.Data)
	}
	if directData.BlockID != "slot_loyalty_block" || directData.RecipeName != "loyalty_stamp_card" {
		t.Fatalf("unexpected block metadata: %+v", directData)
	}
	emptyFlow := directData.Root.Children[0].Children[2]
	if emptyFlow.Kind != snippets.ComposedNodeFlowRow || emptyFlow.Children == nil {
		t.Fatalf("expected non-nil Children on empty container flow_row, got: %#v", emptyFlow)
	}

	// Verify full JSON round-trip & recursive promotion + interpolation.
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(composed_block) failed: %v", err)
	}
	var roundTripped snippets.SnippetEnvelope
	if err := json.Unmarshal(raw, &roundTripped); err != nil {
		t.Fatalf("json.Unmarshal(composed_block) failed: %v", err)
	}
	if roundTripped.Motion == nil || roundTripped.Motion.LoopEffect != atoms.MotionLoopGlowRing {
		t.Fatalf("expected loop_effect glow_ring on envelope motion, got %+v", roundTripped.Motion)
	}

	decoded, err := snippets.DecodeSnippetData[snippets.ComposedBlockSnippetData](roundTripped)
	if err != nil {
		t.Fatalf("DecodeSnippetData(composed_block) failed: %v", err)
	}
	col := decoded.Root.Children[0]
	if col.Children[0].Text == nil || col.Children[0].Text.Text != "Loyalty Card · Hyderabad" {
		t.Fatalf("composed text promotion/interpolation failed: %+v", col.Children[0])
	}
	row := col.Children[1]
	if row.Children[0].Tag == nil || row.Children[0].Tag.Label != "TIER vip" {
		t.Fatalf("composed tag promotion/interpolation failed: %+v", row.Children[0])
	}
	if row.Children[1].Icon == nil || row.Children[1].Icon.Name != "check_circle" {
		t.Fatalf("composed icon promotion failed: %+v", row.Children[1])
	}
	if row.Children[2].Image == nil || row.Children[2].Image.URL != "https://cdn.abeta.dev/fresh/stamp.png" {
		t.Fatalf("composed image promotion failed: %+v", row.Children[2])
	}
}
