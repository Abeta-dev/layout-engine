package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

const maxComposedNodeDepth = 6

var validComposedNodeKinds = map[string]struct{}{
	snippets.ComposedNodeColumn:           {},
	snippets.ComposedNodeRow:              {},
	snippets.ComposedNodeBox:              {},
	snippets.ComposedNodeCard:             {},
	snippets.ComposedNodeFlowRow:          {},
	snippets.ComposedNodeScrollRow:        {},
	snippets.ComposedNodeGrid:             {},
	snippets.ComposedNodeText:             {},
	snippets.ComposedNodeButton:           {},
	snippets.ComposedNodeTag:              {},
	snippets.ComposedNodeIcon:             {},
	snippets.ComposedNodeImage:            {},
	snippets.ComposedNodeAvatar:           {},
	snippets.ComposedNodeProgress:         {},
	snippets.ComposedNodeDivider:          {},
	snippets.ComposedNodeSpacer:           {},
	snippets.ComposedNodeStepper:          {},
	snippets.ComposedNodeCountdown:        {},
	snippets.ComposedNodeRollingCounter:   {},
	snippets.ComposedNodeRadioGroup:       {},
	snippets.ComposedNodeSelect:           {},
	snippets.ComposedNodeOtpInput:         {},
	snippets.ComposedNodeChip:             {},
	snippets.ComposedNodeSegmentedControl: {},
	snippets.ComposedNodeRating:           {},
	snippets.ComposedNodePrice:            {},
	snippets.ComposedNodeBadge:            {},
	snippets.ComposedNodeAlertBanner:      {},
	snippets.ComposedNodeStatTile:         {},
	snippets.ComposedNodeKeyValueRow:      {},
	snippets.ComposedNodeQrCode:           {},
	snippets.ComposedNodeTextField:        {},
	snippets.ComposedNodeSwitch:           {},
	snippets.ComposedNodeCheckbox:         {},
	snippets.ComposedNodeSlider:           {},
	snippets.ComposedNodeMediaPager:       {},
}

var validBoxAlignments = map[string]struct{}{
	snippets.BoxAlignTopStart:     {},
	snippets.BoxAlignTopCenter:    {},
	snippets.BoxAlignTopEnd:       {},
	snippets.BoxAlignCenterStart:  {},
	snippets.BoxAlignCenter:       {},
	snippets.BoxAlignCenterEnd:    {},
	snippets.BoxAlignBottomStart:  {},
	snippets.BoxAlignBottomCenter: {},
	snippets.BoxAlignBottomEnd:    {},
}

// IsValidComposedNodeKind returns true if kind is one of the 36 canonical KMP v1 ComposedNodeKinds.
func IsValidComposedNodeKind(kind string) bool {
	_, ok := validComposedNodeKinds[strings.TrimSpace(kind)]
	return ok
}

// SupportedComposedNodeKinds returns a deterministic, sorted copy of all 36 KMP v1 ComposedNodeKinds.
func SupportedComposedNodeKinds() []string {
	out := make([]string, 0, len(validComposedNodeKinds))
	for k := range validComposedNodeKinds {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// IsValidBoxAlignment returns true if alignment is one of the 9 canonical KMP v1 BoxAlign values.
func IsValidBoxAlignment(alignment string) bool {
	_, ok := validBoxAlignments[strings.TrimSpace(alignment)]
	return ok
}

// SupportedBoxAlignments returns a deterministic, sorted copy of all 9 KMP v1 BoxAlign values.
func SupportedBoxAlignments() []string {
	out := make([]string, 0, len(validBoxAlignments))
	for k := range validBoxAlignments {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func (v *DefaultLayoutValidator) validateSwipeCardsProp(
	path string,
	raw any,
	issues *[]domain.ValidationIssue,
) {
	switch cards := raw.(type) {
	case []snippets.SwipeCardItem:
		for i, card := range cards {
			validateTypedSwipeCard(fmt.Sprintf("%s/%d", path, i), card, issues)
		}
	case []map[string]any:
		for i, card := range cards {
			validateMapSwipeCard(fmt.Sprintf("%s/%d", path, i), card, issues)
		}
	case []any:
		for i, elem := range cards {
			cardPath := fmt.Sprintf("%s/%d", path, i)
			switch typed := elem.(type) {
			case snippets.SwipeCardItem:
				validateTypedSwipeCard(cardPath, typed, issues)
			case map[string]any:
				validateMapSwipeCard(cardPath, typed, issues)
			default:
				*issues = append(*issues, domain.ValidationIssue{
					Path:    cardPath,
					Code:    "invalid_swipe_card",
					Message: "swipe card entry must be an object",
				})
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_swipe_cards",
			Message: "swipe_deck cards must be a list of card objects",
		})
	}
}

func validateTypedSwipeCard(path string, card snippets.SwipeCardItem, issues *[]domain.ValidationIssue) {
	if strings.TrimSpace(card.CardID) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/card_id",
			Code:    "required",
			Message: "swipe card requires card_id",
		})
	}
	if strings.TrimSpace(card.Title.Text) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/title",
			Code:    "required",
			Message: "swipe card requires title",
		})
	}
}

