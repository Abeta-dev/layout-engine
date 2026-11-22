package transformer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// NewDefaultRegistry constructs a Registry pre-populated with transformers for all
// 21 Abeta KMP v1 SnippetTypes.
func NewDefaultRegistry() *Registry {
	reg := NewRegistry()

	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeHeader, transformHeader))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeBanner, transformBanner))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeMetricCard, transformMetricCard))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeCard, transformCard))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeActionRow, transformActionRow))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeForm, transformForm))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeEmptyState, transformEmptyState))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeDivider, transformDivider))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeSpacer, transformSpacer))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeCarousel, transformCarousel))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeGrid, transformGrid))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeTabs, transformTabs))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeAccordion, transformAccordion))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeProductCard, transformProductCard))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeSwipeDeck, transformSwipeDeck))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeSearchBar, transformSearchBar))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeChipGroup, transformChipGroup))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeNudgeBar, transformNudgeBar))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeMarqueeStrip, transformMarqueeStrip))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeOrderTracker, transformOrderTracker))
	reg.Register(NewWidgetTransformerFunc(snippets.SnippetTypeComposedBlock, transformComposedBlock))

	return reg
}

// ── Built-In Leaf Snippet Transformers (10 types) ──────────────────────────────

func transformHeader(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "subtitle")
	promoteIconFields(props, "leading_icon")
	promoteTagFields(props, "trailing_tag")
	return decodeEnvelope[snippets.HeaderSnippetData](input, snippets.SnippetTypeHeader, props)
}

func transformBanner(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "body")
	promoteIconFields(props, "icon")
	promoteImageFields(props, "image")
	promoteTagFields(props, "tag")
	return decodeEnvelope[snippets.BannerSnippetData](input, snippets.SnippetTypeBanner, props)
}

func transformMetricCard(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "label", "value")
	promoteTagFields(props, "delta")
	promoteIconFields(props, "icon")
	return decodeEnvelope[snippets.MetricCardSnippetData](input, snippets.SnippetTypeMetricCard, props)
}

func transformCard(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "subtitle", "description")
	promoteImageFields(props, "media")
	promoteTagFields(props, "badge")
	return decodeEnvelope[snippets.CardSnippetData](input, snippets.SnippetTypeCard, props)
}

func transformActionRow(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "subtitle")
	promoteIconFields(props, "leading_icon", "trailing_icon")
	promoteTagFields(props, "trailing_tag")
	return decodeEnvelope[snippets.ActionRowSnippetData](input, snippets.SnippetTypeActionRow, props)
}

func transformForm(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "description")
	var data snippets.FormSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeForm, err)
	}
	if strings.TrimSpace(data.FormID) == "" {
		data.FormID = resolveSnippetID(input)
	}
	if data.Fields == nil {
		data.Fields = []atoms.TextFieldAtom{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeForm, data), nil
}

func transformEmptyState(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "message")
	promoteIconFields(props, "icon")
	promoteImageFields(props, "image")
	return decodeEnvelope[snippets.EmptyStateSnippetData](input, snippets.SnippetTypeEmptyState, props)
}

func transformDivider(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	if _, hasNested := props["divider"]; !hasNested {
		inner := map[string]any{}
		for _, k := range []string{"thickness_dp", "color", "inset_start_dp"} {
			if v, ok := props[k]; ok {
				inner[k] = v
			}
		}
		props["divider"] = inner
	}
	return decodeEnvelope[snippets.DividerSnippetData](input, snippets.SnippetTypeDivider, props)
}

func transformSpacer(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	if _, hasNested := props["spacer"]; !hasNested {
		inner := map[string]any{}
		for _, k := range []string{"width_dp", "height_dp"} {
			if v, ok := props[k]; ok {
				inner[k] = v
			}
		}
		props["spacer"] = inner
	}
	return decodeEnvelope[snippets.SpacerSnippetData](input, snippets.SnippetTypeSpacer, props)
}

func transformProductCard(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "subtitle", "price", "original_price", "rating")
	promoteTagFields(props, "badge")
	promoteImageFields(props, "image")

	var data snippets.ProductCardSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeProductCard, err)
	}
	if strings.TrimSpace(data.ProductID) == "" {
		data.ProductID = resolveSnippetID(input)
	}
	if strings.TrimSpace(data.Stepper.FieldID) == "" {
		data.Stepper.FieldID = "qty_" + data.ProductID
	}
	if strings.TrimSpace(data.Stepper.Label) == "" {
		data.Stepper.Label = "ADD"
	}
	return wrapEnvelope(input, snippets.SnippetTypeProductCard, data), nil
}

// ── Built-In Container Snippet Transformers (4 types) ──────────────────────────

