package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abeta-dev/go-libs/apperror"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/google/uuid"
)

// DiffFunc computes field-level diffs between two JSON-serializable revisions.
type DiffFunc func(before, after any) []domain.DiffItem

// PublishResult captures the newly published immutable snapshot and its audit diff.
type PublishResult struct {
	Snapshot domain.PublishedSnapshot `json:"snapshot"`
	Diff     []domain.DiffItem        `json:"diff"`
}

// RollbackResult captures the newly restored snapshot and synchronized draft layout.
type RollbackResult struct {
	Snapshot domain.PublishedSnapshot `json:"snapshot"`
	Layout   domain.Layout            `json:"layout"`
}

// StudioService provides CRUD, validation, immutable snapshot publishing, rollback,
// fallback promotion, and audit changelog governance for Layout Studio.
type StudioService struct {
	store     domain.Store
	validator domain.LayoutValidator
	composer  domain.PageComposer
	diffFn    DiffFunc
}

// NewStudioService constructs a StudioService.
func NewStudioService(
	store domain.Store,
	validator domain.LayoutValidator,
	composer domain.PageComposer,
	diffFn DiffFunc,
) *StudioService {
	if diffFn == nil {
		diffFn = func(_, _ any) []domain.DiffItem { return nil }
	}
	return &StudioService{
		store:     store,
		validator: validator,
		composer:  composer,
		diffFn:    diffFn,
	}
}

// ── Tenants & Templates ────────────────────────────────────────────────────────

func (s *StudioService) ListTenants(ctx context.Context) ([]domain.Tenant, error) {
	items, err := s.store.ListTenants(ctx)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list tenants")
	}
	return items, nil
}

func (s *StudioService) UpsertTenant(ctx context.Context, tenant domain.Tenant, actor string) (domain.Tenant, error) {
	tenant.Slug = strings.TrimSpace(strings.ToLower(tenant.Slug))
	tenant.Name = strings.TrimSpace(tenant.Name)
	if tenant.Slug == "" || tenant.Name == "" {
		return domain.Tenant{}, apperror.BadRequest("tenant slug and name are required")
	}
	now := time.Now().UTC()
	if tenant.ID == "" {
		if existing, err := s.store.GetTenantBySlug(ctx, tenant.Slug); err == nil {
			tenant.ID = existing.ID
			tenant.CreatedAt = existing.CreatedAt
		} else {
			tenant.ID = "ten_" + uuid.NewString()[:8]
			tenant.CreatedAt = now
		}
	}
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = now
	}
	tenant.UpdatedAt = now
	if !tenant.Active {
		tenant.Active = true
	}
	saved, err := s.store.UpsertTenant(ctx, tenant)
	if err != nil {
		return domain.Tenant{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save tenant")
	}
	s.recordAudit(ctx, saved.ID, "tenant", saved.ID, "upsert", actor, fmt.Sprintf("Upserted tenant %s", saved.Slug), nil)
	return saved, nil
}

func (s *StudioService) ListTemplates(ctx context.Context, tenantSlug string) ([]domain.WidgetTemplate, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListTemplates(ctx, t.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list templates")
	}
	return items, nil
}

// ── Widgets ────────────────────────────────────────────────────────────────────

func (s *StudioService) ListWidgets(ctx context.Context, tenantSlug string) ([]domain.Widget, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListWidgets(ctx, t.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list widgets")
	}
	return items, nil
}

func (s *StudioService) UpsertWidget(ctx context.Context, tenantSlug string, widget domain.Widget, actor string) (domain.Widget, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.Widget{}, err
	}
	widget.TenantID = t.ID
	widget.Name = strings.TrimSpace(widget.Name)
	widget.Slug = strings.TrimSpace(widget.Slug)
	widget.SnippetType = strings.TrimSpace(widget.SnippetType)
	if widget.Name == "" || widget.SnippetType == "" {
		return domain.Widget{}, apperror.BadRequest("widget name and snippet_type are required")
	}
	if widget.Slug == "" {
		widget.Slug = strings.ToLower(strings.ReplaceAll(widget.Name, " ", "-"))
	}
	now := time.Now().UTC()
	var before any
	if widget.ID == "" {
		widget.ID = "w_" + uuid.NewString()[:8]
		widget.CreatedAt = now
	} else if prev, err := s.store.GetWidgetByID(ctx, t.ID, widget.ID); err == nil {
		before = prev
		widget.CreatedAt = prev.CreatedAt
	}
	if widget.CreatedAt.IsZero() {
		widget.CreatedAt = now
	}
	widget.UpdatedAt = now
	if widget.Props == nil {
		widget.Props = map[string]any{}
	}
	saved, err := s.store.UpsertWidget(ctx, widget)
	if err != nil {
		return domain.Widget{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save widget")
	}
	s.recordAudit(ctx, t.ID, "widget", saved.ID, "upsert", actor, fmt.Sprintf("Saved widget %s (%s)", saved.Name, saved.SnippetType), s.diffFn(before, saved))
	return saved, nil
}

