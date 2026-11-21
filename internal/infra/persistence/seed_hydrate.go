package persistence

import (
	"encoding/json"
	"strings"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// collectFrozenWidgets recursively captures every widget referenced by layout slots,
// container child_widget_ids, and tabs[*].widget_ids so snapshots stay self-contained.
func collectFrozenWidgets(slots []domain.LayoutSlot, all map[string]domain.Widget) map[string]domain.Widget {
	frozen := make(map[string]domain.Widget, len(slots))
	var visit func(id string)
	visit = func(id string) {
		if id == "" {
			return
		}
		if _, ok := frozen[id]; ok {
			return
		}
		w, ok := all[id]
		if !ok {
			return
		}
		frozen[id] = w

		for _, cid := range extractSeedIDs(w.Props["child_widget_ids"]) {
			visit(cid)
		}
		if rawTabs, ok := w.Props["tabs"].([]any); ok {
			for _, tabRaw := range rawTabs {
				if tabMap, ok := tabRaw.(map[string]any); ok {
					for _, cid := range extractSeedIDs(tabMap["widget_ids"]) {
						visit(cid)
					}
				}
			}
		}
		if typedTabs, ok := w.Props["tabs"].([]map[string]any); ok {
			for _, tabMap := range typedTabs {
				for _, cid := range extractSeedIDs(tabMap["widget_ids"]) {
					visit(cid)
				}
			}
		}
	}
	for _, sl := range slots {
		visit(sl.WidgetID)
	}
	return frozen
}

func extractSeedIDs(raw any) []string {
	switch list := raw.(type) {
	case []string:
		return list
	case []any:
		out := make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// composeSeedFallbackPage builds a fully hydrated, KMP-v1-valid fallback PageResponse
// for each seeded layout so fallback serving works immediately out of the box.
func composeSeedFallbackPage(layout domain.Layout, widgets map[string]domain.Widget) snippets.PageResponse {
	envs := make([]snippets.SnippetEnvelope, 0, len(layout.Slots))
	for _, sl := range layout.Slots {
		if !sl.Enabled {
			continue
		}
		w, ok := widgets[sl.WidgetID]
		if !ok {
			continue
		}
		envs = append(envs, hydrateSeedEnvelope(w, "", widgets, map[string]bool{}))
	}

	toolbar := cloneToolbarReplaceCity(layout.PageConfig.Toolbar, "Bengaluru")
	title := strings.ReplaceAll(layout.Title, "{{city}}", "Bengaluru")

	return snippets.PageResponse{
		PageID:         layout.Slug,
		Title:          title,
		Toolbar:        toolbar,
		Snippets:       envs,
		BottomBar:      layout.PageConfig.BottomBar,
		FloatingCart:   layout.PageConfig.FloatingCart,
		PageActions:    layout.PageConfig.PageActions,
		Pagination:     layout.PageConfig.Pagination,
		PullToRefresh:  layout.PageConfig.PullToRefresh,
		AdaptiveLayout: layout.PageConfig.AdaptiveLayout,
		Tracking: &atoms.TrackingMeta{
			EventName:  "page_loaded",
			Properties: map[string]any{"page_id": layout.Slug, "fallback": "true"},
		},
	}
}

func cloneToolbarReplaceCity(tb *snippets.ToolbarSpec, city string) *snippets.ToolbarSpec {
	if tb == nil {
		return nil
	}
	raw, err := json.Marshal(tb)
	if err != nil {
		return tb
	}
	replaced := strings.ReplaceAll(string(raw), "{{city}}", city)
	var out snippets.ToolbarSpec
	if err := json.Unmarshal([]byte(replaced), &out); err != nil {
		return tb
	}
	return &out
}

func hydrateSeedEnvelope(
	w domain.Widget,
	parentID string,
	widgets map[string]domain.Widget,
	visited map[string]bool,
) snippets.SnippetEnvelope {
	nextVisited := make(map[string]bool, len(visited)+1)
	for k, v := range visited {
		nextVisited[k] = v
	}
	nextVisited[w.ID] = true

	data := cloneAndReplaceCity(w.Props, "Bengaluru")

	switch w.SnippetType {
	case snippets.SnippetTypeGrid, snippets.SnippetTypeCarousel:
		childIDs := extractSeedIDs(data["child_widget_ids"])
		if len(childIDs) > 0 {
			items := make([]snippets.SnippetEnvelope, 0, len(childIDs))
			for _, cid := range childIDs {
				if nextVisited[cid] {
					continue
				}
				if child, ok := widgets[cid]; ok {
					items = append(items, hydrateSeedEnvelope(child, w.ID, widgets, nextVisited))
				}
			}
			data["items"] = items
			delete(data, "child_widget_ids")
		}
	case snippets.SnippetTypeTabs:
		if rawTabs, ok := data["tabs"].([]any); ok {
			hydratedTabs := make([]map[string]any, 0, len(rawTabs))
			for _, tabRaw := range rawTabs {
				tabMap, ok := tabRaw.(map[string]any)
				if !ok {
					continue
				}
				tabCopy := make(map[string]any, len(tabMap))
				for k, v := range tabMap {
					tabCopy[k] = v
				}
				cids := extractSeedIDs(tabCopy["widget_ids"])
				tabSnippets := make([]snippets.SnippetEnvelope, 0, len(cids))
				for _, cid := range cids {
					if nextVisited[cid] {
						continue
					}
					if child, ok := widgets[cid]; ok {
						tabSnippets = append(tabSnippets, hydrateSeedEnvelope(child, w.ID, widgets, nextVisited))
					}
				}
				tabCopy["snippets"] = tabSnippets
				delete(tabCopy, "widget_ids")
				hydratedTabs = append(hydratedTabs, tabCopy)
			}
			data["tabs"] = hydratedTabs
		}
	}

	var tracking *atoms.TrackingMeta
	if len(w.Tracking) > 0 {
		raw, err := json.Marshal(w.Tracking)
		if err == nil {
			var tm atoms.TrackingMeta
			if json.Unmarshal(raw, &tm) == nil && tm.EventName != "" {
				tracking = &tm
			}
		}
	}

	var motion *atoms.MotionSpec
	if len(w.Motion) > 0 {
		raw, err := json.Marshal(w.Motion)
		if err == nil {
			var ms atoms.MotionSpec
			if json.Unmarshal(raw, &ms) == nil {
				motion = &ms
			}
		}
	}

	return snippets.SnippetEnvelope{
		SnippetID:       w.ID,
		SnippetParentID: parentID,
		SnippetType:     w.SnippetType,
		Data:            data,
		Tracking:        tracking,
		Motion:          motion,
	}
}

func cloneAndReplaceCity(props map[string]any, city string) map[string]any {
	raw, err := json.Marshal(props)
	if err != nil {
		return map[string]any{}
	}
	replaced := strings.ReplaceAll(string(raw), "{{city}}", city)
	var out map[string]any
	if err := json.Unmarshal([]byte(replaced), &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}