func validateMapSwipeCard(path string, m map[string]any, issues *[]domain.ValidationIssue) {
	cardID, _ := m["card_id"].(string)
	if strings.TrimSpace(cardID) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/card_id",
			Code:    "required",
			Message: "swipe card requires card_id",
		})
	}
	if extractAtomText(m["title"]) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/title",
			Code:    "required",
			Message: "swipe card requires title",
		})
	}
}

func (v *DefaultLayoutValidator) validateChipGroupProp(
	path string,
	raw any,
	issues *[]domain.ValidationIssue,
) {
	switch chips := raw.(type) {
	case []snippets.FilterChipItem:
		for i, chip := range chips {
			validateTypedFilterChip(fmt.Sprintf("%s/%d", path, i), chip, issues)
		}
	case []map[string]any:
		for i, chip := range chips {
			validateMapFilterChip(fmt.Sprintf("%s/%d", path, i), chip, issues)
		}
	case []any:
		for i, elem := range chips {
			chipPath := fmt.Sprintf("%s/%d", path, i)
			switch typed := elem.(type) {
			case snippets.FilterChipItem:
				validateTypedFilterChip(chipPath, typed, issues)
			case map[string]any:
				validateMapFilterChip(chipPath, typed, issues)
			default:
				*issues = append(*issues, domain.ValidationIssue{
					Path:    chipPath,
					Code:    "invalid_chip",
					Message: "chip entry must be an object",
				})
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_chips",
			Message: "chip_group chips must be a list of chip objects",
		})
	}
}

func validateTypedFilterChip(path string, chip snippets.FilterChipItem, issues *[]domain.ValidationIssue) {
	if strings.TrimSpace(chip.ID) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/id",
			Code:    "required",
			Message: "chip item requires id",
		})
	}
	if strings.TrimSpace(chip.Label) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/label",
			Code:    "required",
			Message: "chip item requires label",
		})
	}
}

func validateMapFilterChip(path string, m map[string]any, issues *[]domain.ValidationIssue) {
	id, _ := m["id"].(string)
	label, _ := m["label"].(string)
	if strings.TrimSpace(id) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/id",
			Code:    "required",
			Message: "chip item requires id",
		})
	}
	if strings.TrimSpace(label) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/label",
			Code:    "required",
			Message: "chip item requires label",
		})
	}
}

func extractAtomText(raw any) string {
	switch val := raw.(type) {
	case string:
		return strings.TrimSpace(val)
	case map[string]any:
		s, _ := val["text"].(string)
		return strings.TrimSpace(s)
	case atoms.TextAtom:
		return strings.TrimSpace(val.Text)
	case *atoms.TextAtom:
		if val != nil {
			return strings.TrimSpace(val.Text)
		}
	}
	return ""
}

func (v *DefaultLayoutValidator) validateNudgeBarProps(
	path string,
	props map[string]any,
	issues *[]domain.ValidationIssue,
) {
	if extractAtomText(props["title"]) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/title",
			Code:    "required",
			Message: "nudge_bar requires non-empty title",
		})
	}
	if rawProg, exists := props["progress"]; exists && rawProg != nil {
		prog, ok := extractStrictFloat64(rawProg)
		if !ok || math.IsNaN(prog) || math.IsInf(prog, 0) || prog < 0 || prog > 1 {
			*issues = append(*issues, domain.ValidationIssue{
				Path:    path + "/progress",
				Code:    "invalid_progress",
				Message: "nudge_bar progress must be a number between 0 and 1",
			})
		}
	}
}

