// Package usecase implements Clean Architecture application services for dynamic page
// resolution, quick-commerce Studio authoring/governance, and SDUI action execution.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/abeta-dev/go-libs/apperror"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/theming"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

const DefaultTenantSlug = "abeta-fresh"

// ResolveService resolves dynamic published layouts (or drafts for Studio preview),
// rule-targeted splash configs, and runtime design token catalogs.
type ResolveService struct {
	store     domain.Store
	evaluator domain.RuleEvaluator
	slotting  domain.SlottingEngine
	composer  domain.PageComposer
}

// NewResolveService constructs a ResolveService.
func NewResolveService(
	store domain.Store,
	evaluator domain.RuleEvaluator,
	slotting domain.SlottingEngine,
	composer domain.PageComposer,
) *ResolveService {
	return &ResolveService{
		store:     store,
		evaluator: evaluator,
		slotting:  slotting,
		composer:  composer,
	}
}

// ListPublishedScreens returns all layout slugs available for a tenant (published or fallback).
func (s *ResolveService) ListPublishedScreens(ctx context.Context, tenantSlug string) ([]string, error) {
	tenant, err := s.resolveTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	layouts, err := s.store.ListLayouts(ctx, tenant.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list layouts")
	}
	slugs := make([]string, 0, len(layouts))
	for _, l := range layouts {
		if l.PublishedVersion > 0 {
			slugs = append(slugs, l.Slug)
		}
	}
	sort.Strings(slugs)
	return slugs, nil
}

// ResolvePage resolves the active published snapshot for layoutSlug under reqCtx,
// applying targeting rules, schedule windows, dynamic slotting, widget transformation,
// and automatic fallback protection on error.
func (s *ResolveService) ResolvePage(
	ctx context.Context,
	layoutSlug string,
	reqCtx domain.RequestContext,
) (domain.ResolvedPageResult, error) {
	return s.resolveInternal(ctx, layoutSlug, reqCtx, false)
}

// PreviewPage resolves either the current editable Draft (when draft=true) or the active
// Published Snapshot (when draft=false) with full ResolutionTrace for Layout Studio.
func (s *ResolveService) PreviewPage(
	ctx context.Context,
	layoutSlug string,
	reqCtx domain.RequestContext,
	draft bool,
) (domain.ResolvedPageResult, error) {
	return s.resolveInternal(ctx, layoutSlug, reqCtx, draft)
}

