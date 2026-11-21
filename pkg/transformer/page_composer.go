package transformer

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// Compile-time interface verification.
var _ domain.PageComposer = (*DefaultPageComposer)(nil)

// DefaultPageComposer composes ordered LayoutSlots and CMS Widgets into a v1 KMP
// PageResponse using a WidgetTransformer Registry.
type DefaultPageComposer struct {
	registry *Registry
}

// NewPageComposer creates a DefaultPageComposer backed by registry.
// If registry is nil, NewDefaultRegistry() is used.
func NewPageComposer(registry *Registry) *DefaultPageComposer {
	if registry == nil {
		registry = NewDefaultRegistry()
	}
	return &DefaultPageComposer{registry: registry}
}

// Compose resolves enabled slots into KMP v1 SnippetEnvelopes and builds a complete
// PageResponse with interpolated page chrome, actions, and tracking metadata.
// Missing or inactive widgets are skipped cleanly; if enabled slots were configured
// and none resolved to an active widget, an error is returned.
func (c *DefaultPageComposer) Compose(
	layoutSlug string,
	title string,
	pageConfig domain.LayoutPageConfig,
	slots []domain.LayoutSlot,
	widgets map[string]domain.Widget,
	ctx domain.RequestContext,
) (snippets.PageResponse, error) {
	reg := c.registry
	if reg == nil {
		reg = NewDefaultRegistry()
	}

	resolved := make([]snippets.SnippetEnvelope, 0, len(slots))
	enabledCount := 0

	for _, slot := range slots {
		if !slot.Enabled {
			continue
		}
		enabledCount++

		widget, ok := widgets[slot.WidgetID]
		if !ok || !widget.Active {
			continue
		}

		mergedProps := MergeAndInterpolateProps(widget.Props, slot.Overrides, ctx)
		env, err := reg.Transform(TransformInput{
			Slot:     slot,
			Widget:   widget,
			Props:    mergedProps,
			Widgets:  widgets,
			Context:  ctx,
			Registry: reg,
		})
		if err != nil {
			return snippets.PageResponse{}, fmt.Errorf(
				"page composer: transform slot %q (widget %q): %w",
				slot.SlotID,
				slot.WidgetID,
				err,
			)
		}
		resolved = append(resolved, env)
	}

	if enabledCount > 0 && len(resolved) == 0 {
		return snippets.PageResponse{}, fmt.Errorf(
			"page composer: layout %q had %d enabled slot(s) but none resolved to an active widget",
			layoutSlug,
			enabledCount,
		)
	}

	page := snippets.PageResponse{
		PageID:         layoutSlug,
		Title:          InterpolateString(title, ctx),
		Toolbar:        interpolateStructPtr(pageConfig.Toolbar, ctx),
		Snippets:       resolved,
		BottomBar:      interpolateStructPtr(pageConfig.BottomBar, ctx),
		FloatingCart:   interpolateStructPtr(pageConfig.FloatingCart, ctx),
		AdaptiveLayout: cloneAdaptiveSpec(pageConfig.AdaptiveLayout),
		Pagination:     interpolateStructPtr(pageConfig.Pagination, ctx),
		PullToRefresh:  interpolateStructPtr(pageConfig.PullToRefresh, ctx),
		PageActions:    interpolateActions(pageConfig.PageActions, ctx),
		Tracking:       buildPageTracking(layoutSlug, len(resolved), ctx),
	}

	return page, nil
}

func interpolateStructPtr[T any](src *T, ctx domain.RequestContext) *T {
	if src == nil {
		return nil
	}
	raw, err := json.Marshal(src)
	if err != nil {
		cp := *src
		return &cp
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		cp := *src
		return &cp
	}
	interpolated := interpolateValue(generic, ctx)
	updatedRaw, err := json.Marshal(interpolated)
	if err != nil {
		cp := *src
		return &cp
	}
	var out T
	if err := json.Unmarshal(updatedRaw, &out); err != nil {
		cp := *src
		return &cp
	}
	return &out
}

func interpolateActions(src []actions.ActionTrigger, ctx domain.RequestContext) []actions.ActionTrigger {
	if len(src) == 0 {
		return nil
	}
	out := make([]actions.ActionTrigger, 0, len(src))
	for i := range src {
		if cloned := interpolateStructPtr(&src[i], ctx); cloned != nil {
			out = append(out, *cloned)
		}
	}
	return out
}

func cloneAdaptiveSpec(src *snippets.AdaptivePageSpec) *snippets.AdaptivePageSpec {
	if src == nil {
		return nil
	}
	cp := *src
	if src.MaxWidthDp != nil {
		w := *src.MaxWidthDp
		cp.MaxWidthDp = &w
	}
	if len(src.SecondarySnippetIDs) > 0 {
		cp.SecondarySnippetIDs = append([]string(nil), src.SecondarySnippetIDs...)
	}
	return &cp
}

func buildPageTracking(layoutSlug string, snippetCount int, ctx domain.RequestContext) *atoms.TrackingMeta {
	props := map[string]any{
		"layout_slug":   layoutSlug,
		"snippet_count": strconv.Itoa(snippetCount),
	}
	if ctx.TenantSlug != "" {
		props["tenant_slug"] = ctx.TenantSlug
	}
	if ctx.Platform != "" {
		props["platform"] = ctx.Platform
	}
	if ctx.City != "" {
		props["city"] = ctx.City
	}
	if ctx.Zone != "" {
		props["zone"] = ctx.Zone
	}
	if ctx.Segment != "" {
		props["segment"] = ctx.Segment
	}
	return &atoms.TrackingMeta{
		EventName:  "page_view",
		Properties: props,
	}
}