func transformCarousel(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title")

	var data snippets.CarouselSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeCarousel, err)
	}

	parentID := resolveSnippetID(input)
	children, err := resolveChildWidgets(input, parentID, props["child_widget_ids"])
	if err != nil {
		return snippets.SnippetEnvelope{}, err
	}
	data.Items = append(data.Items, children...)
	if data.Items == nil {
		data.Items = []snippets.SnippetEnvelope{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeCarousel, data), nil
}

func transformGrid(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title")

	var data snippets.GridSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeGrid, err)
	}
	if data.Columns <= 0 {
		data.Columns = 2
	}

	parentID := resolveSnippetID(input)
	children, err := resolveChildWidgets(input, parentID, props["child_widget_ids"])
	if err != nil {
		return snippets.SnippetEnvelope{}, err
	}
	data.Items = append(data.Items, children...)
	if data.Items == nil {
		data.Items = []snippets.SnippetEnvelope{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeGrid, data), nil
}

func transformAccordion(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title", "subtitle")
	if _, hasSnippets := props["snippets"]; !hasSnippets {
		if items, hasItems := props["items"]; hasItems {
			props["snippets"] = items
		}
	}

	var data snippets.AccordionSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeAccordion, err)
	}

	parentID := resolveSnippetID(input)
	children, err := resolveChildWidgets(input, parentID, props["child_widget_ids"])
	if err != nil {
		return snippets.SnippetEnvelope{}, err
	}
	data.Snippets = append(data.Snippets, children...)
	if data.Snippets == nil {
		data.Snippets = []snippets.SnippetEnvelope{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeAccordion, data), nil
}

type rawTabSpec struct {
	ID             string                     `json:"id"`
	Title          any                        `json:"title"`
	Badge          any                        `json:"badge,omitempty"`
	Snippets       []snippets.SnippetEnvelope `json:"snippets,omitempty"`
	Items          []snippets.SnippetEnvelope `json:"items,omitempty"`
	WidgetIDs      []string                   `json:"widget_ids,omitempty"`
	ChildWidgetIDs []string                   `json:"child_widget_ids,omitempty"`
}

type rawTabsData struct {
	Tabs          []rawTabSpec        `json:"tabs"`
	SelectedIndex int                 `json:"selected_index"`
	Layout        *atoms.LayoutConfig `json:"layout,omitempty"`
}

func transformTabs(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)

	var raw rawTabsData
	if err := decodeMapToStruct(props, &raw); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeTabs, err)
	}

	parentID := resolveSnippetID(input)
	resolvedTabs := make([]snippets.TabItemSpec, 0, len(raw.Tabs))
	for _, rt := range raw.Tabs {
		tab := snippets.TabItemSpec{
			ID:       rt.ID,
			Title:    extractTabTitle(rt.Title),
			Snippets: []snippets.SnippetEnvelope{},
		}
		if rt.Badge != nil {
			badge, err := decodeTag(rt.Badge)
			if err != nil {
				return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode tab %q badge: %w", rt.ID, err)
			}
			tab.Badge = badge
		}

		if len(rt.Snippets) > 0 {
			tab.Snippets = append(tab.Snippets, rt.Snippets...)
		}
		if len(rt.Items) > 0 {
			tab.Snippets = append(tab.Snippets, rt.Items...)
		}

		childIDs := append(append([]string(nil), rt.WidgetIDs...), rt.ChildWidgetIDs...)
		if len(childIDs) > 0 {
			children, err := resolveChildWidgets(input, parentID, childIDs)
			if err != nil {
				return snippets.SnippetEnvelope{}, err
			}
			tab.Snippets = append(tab.Snippets, children...)
		}

		resolvedTabs = append(resolvedTabs, tab)
	}

	selectedIdx := raw.SelectedIndex
	if selectedIdx < 0 {
		selectedIdx = 0
	}
	if len(resolvedTabs) > 0 && selectedIdx >= len(resolvedTabs) {
		selectedIdx = len(resolvedTabs) - 1
	}

	data := snippets.TabsSnippetData{
		Tabs:          resolvedTabs,
		SelectedIndex: selectedIdx,
		Layout:        raw.Layout,
	}
	return wrapEnvelope(input, snippets.SnippetTypeTabs, data), nil
}

// ── Built-In Interactive & Discovery Snippet Transformers (3 types) ────────────

func transformSwipeDeck(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "empty_title", "empty_message")

	promoteCard := func(cardMap map[string]any) {
		promoteTextFields(cardMap, "title", "subtitle", "bio", "distance_text")
		promoteTagFields(cardMap, "badge")
		promoteImageFields(cardMap, "image")
		promoteTagSliceField(cardMap, "tags")
	}

	switch cards := props["cards"].(type) {
	case []any:
		for _, elem := range cards {
			if cardMap, ok := elem.(map[string]any); ok {
				promoteCard(cardMap)
			}
		}
	case []map[string]any:
		for _, cardMap := range cards {
			promoteCard(cardMap)
		}
	}

	var data snippets.SwipeDeckSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeSwipeDeck, err)
	}
	if strings.TrimSpace(data.DeckID) == "" {
		data.DeckID = resolveSnippetID(input)
	}
	if data.Cards == nil {
		data.Cards = []snippets.SwipeCardItem{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeSwipeDeck, data), nil
}