func (s *StudioService) DeleteWidget(ctx context.Context, tenantSlug, widgetID, actor string) error {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return err
	}
	if err := s.store.DeleteWidget(ctx, t.ID, widgetID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperror.NotFound("widget not found")
		}
		return apperror.Wrap(err, apperror.CodeInternal, "failed to delete widget")
	}
	s.recordAudit(ctx, t.ID, "widget", widgetID, "delete", actor, fmt.Sprintf("Deleted widget %s", widgetID), nil)
	return nil
}

// ── Layouts, Validation, Publishing & Rollback ─────────────────────────────────

func (s *StudioService) ListLayouts(ctx context.Context, tenantSlug string) ([]domain.Layout, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListLayouts(ctx, t.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list layouts")
	}
	return items, nil
}

func (s *StudioService) GetLayout(ctx context.Context, tenantSlug, slug string) (domain.Layout, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.Layout{}, err
	}
	layout, err := s.store.GetLayoutBySlug(ctx, t.ID, strings.TrimSpace(slug))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Layout{}, apperror.NotFound(fmt.Sprintf("layout %q not found", slug))
		}
		return domain.Layout{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load layout")
	}
	return layout, nil
}

func (s *StudioService) UpsertLayout(ctx context.Context, tenantSlug, slug string, layout domain.Layout, actor string) (domain.Layout, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.Layout{}, err
	}
	cleanSlug := strings.Trim(strings.TrimSpace(slug), "/")
	if cleanSlug == "" {
		cleanSlug = strings.Trim(strings.TrimSpace(layout.Slug), "/")
	}
	if cleanSlug == "" {
		return domain.Layout{}, apperror.BadRequest("layout slug is required")
	}
	layout.TenantID = t.ID
	layout.Slug = cleanSlug
	if strings.TrimSpace(layout.Title) == "" {
		layout.Title = cleanSlug
	}
	now := time.Now().UTC()
	var before any
	if existing, err := s.store.GetLayoutBySlug(ctx, t.ID, cleanSlug); err == nil {
		before = existing
		layout.ID = existing.ID
		layout.CreatedAt = existing.CreatedAt
		layout.PublishedVersion = existing.PublishedVersion
		layout.DraftVersion = existing.DraftVersion + 1
		if layout.Status == "" {
			layout.Status = existing.Status
		}
	} else {
		if layout.ID == "" {
			layout.ID = "lay_" + uuid.NewString()[:8]
		}
		layout.CreatedAt = now
		layout.DraftVersion = 1
		layout.Status = domain.LayoutStatusDraft
	}
	for i := range layout.Slots {
		if strings.TrimSpace(layout.Slots[i].SlotID) == "" {
			layout.Slots[i].SlotID = fmt.Sprintf("slot_%d_%s", i, uuid.NewString()[:4])
		}
		layout.Slots[i].Position = i
	}
	layout.UpdatedAt = now

	saved, err := s.store.UpsertLayout(ctx, layout)
	if err != nil {
		return domain.Layout{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save layout draft")
	}
	s.recordAudit(ctx, t.ID, "layout", saved.Slug, "save_draft", actor, fmt.Sprintf("Saved draft v%d for %s", saved.DraftVersion, saved.Slug), s.diffFn(before, saved))
	return saved, nil
}

func (s *StudioService) ValidateLayout(ctx context.Context, tenantSlug, slug string) ([]domain.ValidationIssue, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	layout, err := s.store.GetLayoutBySlug(ctx, t.ID, strings.TrimSpace(slug))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperror.NotFound(fmt.Sprintf("layout %q not found", slug))
		}
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to load layout")
	}
	widgets, err := s.widgetMap(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	return s.validator.ValidateLayout(layout, widgets), nil
}

