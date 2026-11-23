package transformer

import (
	"fmt"
	"strings"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
)

// ── Nudges, Tickers, Order Tracking & Atomic Molding Transformers (4 types) ────

func transformNudgeBar(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "subtitle")
	promoteIconFields(props, "icon")
	promoteTagFields(props, "badge")

	var data snippets.NudgeBarSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeNudgeBar, err)
	}
	if strings.TrimSpace(data.NudgeID) == "" {
		data.NudgeID = resolveSnippetID(input)
	}
	if raw, exists := props["dismissible"]; !exists || raw == nil {
		data.Dismissible = true
	}
	return wrapEnvelope(input, snippets.SnippetTypeNudgeBar, data), nil
}

func transformMarqueeStrip(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteMarqueeItem := func(itemMap map[string]any) {
		promoteIconFields(itemMap, "icon")
		promoteTagFields(itemMap, "badge")
	}

	switch items := props["items"].(type) {
	case []any:
		for _, elem := range items {
			if itemMap, ok := elem.(map[string]any); ok {
				promoteMarqueeItem(itemMap)
			}
		}
	case []map[string]any:
		for _, itemMap := range items {
			promoteMarqueeItem(itemMap)
		}
	}

	var data snippets.MarqueeStripSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeMarqueeStrip, err)
	}
	if strings.TrimSpace(data.StripID) == "" {
		data.StripID = resolveSnippetID(input)
	}
	if data.VelocityDpPerSec <= 0 {
		data.VelocityDpPerSec = 48
	}
	if strings.TrimSpace(data.Separator) == "" {
		data.Separator = "•"
	}
	if data.Items == nil {
		data.Items = []snippets.MarqueeItem{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeMarqueeStrip, data), nil
}

func transformOrderTracker(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "headline")
	promoteTagFields(props, "eta_badge")

	switch steps := props["steps"].(type) {
	case []any:
		for _, elem := range steps {
			if stepMap, ok := elem.(map[string]any); ok {
				promoteIconFields(stepMap, "icon")
			}
		}
	case []map[string]any:
		for _, stepMap := range steps {
			promoteIconFields(stepMap, "icon")
		}
	}

	var data snippets.OrderTrackerSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeOrderTracker, err)
	}
	if strings.TrimSpace(data.TrackerID) == "" {
		data.TrackerID = resolveSnippetID(input)
	}
	if data.Steps == nil {
		data.Steps = []snippets.OrderTrackerStep{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeOrderTracker, data), nil
}

func transformComposedBlock(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	if rootMap, ok := props["root"].(map[string]any); ok {
		promoteComposedNodeMap(rootMap)
	}

	var data snippets.ComposedBlockSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeComposedBlock, err)
	}
	if strings.TrimSpace(data.BlockID) == "" {
		data.BlockID = resolveSnippetID(input)
	}
	ensureComposedContainerChildren(&data.Root)
	return wrapEnvelope(input, snippets.SnippetTypeComposedBlock, data), nil
}

func promoteComposedNodeMap(nodeMap map[string]any) {
	if nodeMap == nil {
		return
	}
	promoteTextFields(nodeMap, "text")
	promoteTagFields(nodeMap, "tag")
	promoteIconFields(nodeMap, "icon")
	promoteImageFields(nodeMap, "image")
	promoteExtendedComposedNodeAtoms(nodeMap)

	switch children := nodeMap["children"].(type) {
	case []any:
		for _, elem := range children {
			if chMap, ok := elem.(map[string]any); ok {
				promoteComposedNodeMap(chMap)
			}
		}
	case []map[string]any:
		for _, chMap := range children {
			promoteComposedNodeMap(chMap)
		}
	}
}

func promoteExtendedComposedNodeAtoms(nodeMap map[string]any) {
	if s, ok := nodeMap["chip"].(string); ok {
		nodeMap["chip"] = map[string]any{"id": s, "label": s}
	} else if chipMap, ok := nodeMap["chip"].(map[string]any); ok {
		promoteIconFields(chipMap, "leading_icon")
	}

	if s, ok := nodeMap["alert_banner"].(string); ok {
		nodeMap["alert_banner"] = map[string]any{"title": s}
	} else if alertMap, ok := nodeMap["alert_banner"].(map[string]any); ok {
		promoteIconFields(alertMap, "icon")
	}

	if statMap, ok := nodeMap["stat_tile"].(map[string]any); ok {
		promoteIconFields(statMap, "icon")
	}

	if kvMap, ok := nodeMap["key_value_row"].(map[string]any); ok {
		promoteColorRefFields(kvMap, "value_color")
	}

	if s, ok := nodeMap["price"].(string); ok {
		nodeMap["price"] = map[string]any{"amount_text": s}
	}
	if s, ok := nodeMap["badge"].(string); ok {
		nodeMap["badge"] = map[string]any{"label": s}
	}
	if s, ok := nodeMap["qr_code"].(string); ok {
		nodeMap["qr_code"] = map[string]any{"payload": s}
	}
}

func isComposedContainerKind(kind string) bool {
	switch kind {
	case snippets.ComposedNodeColumn,
		snippets.ComposedNodeRow,
		snippets.ComposedNodeBox,
		snippets.ComposedNodeCard,
		snippets.ComposedNodeFlowRow:
		return true
	default:
		return false
	}
}

func ensureComposedContainerChildren(node *snippets.ComposedNode) {
	if node == nil {
		return
	}
	if isComposedContainerKind(node.Kind) && node.Children == nil {
		node.Children = []snippets.ComposedNode{}
	}
	for i := range node.Children {
		ensureComposedContainerChildren(&node.Children[i])
	}
}