func transformSearchBar(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteIconFields(props, "leading_icon", "trailing_icon")

	var data snippets.SearchBarSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeSearchBar, err)
	}
	if strings.TrimSpace(data.SearchID) == "" {
		data.SearchID = resolveSnippetID(input)
	}
	if strings.TrimSpace(data.Placeholder) == "" {
		data.Placeholder = "Search..."
	}
	if raw, exists := props["clearable"]; !exists || raw == nil {
		data.Clearable = true
	}
	return wrapEnvelope(input, snippets.SnippetTypeSearchBar, data), nil
}

func transformChipGroup(input TransformInput) (snippets.SnippetEnvelope, error) {
	props := cloneProps(input.Props)
	promoteTextFields(props, "title")

	switch chips := props["chips"].(type) {
	case []any:
		for _, elem := range chips {
			if chipMap, ok := elem.(map[string]any); ok {
				promoteIconFields(chipMap, "icon")
			}
		}
	case []map[string]any:
		for _, chipMap := range chips {
			promoteIconFields(chipMap, "icon")
		}
	}

	var data snippets.ChipGroupSnippetData
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippets.SnippetTypeChipGroup, err)
	}
	if strings.TrimSpace(data.GroupID) == "" {
		data.GroupID = resolveSnippetID(input)
	}
	if data.Chips == nil {
		data.Chips = []snippets.FilterChipItem{}
	}
	return wrapEnvelope(input, snippets.SnippetTypeChipGroup, data), nil
}

// ── Recursive Child Resolution & Envelope Decoding ─────────────────────────────

func resolveChildWidgets(
	input TransformInput,
	parentID string,
	rawIDs any,
) ([]snippets.SnippetEnvelope, error) {
	ids, ok := toStringSlice(rawIDs)
	if !ok || len(ids) == 0 {
		return nil, nil
	}

	reg := input.Registry
	if reg == nil {
		reg = NewDefaultRegistry()
	}

	baseVisited := make(map[string]bool, len(input.visited)+1)
	for k, v := range input.visited {
		baseVisited[k] = v
	}
	if input.Widget.ID != "" {
		baseVisited[input.Widget.ID] = true
	}

	out := make([]snippets.SnippetEnvelope, 0, len(ids))
	for _, childID := range ids {
		id := strings.TrimSpace(childID)
		if id == "" {
			continue
		}
		if baseVisited[id] {
			return nil, fmt.Errorf("transformer: cyclic widget reference detected for %q", id)
		}
		childWidget, exists := input.Widgets[id]
		if !exists || !childWidget.Active {
			continue
		}

		branchVisited := make(map[string]bool, len(baseVisited)+1)
		for k, v := range baseVisited {
			branchVisited[k] = v
		}
		branchVisited[id] = true

		childSlotID := strings.TrimSpace(childWidget.Slug)
		if childSlotID == "" {
			childSlotID = childWidget.ID
		}

		childEnv, err := reg.Transform(TransformInput{
			Slot: domain.LayoutSlot{
				SlotID:   childSlotID,
				WidgetID: childWidget.ID,
				Enabled:  true,
			},
			Widget:   childWidget,
			Props:    MergeAndInterpolateProps(childWidget.Props, nil, input.Context),
			Widgets:  input.Widgets,
			Context:  input.Context,
			Registry: reg,
			parentID: parentID,
			visited:  branchVisited,
		})
		if err != nil {
			return nil, fmt.Errorf("transformer: child widget %q inside %q: %w", id, parentID, err)
		}
		out = append(out, childEnv)
	}
	return out, nil
}

func decodeEnvelope[T any](
	input TransformInput,
	snippetType string,
	props map[string]any,
) (snippets.SnippetEnvelope, error) {
	var data T
	if err := decodeMapToStruct(props, &data); err != nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: decode %s: %w", snippetType, err)
	}
	return wrapEnvelope(input, snippetType, data), nil
}

func wrapEnvelope(input TransformInput, snippetType string, data any) snippets.SnippetEnvelope {
	return snippets.SnippetEnvelope{
		SnippetID:       resolveSnippetID(input),
		SnippetParentID: input.parentID,
		SnippetType:     snippetType,
		Data:            data,
	}
}

func decodeMapToStruct(props map[string]any, dst any) error {
	raw, err := json.Marshal(props)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}