// PublishLayout validates a layout draft, captures an immutable versioned snapshot with all
// referenced (and nested child) widgets frozen, updates the published pointer, refreshes
// the fallback layout if none exists, and records an audit changelog with JSON diff.
func (s *StudioService) PublishLayout(ctx context.Context, tenantSlug, slug, actor, comment string) (PublishResult, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return PublishResult{}, err
	}
	cleanSlug := strings.TrimSpace(slug)
	layout, err := s.store.GetLayoutBySlug(ctx, t.ID, cleanSlug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return PublishResult{}, apperror.NotFound(fmt.Sprintf("layout %q not found", cleanSlug))
		}
		return PublishResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load layout")
	}
	widgets, err := s.widgetMap(ctx, t.ID)
	if err != nil {
		return PublishResult{}, err
	}
	if issues := s.validator.ValidateLayout(layout, widgets); len(issues) > 0 {
		return PublishResult{}, apperror.NewWithDetails(apperror.CodeBadRequest, "layout validation failed", map[string]any{"issues": issues})
	}

	var prevSnap any
	if prev, err := s.store.GetLatestSnapshot(ctx, t.ID, cleanSlug); err == nil {
		prevSnap = prev
	}

	nextVersion := layout.PublishedVersion + 1
	now := time.Now().UTC()
	if strings.TrimSpace(actor) == "" {
		actor = "studio-admin"
	}

	frozenWidgets := freezeReferencedWidgets(layout.Slots, widgets)
	snap := domain.PublishedSnapshot{
		ID:              "snap_" + uuid.NewString()[:8],
		TenantID:        t.ID,
		LayoutID:        layout.ID,
		LayoutSlug:      layout.Slug,
		Version:         nextVersion,
		Title:           layout.Title,
		PageConfig:      layout.PageConfig,
		Slots:           layout.Slots,
		ResolvedWidgets: frozenWidgets,
		PublishedBy:     actor,
		Comment:         strings.TrimSpace(comment),
		PublishedAt:     now,
	}
	if err := s.store.SaveSnapshot(ctx, snap); err != nil {
		return PublishResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save snapshot")
	}

	layout.PublishedVersion = nextVersion
	layout.Status = domain.LayoutStatusPublished
	layout.UpdatedAt = now
	if _, err := s.store.UpsertLayout(ctx, layout); err != nil {
		return PublishResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to update published layout version")
	}

	// Ensure a safe fallback exists on first publish.
	if _, fbErr := s.store.GetFallback(ctx, t.ID, cleanSlug); errors.Is(fbErr, domain.ErrNotFound) {
		_, _ = s.PromoteFallback(ctx, tenantSlug, cleanSlug, actor)
	}

	diff := s.diffFn(prevSnap, snap)
	s.recordAudit(ctx, t.ID, "layout", layout.Slug, "publish", actor, fmt.Sprintf("Published %s v%d", layout.Slug, nextVersion), diff)
	return PublishResult{Snapshot: snap, Diff: diff}, nil
}

func (s *StudioService) ListSnapshots(ctx context.Context, tenantSlug, slug string) ([]domain.PublishedSnapshot, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	snaps, err := s.store.ListSnapshots(ctx, t.ID, strings.TrimSpace(slug))
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list snapshots")
	}
	return snaps, nil
}

