package transformer

import (
	"encoding/json"
	"strings"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// MergeAndInterpolateProps deep-merges overrides over base without mutating either input,
// then interpolates context placeholders ({{city}}, {{zone}}, {{segment}}, {{platform}},
// {{app_version}}, {{tenant}}, {{window_class}}, and RequestContext.Attributes) across
// all string leaves.
func MergeAndInterpolateProps(base, overrides map[string]any, ctx domain.RequestContext) map[string]any {
	merged := deepMergeMaps(base, overrides)
	interpolated, _ := interpolateValue(merged, ctx).(map[string]any)
	if interpolated == nil {
		return map[string]any{}
	}
	return interpolated
}

// InterpolateString replaces supported {{placeholder}} tokens in s with values from ctx.
// Unknown placeholders are left intact.
func InterpolateString(s string, ctx domain.RequestContext) string {
	if !strings.Contains(s, "{{") {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	rest := s
	for {
		start := strings.Index(rest, "{{")
		if start < 0 {
			b.WriteString(rest)
			break
		}
		b.WriteString(rest[:start])
		afterOpen := rest[start+2:]
		end := strings.Index(afterOpen, "}}")
		if end < 0 {
			b.WriteString(rest[start:])
			break
		}

		rawKey := strings.TrimSpace(afterOpen[:end])
		if val, ok := lookupTemplateToken(rawKey, ctx); ok {
			b.WriteString(val)
		} else {
			b.WriteString(rest[start : start+2+end+2])
		}
		rest = afterOpen[end+2:]
	}
	return b.String()
}

func lookupTemplateToken(key string, ctx domain.RequestContext) (string, bool) {
	switch strings.ToLower(key) {
	case "city":
		return ctx.City, true
	case "zone":
		return ctx.Zone, true
	case "segment":
		return ctx.Segment, true
	case "platform":
		return ctx.Platform, true
	case "app_version":
		return ctx.AppVersion, true
	case "tenant", "tenant_slug":
		return ctx.TenantSlug, true
	case "window_class":
		return ctx.WindowClass, true
	case "user_id":
		return ctx.UserID, true
	}

	if len(ctx.Attributes) == 0 {
		return "", false
	}
	attrKey := strings.TrimPrefix(key, "attr.")
	if v, ok := ctx.Attributes[attrKey]; ok {
		return v, true
	}
	for k, v := range ctx.Attributes {
		if strings.EqualFold(k, attrKey) {
			return v, true
		}
	}
	return "", false
}

func deepMergeMaps(base, overrides map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(overrides))
	for k, v := range base {
		out[k] = deepCloneValue(v)
	}
	for k, v := range overrides {
		existingMap, existingIsMap := out[k].(map[string]any)
		overrideMap, overrideIsMap := v.(map[string]any)
		if existingIsMap && overrideIsMap {
			out[k] = deepMergeMaps(existingMap, overrideMap)
		} else {
			out[k] = deepCloneValue(v)
		}
	}
	return out
}

func deepCloneValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		cp := make(map[string]any, len(val))
		for k, elem := range val {
			cp[k] = deepCloneValue(elem)
		}
		return cp
	case []any:
		cp := make([]any, len(val))
		for i, elem := range val {
			cp[i] = deepCloneValue(elem)
		}
		return cp
	case []map[string]any:
		cp := make([]map[string]any, len(val))
		for i, elem := range val {
			cloned, _ := deepCloneValue(elem).(map[string]any)
			cp[i] = cloned
		}
		return cp
	case []string:
		cp := make([]string, len(val))
		copy(cp, val)
		return cp
	default:
		return val
	}
}

