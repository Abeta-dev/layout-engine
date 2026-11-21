package engine

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// Compile-time interface verification.
var _ domain.LayoutValidator = (*DefaultLayoutValidator)(nil)

var urlSafeSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

var supportedSnippetTypeList = []string{
	snippets.SnippetTypeAccordion,
	snippets.SnippetTypeActionRow,
	snippets.SnippetTypeBanner,
	snippets.SnippetTypeCard,
	snippets.SnippetTypeCarousel,
	snippets.SnippetTypeDivider,
	snippets.SnippetTypeEmptyState,
	snippets.SnippetTypeForm,
	snippets.SnippetTypeGrid,
	snippets.SnippetTypeHeader,
	snippets.SnippetTypeMetricCard,
	snippets.SnippetTypeProductCard,
	snippets.SnippetTypeSpacer,
	snippets.SnippetTypeTabs,
}

var supportedSnippetTypeSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(supportedSnippetTypeList))
	for _, t := range supportedSnippetTypeList {
		m[t] = struct{}{}
	}
	return m
}()

// IsValidSnippetType returns true if t is one of the 14 supported KMP v1 SnippetTypes.
func IsValidSnippetType(t string) bool {
	_, ok := supportedSnippetTypeSet[t]
	return ok
}

// SupportedSnippetTypes returns a deterministic, sorted copy of all 14 KMP v1 SnippetTypes.
func SupportedSnippetTypes() []string {
	out := make([]string, len(supportedSnippetTypeList))
	copy(out, supportedSnippetTypeList)
	slices.Sort(out)
	return out
}

// DefaultLayoutValidator validates draft layouts, slot bindings, snippet types,
// and nested container structures before publishing.
type DefaultLayoutValidator struct{}

// NewLayoutValidator creates a new DefaultLayoutValidator.
func NewLayoutValidator() *DefaultLayoutValidator {
	return &DefaultLayoutValidator{}
}

// ValidateLayout returns all schema and referential integrity issues found on layout.
func (v *DefaultLayoutValidator) ValidateLayout(
	layout domain.Layout,
	widgets map[string]domain.Widget,
) []domain.ValidationIssue {
	issues := make([]domain.ValidationIssue, 0)

	slug := strings.TrimSpace(layout.Slug)
	if slug == "" {
		issues = append(issues, domain.ValidationIssue{
			Path:    "/slug",
			Code:    "required",
			Message: "layout slug is required",
		})
	} else if !urlSafeSlugPattern.MatchString(slug) {
		issues = append(issues, domain.ValidationIssue{
			Path:    "/slug",
			Code:    "invalid_slug",
			Message: fmt.Sprintf("layout slug %q must be URL-safe (lowercase alphanumeric, hyphens, underscores)", layout.Slug),
		})
	}

	if strings.TrimSpace(layout.Title) == "" {
		issues = append(issues, domain.ValidationIssue{
			Path:    "/title",
			Code:    "required",
			Message: "layout title is required",
		})
	}

	if layout.PageConfig.AdaptiveLayout != nil && layout.PageConfig.AdaptiveLayout.LayoutMode != "" {
		switch layout.PageConfig.AdaptiveLayout.LayoutMode {
		case snippets.LayoutModeSingleColumn, snippets.LayoutModeCentered, snippets.LayoutModeSplitPane:
		default:
			issues = append(issues, domain.ValidationIssue{
				Path:    "/page_config/adaptive_layout/layout_mode",
				Code:    "invalid_layout_mode",
				Message: fmt.Sprintf("unsupported adaptive layout mode %q", layout.PageConfig.AdaptiveLayout.LayoutMode),
			})
		}
	}

	seenSlotIDs := make(map[string]struct{}, len(layout.Slots))
	for i, slot := range layout.Slots {
		slotPath := fmt.Sprintf("/slots/%d", i)
		slotID := strings.TrimSpace(slot.SlotID)
		if slotID == "" {
			issues = append(issues, domain.ValidationIssue{
				Path:    slotPath + "/slot_id",
				Code:    "required",
				Message: "slot_id is required",
			})
		} else if _, dup := seenSlotIDs[slotID]; dup {
			issues = append(issues, domain.ValidationIssue{
				Path:    slotPath + "/slot_id",
				Code:    "duplicate_slot_id",
				Message: fmt.Sprintf("duplicate slot_id %q", slotID),
			})
		} else {
			seenSlotIDs[slotID] = struct{}{}
		}

		if !slot.Enabled {
			continue
		}

		widgetID := strings.TrimSpace(slot.WidgetID)
		if widgetID == "" {
			issues = append(issues, domain.ValidationIssue{
				Path:    slotPath + "/widget_id",
				Code:    "required",
				Message: "enabled slot requires widget_id",
			})
			continue
		}

		widget, ok := widgets[widgetID]
		if !ok {
			issues = append(issues, domain.ValidationIssue{
				Path:    slotPath + "/widget_id",
				Code:    "missing_widget",
				Message: fmt.Sprintf("referenced widget %q does not exist", widgetID),
			})
			continue
		}

		visited := map[string]bool{widget.ID: true}
		v.validateWidgetBinding(slotPath, widget, slot.Overrides, widgets, visited, &issues)
	}

	return issues
}