func (s *ResolveService) resolveInternal(
	ctx context.Context,
	layoutSlug string,
	reqCtx domain.RequestContext,
	previewDraft bool,
) (domain.ResolvedPageResult, error) {
	started := time.Now()
	cleanSlug := strings.Trim(strings.TrimSpace(layoutSlug), "/")
	if cleanSlug == "" {
		return domain.ResolvedPageResult{}, apperror.BadRequest("layout slug must not be empty")
	}
	if reqCtx.Now.IsZero() {
		reqCtx.Now = time.Now().UTC()
	}

	tenant, err := s.resolveTenant(ctx, reqCtx.TenantSlug)
	if err != nil {
		return domain.ResolvedPageResult{}, err
	}
	reqCtx.TenantSlug = tenant.Slug

	trace := domain.ResolutionTrace{
		TenantSlug:   tenant.Slug,
		LayoutSlug:   cleanSlug,
		PreviewDraft: previewDraft,
		ResolvedAt:   reqCtx.Now,
	}

	var (
		title      string
		pageConfig domain.LayoutPageConfig
		baseSlots  []domain.LayoutSlot
		widgetMap  map[string]domain.Widget
	)

	if previewDraft {
		layout, err := s.store.GetLayoutBySlug(ctx, tenant.ID, cleanSlug)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ResolvedPageResult{}, apperror.NotFound(fmt.Sprintf("layout %q not found", cleanSlug))
			}
			return domain.ResolvedPageResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load draft layout")
		}
		trace.SnapshotVersion = layout.DraftVersion
		title = layout.Title
		pageConfig = layout.PageConfig
		baseSlots = layout.Slots
		allWidgets, err := s.store.ListWidgets(ctx, tenant.ID)
		if err != nil {
			return domain.ResolvedPageResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load tenant widgets")
		}
		widgetMap = make(map[string]domain.Widget, len(allWidgets))
		for _, w := range allWidgets {
			widgetMap[w.ID] = w
		}
	} else {
		snap, err := s.store.GetLatestSnapshot(ctx, tenant.ID, cleanSlug)
		if err != nil {
			if fb, fbErr := s.store.GetFallback(ctx, tenant.ID, cleanSlug); fbErr == nil {
				trace.UsedFallback = true
				trace.SnapshotVersion = fb.SnapshotVersion
				trace.FallbackReason = fmt.Sprintf("snapshot unavailable: %v", err)
				trace.DurationMs = time.Since(started).Milliseconds()
				return domain.ResolvedPageResult{Page: fb.Page, Trace: trace}, nil
			}
			if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrNoPublishedSnapshot) {
				return domain.ResolvedPageResult{}, apperror.NotFound(fmt.Sprintf("published layout %q not found", cleanSlug))
			}
			return domain.ResolvedPageResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load published snapshot")
		}
		trace.SnapshotVersion = snap.Version
		title = snap.Title
		pageConfig = snap.PageConfig
		baseSlots = snap.Slots
		widgetMap = make(map[string]domain.Widget, len(snap.ResolvedWidgets))
		for id, w := range snap.ResolvedWidgets {
			widgetMap[id] = w
		}
	}

	rules, err := s.store.ListRules(ctx, tenant.ID)
	if err != nil {
		return s.fallbackOrError(ctx, tenant.ID, cleanSlug, trace, started, previewDraft, fmt.Errorf("load rules: %w", err))
	}
	ruleMap := make(map[string]domain.TargetingRule, len(rules))
	for _, r := range rules {
		ruleMap[r.ID] = r
		if r.Key != "" {
			ruleMap[r.Key] = r
		}
	}

	// Filter base slots by enabled flag, schedule window, and attached targeting rule.
	filteredBase := make([]domain.LayoutSlot, 0, len(baseSlots))
	matchedRuleSet := make(map[string]struct{})
	for _, slot := range baseSlots {
		if !slot.Enabled {
			trace.SkippedSlotIDs = append(trace.SkippedSlotIDs, slot.SlotID)
			continue
		}
		if slot.Schedule != nil && !s.evaluator.MatchesSchedule(slot.Schedule, reqCtx.Now) {
			trace.SkippedSlotIDs = append(trace.SkippedSlotIDs, slot.SlotID)
			continue
		}
		if slot.RuleID != "" {
			rule, ok := ruleMap[slot.RuleID]
			if !ok || !s.evaluator.MatchesRule(rule, reqCtx) {
				trace.SkippedSlotIDs = append(trace.SkippedSlotIDs, slot.SlotID)
				continue
			}
			matchedRuleSet[rule.ID] = struct{}{}
		}
		filteredBase = append(filteredBase, slot)
	}

	// Apply dynamic slotting rules (priority-ordered VIP / scheduled / promotional injections).
	slottingRules, err := s.store.ListSlottingRules(ctx, tenant.ID, cleanSlug)
	if err != nil {
		return s.fallbackOrError(ctx, tenant.ID, cleanSlug, trace, started, previewDraft, fmt.Errorf("load slotting rules: %w", err))
	}
	mergedSlots, appliedSlotting := s.slotting.Apply(filteredBase, slottingRules, ruleMap, reqCtx)
	trace.AppliedSlottingIDs = appliedSlotting

	// Ensure widgets injected by slotting rules (or nested children) are hydrated into widgetMap.
	missingIDs := make([]string, 0)
	for _, slot := range mergedSlots {
		if _, exists := widgetMap[slot.WidgetID]; !exists && slot.WidgetID != "" {
			missingIDs = append(missingIDs, slot.WidgetID)
		}
	}
	if len(missingIDs) > 0 {
		extra, err := s.store.GetWidgetsByIDs(ctx, tenant.ID, missingIDs)
		if err == nil {
			for id, w := range extra {
				widgetMap[id] = w
			}
		}
	}

	// Collect matched rule IDs deterministically.
	for _, sr := range slottingRules {
		for _, appliedID := range appliedSlotting {
			if sr.ID == appliedID && sr.RuleID != "" {
				if r, ok := ruleMap[sr.RuleID]; ok {
					matchedRuleSet[r.ID] = struct{}{}
				}
			}
		}
	}
	for id := range matchedRuleSet {
		trace.MatchedRuleIDs = append(trace.MatchedRuleIDs, id)
	}
	sort.Strings(trace.MatchedRuleIDs)

	page, err := s.composer.Compose(cleanSlug, title, pageConfig, mergedSlots, widgetMap, reqCtx)
	if err != nil {
		return s.fallbackOrError(ctx, tenant.ID, cleanSlug, trace, started, previewDraft, fmt.Errorf("compose page: %w", err))
	}

	trace.DurationMs = time.Since(started).Milliseconds()
	return domain.ResolvedPageResult{
		Page:  page,
		Trace: trace,
	}, nil
}