func interpolateValue(v any, ctx domain.RequestContext) any {
	switch val := v.(type) {
	case string:
		return InterpolateString(val, ctx)
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, elem := range val {
			out[k] = interpolateValue(elem, ctx)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, elem := range val {
			out[i] = interpolateValue(elem, ctx)
		}
		return out
	case []map[string]any:
		out := make([]map[string]any, len(val))
		for i, elem := range val {
			interpolated, _ := interpolateValue(elem, ctx).(map[string]any)
			out[i] = interpolated
		}
		return out
	case []string:
		out := make([]string, len(val))
		for i, elem := range val {
			out[i] = InterpolateString(elem, ctx)
		}
		return out
	default:
		return val
	}
}

func cloneProps(props map[string]any) map[string]any {
	if props == nil {
		return map[string]any{}
	}
	cp, _ := deepCloneValue(props).(map[string]any)
	if cp == nil {
		return map[string]any{}
	}
	return cp
}

func promoteTextFields(props map[string]any, keys ...string) {
	for _, k := range keys {
		if s, ok := props[k].(string); ok {
			props[k] = map[string]any{"text": s}
		}
	}
}

func promoteTagFields(props map[string]any, keys ...string) {
	for _, k := range keys {
		if s, ok := props[k].(string); ok {
			props[k] = map[string]any{"label": s}
		}
	}
}

func promoteIconFields(props map[string]any, keys ...string) {
	for _, k := range keys {
		if s, ok := props[k].(string); ok {
			props[k] = map[string]any{"name": s}
		}
	}
}

func promoteImageFields(props map[string]any, keys ...string) {
	for _, k := range keys {
		if s, ok := props[k].(string); ok {
			props[k] = map[string]any{"url": s}
		}
	}
}

func promoteColorRefFields(props map[string]any, keys ...string) {
	for _, k := range keys {
		if s, ok := props[k].(string); ok {
			props[k] = map[string]any{"token": s}
		}
	}
}

func promoteTagSliceField(props map[string]any, key string) {
	raw, exists := props[key]
	if !exists || raw == nil {
		return
	}
	switch list := raw.(type) {
	case []string:
		out := make([]any, len(list))
		for i, s := range list {
			out[i] = map[string]any{"label": s}
		}
		props[key] = out
	case []any:
		out := make([]any, len(list))
		for i, elem := range list {
			if s, ok := elem.(string); ok {
				out[i] = map[string]any{"label": s}
			} else {
				out[i] = elem
			}
		}
		props[key] = out
	}
}

func promoteImageSliceField(props map[string]any, key string) {
	raw, exists := props[key]
	if !exists || raw == nil {
		return
	}
	switch list := raw.(type) {
	case []string:
		out := make([]any, len(list))
		for i, s := range list {
			out[i] = map[string]any{"url": s}
		}
		props[key] = out
	case []any:
		out := make([]any, len(list))
		for i, elem := range list {
			if s, ok := elem.(string); ok {
				out[i] = map[string]any{"url": s}
			} else {
				out[i] = elem
			}
		}
		props[key] = out
	}
}

func resolveSticky(input TransformInput) bool {
	if input.Slot.Overrides != nil {
		if raw, exists := input.Slot.Overrides["sticky"]; exists {
			if b, ok := raw.(bool); ok {
				return b
			}
		}
	}
	if input.Slot.Sticky {
		return true
	}
	if input.Props != nil {
		if b, ok := input.Props["sticky"].(bool); ok {
			return b
		}
	}
	return false
}

func extractTabTitle(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case map[string]any:
		if s, ok := val["text"].(string); ok {
			return s
		}
	}
	return ""
}

func decodeTag(v any) (*atoms.TagAtom, error) {
	if s, ok := v.(string); ok {
		return &atoms.TagAtom{Label: s}, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var tag atoms.TagAtom
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func toStringSlice(raw any) ([]string, bool) {
	switch list := raw.(type) {
	case nil:
		return nil, false
	case []string:
		return list, true
	case []any:
		out := make([]string, 0, len(list))
		for _, elem := range list {
			if s, ok := elem.(string); ok {
				out = append(out, s)
			}
		}
		return out, true
	default:
		return nil, false
	}
}
