// Package transformer provides an Open/Closed SDUI widget transformer kit that converts
// dynamic layout widgets into v1 Abeta KMP SnippetEnvelope and PageResponse wire payloads.
package transformer

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

// TransformInput carries all context required to transform a CMS Widget bound to a
// LayoutSlot into a strongly-typed KMP v1 SnippetEnvelope.
type TransformInput struct {
	Slot     domain.LayoutSlot
	Widget   domain.Widget
	Props    map[string]any
	Widgets  map[string]domain.Widget
	Context  domain.RequestContext
	Registry *Registry

	parentID string
	visited  map[string]bool
}

// WidgetTransformer converts a Widget of a single SnippetType into a SnippetEnvelope.
type WidgetTransformer interface {
	SnippetType() string
	Transform(input TransformInput) (snippets.SnippetEnvelope, error)
}

// WidgetTransformerFunc adapts a plain function to the WidgetTransformer interface.
type WidgetTransformerFunc struct {
	Type string
	Fn   func(input TransformInput) (snippets.SnippetEnvelope, error)
}

// NewWidgetTransformerFunc creates a WidgetTransformerFunc for snippetType.
func NewWidgetTransformerFunc(
	snippetType string,
	fn func(input TransformInput) (snippets.SnippetEnvelope, error),
) WidgetTransformerFunc {
	return WidgetTransformerFunc{Type: snippetType, Fn: fn}
}

// SnippetType returns the wire snippet type handled by this transformer.
func (f WidgetTransformerFunc) SnippetType() string {
	return f.Type
}

// Transform delegates to Fn.
func (f WidgetTransformerFunc) Transform(input TransformInput) (snippets.SnippetEnvelope, error) {
	if f.Fn == nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: nil transform function for %q", f.Type)
	}
	return f.Fn(input)
}

// Registry is an Open/Closed, concurrency-safe registry of WidgetTransformers keyed by
// KMP snippet_type. Callers may register custom 22nd+ snippet types without modifying
// built-in transformers.
type Registry struct {
	mu           sync.RWMutex
	transformers map[string]WidgetTransformer
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		transformers: make(map[string]WidgetTransformer),
	}
}

// Register registers or replaces the transformer for t.SnippetType().
func (r *Registry) Register(t WidgetTransformer) {
	if r == nil || t == nil {
		return
	}
	kind := strings.TrimSpace(t.SnippetType())
	if kind == "" {
		return
	}
	r.mu.Lock()
	r.transformers[kind] = t
	r.mu.Unlock()
}

// Transform looks up the transformer for input.Widget.SnippetType, executes it, and
// enriches envelope identity, parent hierarchy, tracking metadata, and motion specs.
func (r *Registry) Transform(input TransformInput) (snippets.SnippetEnvelope, error) {
	if r == nil {
		return snippets.SnippetEnvelope{}, fmt.Errorf("transformer: nil registry")
	}

	kind := strings.TrimSpace(input.Widget.SnippetType)
	r.mu.RLock()
	t, ok := r.transformers[kind]
	r.mu.RUnlock()
	if !ok {
		return snippets.SnippetEnvelope{}, fmt.Errorf(
			"transformer: unsupported snippet type %q for widget %q",
			input.Widget.SnippetType,
			input.Widget.ID,
		)
	}

	if input.Registry == nil {
		input.Registry = r
	}
	if input.Props == nil {
		input.Props = MergeAndInterpolateProps(input.Widget.Props, input.Slot.Overrides, input.Context)
	}

	env, err := t.Transform(input)
	if err != nil {
		return snippets.SnippetEnvelope{}, err
	}

	if strings.TrimSpace(env.SnippetID) == "" {
		env.SnippetID = resolveSnippetID(input)
	}
	if strings.TrimSpace(env.SnippetType) == "" {
		env.SnippetType = kind
	}
	if strings.TrimSpace(env.SnippetParentID) == "" && input.parentID != "" {
		env.SnippetParentID = input.parentID
	}
	if env.Tracking == nil && len(input.Widget.Tracking) > 0 {
		tracking, err := decodeTracking(input.Widget.Tracking, input.Context)
		if err != nil {
			return snippets.SnippetEnvelope{}, fmt.Errorf(
				"transformer: decode tracking for widget %q: %w",
				input.Widget.ID,
				err,
			)
		}
		env.Tracking = tracking
	}
	if env.Motion == nil && len(input.Widget.Motion) > 0 {
		motion, err := decodeMotion(input.Widget.Motion)
		if err != nil {
			return snippets.SnippetEnvelope{}, fmt.Errorf(
				"transformer: decode motion for widget %q: %w",
				input.Widget.ID,
				err,
			)
		}
		env.Motion = motion
	}

	return env, nil
}

// SupportedTypes returns a sorted list of all registered snippet type identifiers.
func (r *Registry) SupportedTypes() []string {
	if r == nil {
		return []string{}
	}
	r.mu.RLock()
	out := make([]string, 0, len(r.transformers))
	for k := range r.transformers {
		out = append(out, k)
	}
	r.mu.RUnlock()
	slices.Sort(out)
	return out
}

func resolveSnippetID(input TransformInput) string {
	if s := strings.TrimSpace(input.Slot.SlotID); s != "" {
		return s
	}
	if s := strings.TrimSpace(input.Widget.Slug); s != "" {
		return s
	}
	return strings.TrimSpace(input.Widget.ID)
}

func decodeTracking(raw map[string]any, ctx domain.RequestContext) (*atoms.TrackingMeta, error) {
	interpolated := MergeAndInterpolateProps(raw, nil, ctx)
	data, err := json.Marshal(interpolated)
	if err != nil {
		return nil, err
	}
	var meta atoms.TrackingMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	if meta.EventName == "" && len(meta.Properties) == 0 {
		return nil, nil
	}
	if len(meta.Properties) > 0 {
		strProps := make(map[string]any, len(meta.Properties))
		for k, v := range meta.Properties {
			if v == nil {
				continue
			}
			strProps[k] = fmt.Sprint(v)
		}
		meta.Properties = strProps
	}
	return &meta, nil
}

func decodeMotion(raw map[string]any) (*atoms.MotionSpec, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var spec atoms.MotionSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}