func (v *DefaultLayoutValidator) validateWidgetBinding(
	path string,
	widget domain.Widget,
	overrides map[string]any,
	widgets map[string]domain.Widget,
	visited map[string]bool,
	issues *[]domain.ValidationIssue,
) {
	if !IsValidSnippetType(widget.SnippetType) {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/snippet_type",
			Code:    "invalid_snippet_type",
			Message: fmt.Sprintf("widget %q has unsupported snippet_type %q", widget.ID, widget.SnippetType),
		})
		return
	}

	props := mergeShallowProps(widget.Props, overrides)

	switch widget.SnippetType {
	case snippets.SnippetTypeCarousel, snippets.SnippetTypeGrid:
		if widget.SnippetType == snippets.SnippetTypeGrid {
			if rawCols, exists := props["columns"]; exists {
				if cols, ok := toFloat64(rawCols); !ok || cols <= 0 {
					*issues = append(*issues, domain.ValidationIssue{
						Path:    path + "/props/columns",
						Code:    "invalid_columns",
						Message: "grid columns must be a positive integer",
					})
				}
			}
		}
		if rawItems, exists := props["items"]; exists && rawItems != nil {
			v.validateInlineEnvelopes(path+"/props/items", rawItems, issues)
		}
		if rawChildIDs, exists := props["child_widget_ids"]; exists && rawChildIDs != nil {
			v.validateChildWidgetRefs(path+"/props/child_widget_ids", rawChildIDs, widgets, visited, issues)
		}

	case snippets.SnippetTypeAccordion:
		if rawSnippets, exists := props["snippets"]; exists && rawSnippets != nil {
			v.validateInlineEnvelopes(path+"/props/snippets", rawSnippets, issues)
		}
		if rawItems, exists := props["items"]; exists && rawItems != nil {
			v.validateInlineEnvelopes(path+"/props/items", rawItems, issues)
		}
		if rawChildIDs, exists := props["child_widget_ids"]; exists && rawChildIDs != nil {
			v.validateChildWidgetRefs(path+"/props/child_widget_ids", rawChildIDs, widgets, visited, issues)
		}

	case snippets.SnippetTypeTabs:
		if rawTabs, exists := props["tabs"]; exists && rawTabs != nil {
			v.validateTabsProp(path+"/props/tabs", rawTabs, widgets, visited, issues)
		}
	}
}

func (v *DefaultLayoutValidator) validateInlineEnvelopes(
	path string,
	raw any,
	issues *[]domain.ValidationIssue,
) {
	switch list := raw.(type) {
	case []snippets.SnippetEnvelope:
		for i, env := range list {
			itemPath := fmt.Sprintf("%s/%d", path, i)
			if strings.TrimSpace(env.SnippetID) == "" {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    itemPath + "/snippet_id",
					Code:    "required",
					Message: "inline snippet requires snippet_id",
				})
			}
			if !IsValidSnippetType(env.SnippetType) {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    itemPath + "/snippet_type",
					Code:    "invalid_snippet_type",
					Message: fmt.Sprintf("inline snippet has unsupported snippet_type %q", env.SnippetType),
				})
			}
		}
	case []any:
		for i, elem := range list {
			itemPath := fmt.Sprintf("%s/%d", path, i)
			switch typed := elem.(type) {
			case snippets.SnippetEnvelope:
				if strings.TrimSpace(typed.SnippetID) == "" {
					*issues = append(*issues, domain.ValidationIssue{
						Path:    itemPath + "/snippet_id",
						Code:    "required",
						Message: "inline snippet requires snippet_id",
					})
				}
				if !IsValidSnippetType(typed.SnippetType) {
					*issues = append(*issues, domain.ValidationIssue{
						Path:    itemPath + "/snippet_type",
						Code:    "invalid_snippet_type",
						Message: fmt.Sprintf("inline snippet has unsupported snippet_type %q", typed.SnippetType),
					})
				}
			case map[string]any:
				sid, _ := typed["snippet_id"].(string)
				stype, _ := typed["snippet_type"].(string)
				if strings.TrimSpace(sid) == "" {
					*issues = append(*issues, domain.ValidationIssue{
						Path:    itemPath + "/snippet_id",
						Code:    "required",
						Message: "inline snippet requires snippet_id",
					})
				}
				if !IsValidSnippetType(stype) {
					*issues = append(*issues, domain.ValidationIssue{
						Path:    itemPath + "/snippet_type",
						Code:    "invalid_snippet_type",
						Message: fmt.Sprintf("inline snippet has unsupported snippet_type %q", stype),
					})
				}
			default:
				*issues = append(*issues, domain.ValidationIssue{
					Path:    itemPath,
					Code:    "invalid_container_item",
					Message: "inline container item must be a snippet object",
				})
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_container_items",
			Message: "container items must be a list of snippets",
		})
	}
}

