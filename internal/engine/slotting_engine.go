package engine

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/abeta-dev/layout-engine/internal/domain"
)

// Compile-time interface verification.
var _ domain.SlottingEngine = (*DefaultSlottingEngine)(nil)

// DefaultSlottingEngine merges dynamic slotting rules onto a layout's base slot list
// while respecting rule priorities, schedules, pinned slots, and widget uniqueness.
type DefaultSlottingEngine struct {
	eval domain.RuleEvaluator
}

// NewSlottingEngine creates a DefaultSlottingEngine backed by eval.
// If eval is nil, a DefaultRuleEvaluator is used.
func NewSlottingEngine(eval domain.RuleEvaluator) *DefaultSlottingEngine {
	if eval == nil {
		eval = NewRuleEvaluator()
	}
	return &DefaultSlottingEngine{eval: eval}
}

// Apply merges matching slotting rules onto base slots, deduplicates WidgetIDs so the same
// widget never appears twice, re-indexes sequential positions 0..N-1, and returns the
// merged slot list alongside the applied slotting rule IDs in application order.
func (e *DefaultSlottingEngine) Apply(
	base []domain.LayoutSlot,
	rules []domain.SlottingRule,
	ruleLookup map[string]domain.TargetingRule,
	ctx domain.RequestContext,
) ([]domain.LayoutSlot, []string) {
	eval := e.eval
	if eval == nil {
		eval = NewRuleEvaluator()
	}

	now := ctx.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	// Copy and stably sort base slots by Position ascending.
	merged := make([]domain.LayoutSlot, len(base))
	copy(merged, base)
	slices.SortStableFunc(merged, func(a, b domain.LayoutSlot) int {
		return cmp.Compare(a.Position, b.Position)
	})

	// Copy and filter enabled candidate slotting rules with non-empty WidgetID.
	candidates := make([]domain.SlottingRule, 0, len(rules))
	for _, r := range rules {
		if !r.Enabled || strings.TrimSpace(r.WidgetID) == "" {
			continue
		}
		candidates = append(candidates, r)
	}

	// Sort candidates by Priority descending, tie-breaking by ID ascending.
	slices.SortStableFunc(candidates, func(a, b domain.SlottingRule) int {
		if c := cmp.Compare(b.Priority, a.Priority); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})

	appliedRuleIDs := make([]string, 0, len(candidates))
	slottedWidgets := make(map[string]struct{}, len(candidates))

	for _, r := range candidates {
		if !eval.MatchesSchedule(r.Schedule, now) {
			continue
		}
		if strings.TrimSpace(r.RuleID) != "" {
			targetRule, ok := ruleLookup[r.RuleID]
			if !ok || !eval.MatchesRule(targetRule, ctx) {
				continue
			}
		}

		// Higher-priority slotting rules retain ownership of a WidgetID.
		if _, alreadySlotted := slottedWidgets[r.WidgetID]; alreadySlotted {
			continue
		}
		// Never duplicate a widget that is already locked in a pinned base slot.
		if hasPinnedWidget(merged, r.WidgetID) {
			continue
		}

		injected := domain.LayoutSlot{
			SlotID:   "slotting:" + r.ID,
			WidgetID: r.WidgetID,
			Position: r.TargetPosition,
			RuleID:   r.RuleID,
			Schedule: r.Schedule,
			Enabled:  true,
		}

		var insertedIdx int
		switch r.Action {
		case domain.SlottingActionPrepend:
			insertedIdx = 0
			merged = insertSlotAt(merged, insertedIdx, injected)
		case domain.SlottingActionAppend:
			insertedIdx = len(merged)
			merged = append(merged, injected)
		case domain.SlottingActionInsert:
			insertedIdx = clampIndex(r.TargetPosition, 0, len(merged))
			merged = insertSlotAt(merged, insertedIdx, injected)
		case domain.SlottingActionReplace:
			if len(merged) == 0 {
				insertedIdx = 0
				merged = append(merged, injected)
			} else if r.TargetPosition >= len(merged) {
				insertedIdx = len(merged)
				merged = append(merged, injected)
			} else {
				idx := clampIndex(r.TargetPosition, 0, len(merged)-1)
				if merged[idx].Pinned {
					// Respect pinned slot: keep pinned slot intact and insert right after it.
					insertedIdx = idx + 1
					merged = insertSlotAt(merged, insertedIdx, injected)
				} else {
					insertedIdx = idx
					merged[idx] = injected
				}
			}
		default:
			continue
		}

		// Remove any prior unpinned duplicate of this WidgetID so the slotting rule's
		// placement takes effect without rendering the same widget twice.
		merged = removeOtherDuplicates(merged, r.WidgetID, insertedIdx)
		slottedWidgets[r.WidgetID] = struct{}{}
		appliedRuleIDs = append(appliedRuleIDs, r.ID)
	}

	// Final deduplication pass for any duplicates that existed in the base slots.
	deduped := make([]domain.LayoutSlot, 0, len(merged))
	seen := make(map[string]struct{}, len(merged))
	for _, slot := range merged {
		if slot.WidgetID != "" {
			if _, exists := seen[slot.WidgetID]; exists {
				continue
			}
			seen[slot.WidgetID] = struct{}{}
		}
		slot.Position = len(deduped)
		deduped = append(deduped, slot)
	}

	return deduped, appliedRuleIDs
}

func hasPinnedWidget(slots []domain.LayoutSlot, widgetID string) bool {
	for _, s := range slots {
		if s.Pinned && s.WidgetID == widgetID {
			return true
		}
	}
	return false
}

func insertSlotAt(slots []domain.LayoutSlot, idx int, slot domain.LayoutSlot) []domain.LayoutSlot {
	slots = append(slots, domain.LayoutSlot{})
	copy(slots[idx+1:], slots[idx:])
	slots[idx] = slot
	return slots
}

func removeOtherDuplicates(slots []domain.LayoutSlot, widgetID string, keepIdx int) []domain.LayoutSlot {
	out := make([]domain.LayoutSlot, 0, len(slots))
	for i, s := range slots {
		if i != keepIdx && s.WidgetID == widgetID && !s.Pinned {
			continue
		}
		out = append(out, s)
	}
	return out
}

func clampIndex(val, minVal, maxVal int) int {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}