// RollbackLayout creates a new immutable published version restoring the exact state of targetVersion.
func (s *StudioService) RollbackLayout(ctx context.Context, tenantSlug, slug string, targetVersion int, actor, comment string) (RollbackResult, error) {
	if targetVersion <= 0 {
		return RollbackResult{}, apperror.BadRequest("target version must be > 0")
	}
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return RollbackResult{}, err
	}
	cleanSlug := strings.TrimSpace(slug)
	targetSnap, err := s.store.GetSnapshotByVersion(ctx, t.ID, cleanSlug, targetVersion)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return RollbackResult{}, apperror.NotFound(fmt.Sprintf("snapshot v%d not found for %q", targetVersion, cleanSlug))
		}
		return RollbackResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load target snapshot")
	}
	layout, err := s.store.GetLayoutBySlug(ctx, t.ID, cleanSlug)
	if err != nil {
		return RollbackResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load layout")
	}

	now := time.Now().UTC()
	nextVersion := layout.PublishedVersion + 1
	if strings.TrimSpace(actor) == "" {
		actor = "studio-admin"
	}
	note := strings.TrimSpace(comment)
	if note == "" {
		note = fmt.Sprintf("Rollback to v%d", targetVersion)
	}

	restoredSnap := domain.PublishedSnapshot{
		ID:              "snap_" + uuid.NewString()[:8],
		TenantID:        t.ID,
		LayoutID:        layout.ID,
		LayoutSlug:      layout.Slug,
		Version:         nextVersion,
		Title:           targetSnap.Title,
		PageConfig:      targetSnap.PageConfig,
		Slots:           targetSnap.Slots,
		ResolvedWidgets: targetSnap.ResolvedWidgets,
		PublishedBy:     actor,
		Comment:         note,
		PublishedAt:     now,
	}
	if err := s.store.SaveSnapshot(ctx, restoredSnap); err != nil {
		return RollbackResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save rollback snapshot")
	}

	layout.Title = targetSnap.Title
	layout.PageConfig = targetSnap.PageConfig
	layout.Slots = targetSnap.Slots
	layout.PublishedVersion = nextVersion
	layout.DraftVersion++
	layout.Status = domain.LayoutStatusPublished
	layout.UpdatedAt = now
	savedLayout, err := s.store.UpsertLayout(ctx, layout)
	if err != nil {
		return RollbackResult{}, apperror.Wrap(err, apperror.CodeInternal, "failed to sync rolled-back draft layout")
	}
	s.recordAudit(ctx, t.ID, "layout", layout.Slug, "rollback", actor, fmt.Sprintf("Rolled back %s to v%d (published as v%d)", layout.Slug, targetVersion, nextVersion), nil)
	return RollbackResult{Snapshot: restoredSnap, Layout: savedLayout}, nil
}

// PromoteFallback composes the latest published snapshot with a neutral default context
// and persists it as the guaranteed-safe fallback page for the layout.
func (s *StudioService) PromoteFallback(ctx context.Context, tenantSlug, slug, actor string) (domain.FallbackLayout, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.FallbackLayout{}, err
	}
	cleanSlug := strings.TrimSpace(slug)
	snap, err := s.store.GetLatestSnapshot(ctx, t.ID, cleanSlug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrNoPublishedSnapshot) {
			return domain.FallbackLayout{}, apperror.BadRequest("publish the layout before setting a fallback")
		}
		return domain.FallbackLayout{}, apperror.Wrap(err, apperror.CodeInternal, "failed to load latest snapshot")
	}
	enabledSlots := make([]domain.LayoutSlot, 0, len(snap.Slots))
	for _, sl := range snap.Slots {
		if sl.Enabled && sl.RuleID == "" && sl.Schedule == nil {
			enabledSlots = append(enabledSlots, sl)
		}
	}
	if len(enabledSlots) == 0 {
		enabledSlots = snap.Slots
	}
	page, err := s.composer.Compose(snap.LayoutSlug, snap.Title, snap.PageConfig, enabledSlots, snap.ResolvedWidgets, domain.RequestContext{
		TenantSlug: t.Slug,
		City:       "Bengaluru",
		Segment:    "standard",
		Platform:   "kmp",
		Now:        time.Now().UTC(),
	})
	if err != nil {
		return domain.FallbackLayout{}, apperror.Wrap(err, apperror.CodeInternal, "failed to compose fallback page")
	}
	if strings.TrimSpace(actor) == "" {
		actor = "studio-admin"
	}
	fb := domain.FallbackLayout{
		ID:              "fb_" + snap.LayoutSlug,
		TenantID:        t.ID,
		LayoutSlug:      snap.LayoutSlug,
		SnapshotVersion: snap.Version,
		Page:            page,
		UpdatedBy:       actor,
		UpdatedAt:       time.Now().UTC(),
	}
	if err := s.store.SaveFallback(ctx, fb); err != nil {
		return domain.FallbackLayout{}, apperror.Wrap(err, apperror.CodeInternal, "failed to store fallback layout")
	}
	s.recordAudit(ctx, t.ID, "fallback", snap.LayoutSlug, "promote", actor, fmt.Sprintf("Promoted v%d as fallback for %s", snap.Version, snap.LayoutSlug), nil)
	return fb, nil
}