func extractStrictFloat64(raw any) (float64, bool) {
	switch val := raw.(type) {
	case float64:
		return val, true
	case *float64:
		if val == nil {
			return 0, false
		}
		return *val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func (v *DefaultLayoutValidator) validateMarqueeItemsProp(
	path string,
	raw any,
	issues *[]domain.ValidationIssue,
) {
	switch items := raw.(type) {
	case []snippets.MarqueeItem:
		for i, item := range items {
			if strings.TrimSpace(item.Text) == "" {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    fmt.Sprintf("%s/%d/text", path, i),
					Code:    "required",
					Message: "marquee item requires non-empty text",
				})
			}
		}
	case []map[string]any:
		for i, item := range items {
			validateMapMarqueeItem(fmt.Sprintf("%s/%d", path, i), item, issues)
		}
	case []any:
		for i, elem := range items {
			itemPath := fmt.Sprintf("%s/%d", path, i)
			switch typed := elem.(type) {
			case snippets.MarqueeItem:
				if strings.TrimSpace(typed.Text) == "" {
					*issues = append(*issues, domain.ValidationIssue{
						Path:    itemPath + "/text",
						Code:    "required",
						Message: "marquee item requires non-empty text",
					})
				}
			case map[string]any:
				validateMapMarqueeItem(itemPath, typed, issues)
			default:
				*issues = append(*issues, domain.ValidationIssue{
					Path:    itemPath,
					Code:    "invalid_marquee_item",
					Message: "marquee item entry must be an object",
				})
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_marquee_items",
			Message: "marquee_strip items must be a list of marquee item objects",
		})
	}
}

func validateMapMarqueeItem(path string, m map[string]any, issues *[]domain.ValidationIssue) {
	text, _ := m["text"].(string)
	if strings.TrimSpace(text) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/text",
			Code:    "required",
			Message: "marquee item requires non-empty text",
		})
	}
}

func (v *DefaultLayoutValidator) validateOrderTrackerProps(
	path string,
	props map[string]any,
	issues *[]domain.ValidationIssue,
) {
	if extractAtomText(props["headline"]) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/headline",
			Code:    "required",
			Message: "order_tracker requires non-empty headline",
		})
	}
	rawSteps, exists := props["steps"]
	if !exists || rawSteps == nil {
		return
	}
	stepsPath := path + "/steps"
	switch steps := rawSteps.(type) {
	case []snippets.OrderTrackerStep:
		for i, step := range steps {
			if strings.TrimSpace(step.Label) == "" {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    fmt.Sprintf("%s/%d/label", stepsPath, i),
					Code:    "required",
					Message: "order_tracker step requires non-empty label",
				})
			}
		}
	case []map[string]any:
		for i, step := range steps {
			validateMapOrderTrackerStep(fmt.Sprintf("%s/%d", stepsPath, i), step, issues)
		}
	case []any:
		for i, elem := range steps {
			stepPath := fmt.Sprintf("%s/%d", stepsPath, i)
			switch typed := elem.(type) {
			case snippets.OrderTrackerStep:
				if strings.TrimSpace(typed.Label) == "" {
					*issues = append(*issues, domain.ValidationIssue{
						Path:    stepPath + "/label",
						Code:    "required",
						Message: "order_tracker step requires non-empty label",
					})
				}
			case map[string]any:
				validateMapOrderTrackerStep(stepPath, typed, issues)
			default:
				*issues = append(*issues, domain.ValidationIssue{
					Path:    stepPath,
					Code:    "invalid_order_step",
					Message: "order_tracker step entry must be an object",
				})
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    stepsPath,
			Code:    "invalid_order_steps",
			Message: "order_tracker steps must be a list of step objects",
		})
	}
}

func validateMapOrderTrackerStep(path string, m map[string]any, issues *[]domain.ValidationIssue) {
	label, _ := m["label"].(string)
	if strings.TrimSpace(label) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/label",
			Code:    "required",
			Message: "order_tracker step requires non-empty label",
		})
	}
}