func (v *DefaultLayoutValidator) validateTabsProp(
	path string,
	raw any,
	widgets map[string]domain.Widget,
	visited map[string]bool,
	issues *[]domain.ValidationIssue,
) {
	switch tabs := raw.(type) {
	case []snippets.TabItemSpec:
		for i, tab := range tabs {
			tabPath := fmt.Sprintf("%s/%d", path, i)
			if strings.TrimSpace(tab.ID) == "" {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    tabPath + "/id",
					Code:    "required",
					Message: "tab item requires id",
				})
			}
			if strings.TrimSpace(tab.Title) == "" {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    tabPath + "/title",
					Code:    "required",
					Message: "tab item requires title",
				})
			}
			if len(tab.Snippets) > 0 {
				v.validateInlineEnvelopes(tabPath+"/snippets", tab.Snippets, issues)
			}
		}
	case []any:
		for i, elem := range tabs {
			tabPath := fmt.Sprintf("%s/%d", path, i)
			m, ok := elem.(map[string]any)
			if !ok {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    tabPath,
					Code:    "invalid_tab",
					Message: "tab entry must be an object",
				})
				continue
			}
			id, _ := m["id"].(string)
			title, _ := m["title"].(string)
			if strings.TrimSpace(id) == "" {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    tabPath + "/id",
					Code:    "required",
					Message: "tab item requires id",
				})
			}
			if strings.TrimSpace(title) == "" {
				*issues = append(*issues, domain.ValidationIssue{
					Path:    tabPath + "/title",
					Code:    "required",
					Message: "tab item requires title",
				})
			}
			if rawSnippets, exists := m["snippets"]; exists && rawSnippets != nil {
				v.validateInlineEnvelopes(tabPath+"/snippets", rawSnippets, issues)
			}
			if rawItems, exists := m["items"]; exists && rawItems != nil {
				v.validateInlineEnvelopes(tabPath+"/items", rawItems, issues)
			}
			if rawWidgetIDs, exists := m["widget_ids"]; exists && rawWidgetIDs != nil {
				v.validateChildWidgetRefs(tabPath+"/widget_ids", rawWidgetIDs, widgets, visited, issues)
			}
			if rawChildIDs, exists := m["child_widget_ids"]; exists && rawChildIDs != nil {
				v.validateChildWidgetRefs(tabPath+"/child_widget_ids", rawChildIDs, widgets, visited, issues)
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_tabs",
			Message: "tabs must be a list of tab definitions",
		})
	}
}

func (v *DefaultLayoutValidator) validateChildWidgetRefs(
	path string,
	raw any,
	widgets map[string]domain.Widget,
	visited map[string]bool,
	issues *[]domain.ValidationIssue,
) {
	ids, ok := extractStringSlice(raw)
	if !ok {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_widget_refs",
			Message: "child widget references must be a list of widget IDs",
		})
		return
	}

	for i, childID := range ids {
		refPath := fmt.Sprintf("%s/%d", path, i)
		if strings.TrimSpace(childID) == "" {
			*issues = append(*issues, domain.ValidationIssue{
				Path:    refPath,
				Code:    "required",
				Message: "child widget ID cannot be empty",
			})
			continue
		}
		if visited[childID] {
			*issues = append(*issues, domain.ValidationIssue{
				Path:    refPath,
				Code:    "cycle_detected",
				Message: fmt.Sprintf("cyclic container reference detected for widget %q", childID),
			})
			continue
		}
		child, exists := widgets[childID]
		if !exists {
			*issues = append(*issues, domain.ValidationIssue{
				Path:    refPath,
				Code:    "missing_child_widget",
				Message: fmt.Sprintf("referenced child widget %q does not exist", childID),
			})
			continue
		}
		nextVisited := make(map[string]bool, len(visited)+1)
		for k, val := range visited {
			nextVisited[k] = val
		}
		nextVisited[childID] = true
		v.validateWidgetBinding(refPath, child, nil, widgets, nextVisited, issues)
	}
}

func extractStringSlice(raw any) ([]string, bool) {
	switch list := raw.(type) {
	case []string:
		return list, true
	case []any:
		out := make([]string, 0, len(list))
		for _, item := range list {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}

func mergeShallowProps(base, overrides map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(overrides))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}