func (s *ResolveService) fallbackOrError(
	ctx context.Context,
	tenantID, slug string,
	trace domain.ResolutionTrace,
	started time.Time,
	previewDraft bool,
	cause error,
) (domain.ResolvedPageResult, error) {
	if !previewDraft {
		if fb, err := s.store.GetFallback(ctx, tenantID, slug); err == nil {
			trace.UsedFallback = true
			trace.SnapshotVersion = fb.SnapshotVersion
			trace.FallbackReason = cause.Error()
			trace.DurationMs = time.Since(started).Milliseconds()
			return domain.ResolvedPageResult{Page: fb.Page, Trace: trace}, nil
		}
	}
	return domain.ResolvedPageResult{}, apperror.Wrap(cause, apperror.CodeInternal, "failed to resolve layout")
}

// ResolveSplash selects the highest-priority matching SplashConfig or the tenant default.
func (s *ResolveService) ResolveSplash(ctx context.Context, reqCtx domain.RequestContext) (domain.SplashConfig, error) {
	if reqCtx.Now.IsZero() {
		reqCtx.Now = time.Now().UTC()
	}
	tenant, err := s.resolveTenant(ctx, reqCtx.TenantSlug)
	if err != nil {
		return domain.SplashConfig{}, err
	}
	configs, err := s.store.ListSplashConfigs(ctx, tenant.ID)
	if err != nil {
		return domain.SplashConfig{}, apperror.Wrap(err, apperror.CodeInternal, "failed to list splash configs")
	}
	rules, _ := s.store.ListRules(ctx, tenant.ID)
	ruleMap := make(map[string]domain.TargetingRule, len(rules))
	for _, r := range rules {
		ruleMap[r.ID] = r
		ruleMap[r.Key] = r
	}

	var fallback *domain.SplashConfig
	for i := range configs {
		cfg := configs[i]
		if !cfg.Enabled {
			continue
		}
		if cfg.IsDefault {
			c := cfg
			fallback = &c
		}
		if cfg.Schedule != nil && !s.evaluator.MatchesSchedule(cfg.Schedule, reqCtx.Now) {
			continue
		}
		if cfg.RuleID != "" {
			rule, ok := ruleMap[cfg.RuleID]
			if !ok || !s.evaluator.MatchesRule(rule, reqCtx) {
				continue
			}
			return cfg, nil
		}
	}
	if fallback != nil {
		return *fallback, nil
	}
	return domain.SplashConfig{}, apperror.NotFound("no active splash configuration found")
}

// ResolveTokenCatalog returns the default KMP theming catalog overlaid with any matching tenant TokenOverrides.
func (s *ResolveService) ResolveTokenCatalog(ctx context.Context, reqCtx domain.RequestContext) (map[string]any, error) {
	if reqCtx.Now.IsZero() {
		reqCtx.Now = time.Now().UTC()
	}
	base := map[string]any{
		"version":       "1.0.0",
		"colors":        theming.AllColorTokens(),
		"typography":    theming.AllTypographyTokens(),
		"spacing":       theming.AllSpacingScales(),
		"corner_radius": theming.AllCornerRadiusScales(),
	}
	result := map[string]any{
		"catalog":          base,
		"applied_override": "",
		"overrides":        map[string]string{},
	}

	tenant, err := s.resolveTenant(ctx, reqCtx.TenantSlug)
	if err != nil {
		return result, nil
	}
	overrides, err := s.store.ListTokenOverrides(ctx, tenant.ID)
	if err != nil {
		return result, nil
	}
	rules, _ := s.store.ListRules(ctx, tenant.ID)
	ruleMap := make(map[string]domain.TargetingRule, len(rules))
	for _, r := range rules {
		ruleMap[r.ID] = r
		ruleMap[r.Key] = r
	}

	for _, ov := range overrides {
		if !ov.Enabled {
			continue
		}
		if ov.RuleID != "" {
			rule, ok := ruleMap[ov.RuleID]
			if !ok || !s.evaluator.MatchesRule(rule, reqCtx) {
				continue
			}
		}
		result["applied_override"] = ov.Key
		result["theme_preset"] = ov.ThemePreset
		result["overrides"] = ov.Tokens
		if ov.RuleID != "" {
			break
		}
	}
	return result, nil
}

func (s *ResolveService) resolveTenant(ctx context.Context, slug string) (domain.Tenant, error) {
	clean := strings.TrimSpace(slug)
	if clean == "" {
		clean = DefaultTenantSlug
	}
	t, err := s.store.GetTenantBySlug(ctx, clean)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Tenant{}, apperror.NotFound(fmt.Sprintf("tenant %q not found", clean))
		}
		return domain.Tenant{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load tenant")
	}
	if !t.Active {
		return domain.Tenant{}, apperror.BadRequest(fmt.Sprintf("tenant %q is inactive", clean))
	}
	return t, nil
}