func (v *DefaultLayoutValidator) validateComposedBlockProps(
	path string,
	props map[string]any,
	issues *[]domain.ValidationIssue,
) {
	rawRoot, exists := props["root"]
	if !exists || rawRoot == nil {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/root",
			Code:    "required",
			Message: "composed_block requires a root node object",
		})
		return
	}
	v.validateComposedNode(path+"/root", rawRoot, 0, issues)
}

func (v *DefaultLayoutValidator) validateComposedNode(
	path string,
	raw any,
	depth int,
	issues *[]domain.ValidationIssue,
) {
	if depth > maxComposedNodeDepth {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "max_depth_exceeded",
			Message: fmt.Sprintf("composed_block exceeds maximum nesting depth of %d", maxComposedNodeDepth),
		})
		return
	}

	switch node := raw.(type) {
	case snippets.ComposedNode:
		validateComposedNodeKind(path, node.Kind, issues)
		validateBoxAlignment(path, node.BoxAlignment, issues)
		for i, ch := range node.Children {
			v.validateComposedNode(fmt.Sprintf("%s/children/%d", path, i), ch, depth+1, issues)
		}
	case *snippets.ComposedNode:
		if node == nil {
			*issues = append(*issues, domain.ValidationIssue{
				Path:    path,
				Code:    "invalid_composed_node",
				Message: "composed_block node must be an object",
			})
			return
		}
		validateComposedNodeKind(path, node.Kind, issues)
		validateBoxAlignment(path, node.BoxAlignment, issues)
		for i, ch := range node.Children {
			v.validateComposedNode(fmt.Sprintf("%s/children/%d", path, i), ch, depth+1, issues)
		}
	case map[string]any:
		kind, _ := node["kind"].(string)
		validateComposedNodeKind(path, kind, issues)
		if rawAlign, hasAlign := node["box_alignment"]; hasAlign && rawAlign != nil {
			if alignStr, ok := rawAlign.(string); ok {
				validateBoxAlignment(path, alignStr, issues)
			} else {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    path + "/box_alignment",
					Code:    "invalid_box_alignment",
					Message: "composed_block box_alignment must be a string",
				})
			}
		}
		rawChildren, exists := node["children"]
		if !exists || rawChildren == nil {
			return
		}
		childrenPath := path + "/children"
		switch children := rawChildren.(type) {
		case []snippets.ComposedNode:
			for i, ch := range children {
				v.validateComposedNode(fmt.Sprintf("%s/%d", childrenPath, i), ch, depth+1, issues)
			}
		case []map[string]any:
			for i, ch := range children {
				v.validateComposedNode(fmt.Sprintf("%s/%d", childrenPath, i), ch, depth+1, issues)
			}
		case []any:
			for i, ch := range children {
				v.validateComposedNode(fmt.Sprintf("%s/%d", childrenPath, i), ch, depth+1, issues)
			}
		default:
			*issues = append(*issues, domain.ValidationIssue{
				Path:    childrenPath,
				Code:    "invalid_composed_children",
				Message: "composed_block children must be a list of nodes",
			})
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_composed_node",
			Message: "composed_block node must be an object",
		})
	}
}

func validateComposedNodeKind(path, kind string, issues *[]domain.ValidationIssue) {
	trimmed := strings.TrimSpace(kind)
	if trimmed == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/kind",
			Code:    "required",
			Message: "composed_block node requires kind",
		})
		return
	}
	if _, ok := validComposedNodeKinds[trimmed]; !ok {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/kind",
			Code:    "invalid_composed_node_kind",
			Message: fmt.Sprintf("unsupported composed_block node kind %q", kind),
		})
	}
}

func validateBoxAlignment(path, alignment string, issues *[]domain.ValidationIssue) {
	trimmed := strings.TrimSpace(alignment)
	if trimmed == "" {
		return
	}
	if _, ok := validBoxAlignments[trimmed]; !ok {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/box_alignment",
			Code:    "invalid_box_alignment",
			Message: fmt.Sprintf("unsupported composed_block box_alignment %q", alignment),
		})
	}
}