// ── Rules, Slotting, Splash, Tokens & ChangeLog ────────────────────────────────

func (s *StudioService) ListRules(ctx context.Context, tenantSlug string) ([]domain.TargetingRule, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListRules(ctx, t.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list rules")
	}
	return items, nil
}

func (s *StudioService) UpsertRule(ctx context.Context, tenantSlug string, rule domain.TargetingRule, actor string) (domain.TargetingRule, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.TargetingRule{}, err
	}
	rule.TenantID = t.ID
	rule.Key = strings.TrimSpace(rule.Key)
	rule.Name = strings.TrimSpace(rule.Name)
	if rule.Key == "" || rule.Name == "" {
		return domain.TargetingRule{}, apperror.BadRequest("rule key and name are required")
	}
	if rule.MatchMode == "" {
		rule.MatchMode = "all"
	}
	now := time.Now().UTC()
	if rule.ID == "" {
		rule.ID = "rule_" + uuid.NewString()[:8]
		rule.CreatedAt = now
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now
	saved, err := s.store.UpsertRule(ctx, rule)
	if err != nil {
		return domain.TargetingRule{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save targeting rule")
	}
	s.recordAudit(ctx, t.ID, "rule", saved.ID, "upsert", actor, fmt.Sprintf("Saved rule %s", saved.Key), nil)
	return saved, nil
}

func (s *StudioService) DeleteRule(ctx context.Context, tenantSlug, ruleID, actor string) error {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return err
	}
	if err := s.store.DeleteRule(ctx, t.ID, ruleID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperror.NotFound("rule not found")
		}
		return apperror.Wrap(err, apperror.CodeInternal, "failed to delete rule")
	}
	s.recordAudit(ctx, t.ID, "rule", ruleID, "delete", actor, fmt.Sprintf("Deleted rule %s", ruleID), nil)
	return nil
}

func (s *StudioService) ListSlottingRules(ctx context.Context, tenantSlug, layoutSlug string) ([]domain.SlottingRule, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListSlottingRules(ctx, t.ID, strings.TrimSpace(layoutSlug))
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to list slotting rules")
	}
	return items, nil
}

func (s *StudioService) UpsertSlottingRule(ctx context.Context, tenantSlug string, rule domain.SlottingRule, actor string) (domain.SlottingRule, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.SlottingRule{}, err
	}
	rule.TenantID = t.ID
	rule.LayoutSlug = strings.TrimSpace(rule.LayoutSlug)
	rule.WidgetID = strings.TrimSpace(rule.WidgetID)
	if rule.LayoutSlug == "" || rule.WidgetID == "" || strings.TrimSpace(rule.Name) == "" {
		return domain.SlottingRule{}, apperror.BadRequest("slotting rule name, layout_slug, and widget_id are required")
	}
	if rule.Action == "" {
		rule.Action = domain.SlottingActionInsert
	}
	now := time.Now().UTC()
	if rule.ID == "" {
		rule.ID = "slotrule_" + uuid.NewString()[:8]
		rule.CreatedAt = now
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now
	saved, err := s.store.UpsertSlottingRule(ctx, rule)
	if err != nil {
		return domain.SlottingRule{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save slotting rule")
	}
	s.recordAudit(ctx, t.ID, "slotting_rule", saved.ID, "upsert", actor, fmt.Sprintf("Saved slotting rule %s on %s", saved.Name, saved.LayoutSlug), nil)
	return saved, nil
}

func (s *StudioService) DeleteSlottingRule(ctx context.Context, tenantSlug, id, actor string) error {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return err
	}
	if err := s.store.DeleteSlottingRule(ctx, t.ID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperror.NotFound("slotting rule not found")
		}
		return apperror.Wrap(err, apperror.CodeInternal, "failed to delete slotting rule")
	}
	s.recordAudit(ctx, t.ID, "slotting_rule", id, "delete", actor, fmt.Sprintf("Deleted slotting rule %s", id), nil)
	return nil
}

func (s *StudioService) ListSplashConfigs(ctx context.Context, tenantSlug string) ([]domain.SplashConfig, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	return s.store.ListSplashConfigs(ctx, t.ID)
}

func (s *StudioService) UpsertSplashConfig(ctx context.Context, tenantSlug string, cfg domain.SplashConfig, actor string) (domain.SplashConfig, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.SplashConfig{}, err
	}
	cfg.TenantID = t.ID
	if cfg.ID == "" {
		cfg.ID = "splash_" + uuid.NewString()[:8]
	}
	if cfg.DurationMs <= 0 {
		cfg.DurationMs = 1500
	}
	cfg.UpdatedAt = time.Now().UTC()
	saved, err := s.store.UpsertSplashConfig(ctx, cfg)
	if err != nil {
		return domain.SplashConfig{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save splash config")
	}
	s.recordAudit(ctx, t.ID, "splash", saved.ID, "upsert", actor, fmt.Sprintf("Saved splash config %s", saved.Key), nil)
	return saved, nil
}

func (s *StudioService) ListTokenOverrides(ctx context.Context, tenantSlug string) ([]domain.TokenOverride, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	return s.store.ListTokenOverrides(ctx, t.ID)
}

func (s *StudioService) UpsertTokenOverride(ctx context.Context, tenantSlug string, ov domain.TokenOverride, actor string) (domain.TokenOverride, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return domain.TokenOverride{}, err
	}
	ov.TenantID = t.ID
	if ov.ID == "" {
		ov.ID = "tok_" + uuid.NewString()[:8]
	}
	ov.UpdatedAt = time.Now().UTC()
	saved, err := s.store.UpsertTokenOverride(ctx, ov)
	if err != nil {
		return domain.TokenOverride{}, apperror.Wrap(err, apperror.CodeInternal, "failed to save token override")
	}
	s.recordAudit(ctx, t.ID, "token_override", saved.ID, "upsert", actor, fmt.Sprintf("Saved token override %s", saved.Key), nil)
	return saved, nil
}

func (s *StudioService) ListChangeLogs(ctx context.Context, tenantSlug string, limit int) ([]domain.ChangeLogEntry, error) {
	t, err := s.mustTenant(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.store.ListChangeLogs(ctx, t.ID, limit)
}

func (s *StudioService) mustTenant(ctx context.Context, slug string) (domain.Tenant, error) {
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
	return t, nil
}

func (s *StudioService) widgetMap(ctx context.Context, tenantID string) (map[string]domain.Widget, error) {
	all, err := s.store.ListWidgets(ctx, tenantID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeInternal, "failed to load widgets")
	}
	out := make(map[string]domain.Widget, len(all))
	for _, w := range all {
		out[w.ID] = w
	}
	return out, nil
}

func (s *StudioService) recordAudit(ctx context.Context, tenantID, entityType, entityID, action, actor, summary string, diff []domain.DiffItem) {
	if strings.TrimSpace(actor) == "" {
		actor = "studio-admin"
	}
	_ = s.store.AppendChangeLog(ctx, domain.ChangeLogEntry{
		ID:         "chg_" + uuid.NewString()[:8],
		TenantID:   tenantID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Actor:      actor,
		Summary:    summary,
		Diff:       diff,
		CreatedAt:  time.Now().UTC(),
	})
}

func freezeReferencedWidgets(slots []domain.LayoutSlot, all map[string]domain.Widget) map[string]domain.Widget {
	frozen := make(map[string]domain.Widget, len(slots))
	var visit func(id string)
	visit = func(id string) {
		if id == "" {
			return
		}
		if _, done := frozen[id]; done {
			return
		}
		w, ok := all[id]
		if !ok {
			return
		}
		frozen[id] = w
		if rawIDs, ok := w.Props["child_widget_ids"].([]any); ok {
			for _, item := range rawIDs {
				if cid, ok := item.(string); ok {
					visit(cid)
				}
			}
		}
		if strIDs, ok := w.Props["child_widget_ids"].([]string); ok {
			for _, cid := range strIDs {
				visit(cid)
			}
		}
		if rawTabs, ok := w.Props["tabs"].([]any); ok {
			for _, tabRaw := range rawTabs {
				tabMap, ok := tabRaw.(map[string]any)
				if !ok {
					continue
				}
				if cids, ok := tabMap["widget_ids"].([]any); ok {
					for _, item := range cids {
						if cid, ok := item.(string); ok {
							visit(cid)
						}
					}
				}
			}
		}
	}
	for _, sl := range slots {
		visit(sl.WidgetID)
	}
	return frozen
}
