package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/google/uuid"
)

// ── LayoutStore ────────────────────────────────────────────────────────────────

// ListLayouts returns all layouts belonging to tenantID ordered by slug.
func (s *SQLStore) ListLayouts(ctx context.Context, tenantID string) ([]domain.Layout, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, slug, title, description, status,
		       published_version, draft_version, page_config, slots,
		       tags, default_rule_id, created_at, updated_at
		FROM layouts WHERE tenant_id = ? ORDER BY slug ASC`), tenantID)
	if err != nil {
		return nil, fmt.Errorf("persistence: list layouts: %w", err)
	}
	return collectRows(rows, scanLayout)
}

// GetLayoutBySlug retrieves a layout draft/published record by tenant and slug.
func (s *SQLStore) GetLayoutBySlug(ctx context.Context, tenantID, slug string) (domain.Layout, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, tenant_id, slug, title, description, status,
		       published_version, draft_version, page_config, slots,
		       tags, default_rule_id, created_at, updated_at
		FROM layouts WHERE tenant_id = ? AND slug = ?`), tenantID, slug)
	l, err := scanLayout(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Layout{}, fmt.Errorf("persistence: layout %q: %w", slug, domain.ErrNotFound)
	}
	return l, err
}

// UpsertLayout creates or updates a layout keyed by (tenant_id, slug).
func (s *SQLStore) UpsertLayout(ctx context.Context, layout domain.Layout) (domain.Layout, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(layout.ID) == "" {
		layout.ID = "lay_" + uuid.NewString()[:8]
	}
	if layout.Status == "" {
		layout.Status = domain.LayoutStatusDraft
	}
	if layout.DraftVersion <= 0 {
		layout.DraftVersion = 1
	}
	if layout.CreatedAt.IsZero() {
		layout.CreatedAt = now
	}
	if layout.UpdatedAt.IsZero() {
		layout.UpdatedAt = now
	}

	pageCfgJSON, err := marshalJSON(layout.PageConfig, "{}")
	if err != nil {
		return domain.Layout{}, fmt.Errorf("persistence: marshal layout page_config: %w", err)
	}
	slotsJSON, err := marshalJSON(layout.Slots, "[]")
	if err != nil {
		return domain.Layout{}, fmt.Errorf("persistence: marshal layout slots: %w", err)
	}
	tagsJSON, err := marshalJSON(layout.Tags, "[]")
	if err != nil {
		return domain.Layout{}, fmt.Errorf("persistence: marshal layout tags: %w", err)
	}

	var rawCreated any
	err = s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO layouts (
			id, tenant_id, slug, title, description, status,
			published_version, draft_version, page_config, slots,
			tags, default_rule_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tenant_id, slug) DO UPDATE SET
			title = excluded.title, description = excluded.description, status = excluded.status,
			published_version = excluded.published_version, draft_version = excluded.draft_version,
			page_config = excluded.page_config, slots = excluded.slots, tags = excluded.tags,
			default_rule_id = excluded.default_rule_id, updated_at = excluded.updated_at
		RETURNING id, created_at`),
		layout.ID, layout.TenantID, layout.Slug, layout.Title, layout.Description, string(layout.Status),
		layout.PublishedVersion, layout.DraftVersion, pageCfgJSON, slotsJSON, tagsJSON, layout.DefaultRuleID,
		s.formatTime(layout.CreatedAt), s.formatTime(layout.UpdatedAt),
	).Scan(&layout.ID, &rawCreated)
	if err != nil {
		return domain.Layout{}, fmt.Errorf("persistence: upsert layout %q: %w", layout.Slug, err)
	}
	if parsed, err := parseTime(rawCreated); err == nil && !parsed.IsZero() {
		layout.CreatedAt = parsed
	}
	layout.UpdatedAt = layout.UpdatedAt.UTC()
	if layout.Slots == nil {
		layout.Slots = []domain.LayoutSlot{}
	}
	return layout, nil
}

// DeleteLayout removes a layout by tenant and slug, returning domain.ErrNotFound if absent.
func (s *SQLStore) DeleteLayout(ctx context.Context, tenantID, slug string) error {
	return s.execDelete(ctx, fmt.Sprintf("layout %q", slug),
		`DELETE FROM layouts WHERE tenant_id = ? AND slug = ?`, tenantID, slug)
}

func scanLayout(sc rowScanner) (domain.Layout, error) {
	var (
		l                                         domain.Layout
		statusStr, pageCfgJSON, slotsJSON, tagStr string
		rawCreated, rawUpdated                    any
	)
	if err := sc.Scan(
		&l.ID, &l.TenantID, &l.Slug, &l.Title, &l.Description, &statusStr,
		&l.PublishedVersion, &l.DraftVersion, &pageCfgJSON, &slotsJSON,
		&tagStr, &l.DefaultRuleID, &rawCreated, &rawUpdated,
	); err != nil {
		return domain.Layout{}, err
	}
	l.Status = domain.LayoutStatus(statusStr)
	if err := unmarshalJSON(pageCfgJSON, &l.PageConfig); err != nil {
		return domain.Layout{}, fmt.Errorf("persistence: decode layout page_config: %w", err)
	}
	if err := unmarshalJSON(slotsJSON, &l.Slots); err != nil {
		return domain.Layout{}, fmt.Errorf("persistence: decode layout slots: %w", err)
	}
	if l.Slots == nil {
		l.Slots = []domain.LayoutSlot{}
	}
	if err := unmarshalJSON(tagStr, &l.Tags); err != nil {
		return domain.Layout{}, fmt.Errorf("persistence: decode layout tags: %w", err)
	}
	var err error
	if l.CreatedAt, l.UpdatedAt, err = parseTimestamps(rawCreated, rawUpdated); err != nil {
		return domain.Layout{}, err
	}
	return l, nil
}

// ── SlottingStore ──────────────────────────────────────────────────────────────

// ListSlottingRules returns dynamic slotting rules ordered by priority descending.
// If layoutSlug is empty, all slotting rules for tenantID are returned.
func (s *SQLStore) ListSlottingRules(ctx context.Context, tenantID, layoutSlug string) ([]domain.SlottingRule, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if strings.TrimSpace(layoutSlug) == "" {
		rows, err = s.db.QueryContext(ctx, s.rebind(`
			SELECT id, tenant_id, layout_slug, name, action, target_position,
			       widget_id, rule_id, schedule, priority, enabled, created_at, updated_at
			FROM slotting_rules WHERE tenant_id = ?
			ORDER BY priority DESC, created_at ASC, id ASC`), tenantID)
	} else {
		rows, err = s.db.QueryContext(ctx, s.rebind(`
			SELECT id, tenant_id, layout_slug, name, action, target_position,
			       widget_id, rule_id, schedule, priority, enabled, created_at, updated_at
			FROM slotting_rules WHERE tenant_id = ? AND layout_slug = ?
			ORDER BY priority DESC, created_at ASC, id ASC`), tenantID, layoutSlug)
	}
	if err != nil {
		return nil, fmt.Errorf("persistence: list slotting rules: %w", err)
	}
	return collectRows(rows, scanSlottingRule)
}

// UpsertSlottingRule creates or updates a dynamic slotting rule.
func (s *SQLStore) UpsertSlottingRule(ctx context.Context, rule domain.SlottingRule) (domain.SlottingRule, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(rule.ID) == "" {
		rule.ID = "slotrule_" + uuid.NewString()[:8]
	}
	if rule.Action == "" {
		rule.Action = domain.SlottingActionInsert
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = now
	}

	schedJSON, err := marshalNullableJSON(rule.Schedule)
	if err != nil {
		return domain.SlottingRule{}, fmt.Errorf("persistence: marshal slotting schedule: %w", err)
	}

	var rawCreated any
	err = s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO slotting_rules (
			id, tenant_id, layout_slug, name, action, target_position,
			widget_id, rule_id, schedule, priority, enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			layout_slug = excluded.layout_slug, name = excluded.name, action = excluded.action,
			target_position = excluded.target_position, widget_id = excluded.widget_id,
			rule_id = excluded.rule_id, schedule = excluded.schedule, priority = excluded.priority,
			enabled = excluded.enabled, updated_at = excluded.updated_at
		RETURNING created_at`),
		rule.ID, rule.TenantID, rule.LayoutSlug, rule.Name, string(rule.Action), rule.TargetPosition,
		rule.WidgetID, rule.RuleID, schedJSON, rule.Priority, rule.Enabled,
		s.formatTime(rule.CreatedAt), s.formatTime(rule.UpdatedAt),
	).Scan(&rawCreated)
	if err != nil {
		return domain.SlottingRule{}, fmt.Errorf("persistence: upsert slotting rule %q: %w", rule.ID, err)
	}
	if parsed, err := parseTime(rawCreated); err == nil && !parsed.IsZero() {
		rule.CreatedAt = parsed
	}
	rule.UpdatedAt = rule.UpdatedAt.UTC()
	return rule, nil
}

// DeleteSlottingRule removes a slotting rule by tenant and ID, returning domain.ErrNotFound if absent.
func (s *SQLStore) DeleteSlottingRule(ctx context.Context, tenantID, id string) error {
	return s.execDelete(ctx, fmt.Sprintf("slotting rule %q", id),
		`DELETE FROM slotting_rules WHERE tenant_id = ? AND id = ?`, tenantID, id)
}

func scanSlottingRule(sc rowScanner) (domain.SlottingRule, error) {
	var (
		r                      domain.SlottingRule
		actionStr              string
		schedJSON              sql.NullString
		rawCreated, rawUpdated any
	)
	if err := sc.Scan(
		&r.ID, &r.TenantID, &r.LayoutSlug, &r.Name, &actionStr, &r.TargetPosition,
		&r.WidgetID, &r.RuleID, &schedJSON, &r.Priority, &r.Enabled, &rawCreated, &rawUpdated,
	); err != nil {
		return domain.SlottingRule{}, err
	}
	r.Action = domain.SlottingAction(actionStr)
	if schedJSON.Valid {
		var w domain.ScheduleWindow
		if err := unmarshalJSON(schedJSON.String, &w); err != nil {
			return domain.SlottingRule{}, fmt.Errorf("persistence: decode slotting schedule: %w", err)
		}
		r.Schedule = &w
	}
	var err error
	if r.CreatedAt, r.UpdatedAt, err = parseTimestamps(rawCreated, rawUpdated); err != nil {
		return domain.SlottingRule{}, err
	}
	return r, nil
}

// ── SnapshotStore ──────────────────────────────────────────────────────────────

// SaveSnapshot stores a new immutable published layout snapshot version.
// Attempting to overwrite an existing (tenant_id, layout_slug, version) returns domain.ErrConflict.
func (s *SQLStore) SaveSnapshot(ctx context.Context, snapshot domain.PublishedSnapshot) error {
	if strings.TrimSpace(snapshot.ID) == "" {
		snapshot.ID = "snap_" + uuid.NewString()[:8]
	}
	if snapshot.PublishedAt.IsZero() {
		snapshot.PublishedAt = time.Now().UTC()
	}

	pageCfgJSON, err := marshalJSON(snapshot.PageConfig, "{}")
	if err != nil {
		return fmt.Errorf("persistence: marshal snapshot page_config: %w", err)
	}
	slotsJSON, err := marshalJSON(snapshot.Slots, "[]")
	if err != nil {
		return fmt.Errorf("persistence: marshal snapshot slots: %w", err)
	}
	widgetsJSON, err := marshalJSON(snapshot.ResolvedWidgets, "{}")
	if err != nil {
		return fmt.Errorf("persistence: marshal snapshot resolved_widgets: %w", err)
	}

	_, err = s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO published_snapshots (
			id, tenant_id, layout_id, layout_slug, version, title,
			page_config, slots, resolved_widgets, published_by, comment, published_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		snapshot.ID, snapshot.TenantID, snapshot.LayoutID, snapshot.LayoutSlug, snapshot.Version, snapshot.Title,
		pageCfgJSON, slotsJSON, widgetsJSON, snapshot.PublishedBy, snapshot.Comment, s.formatTime(snapshot.PublishedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("persistence: snapshot %s v%d: %w", snapshot.LayoutSlug, snapshot.Version, domain.ErrConflict)
		}
		return fmt.Errorf("persistence: save snapshot %s v%d: %w", snapshot.LayoutSlug, snapshot.Version, err)
	}
	return nil
}

// GetLatestSnapshot returns the highest-version published snapshot for (tenantID, layoutSlug).
func (s *SQLStore) GetLatestSnapshot(ctx context.Context, tenantID, layoutSlug string) (domain.PublishedSnapshot, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, tenant_id, layout_id, layout_slug, version, title,
		       page_config, slots, resolved_widgets, published_by, comment, published_at
		FROM published_snapshots
		WHERE tenant_id = ? AND layout_slug = ?
		ORDER BY version DESC LIMIT 1`), tenantID, layoutSlug)
	snap, err := scanSnapshot(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PublishedSnapshot{}, fmt.Errorf(
			"persistence: snapshot for %q: %w: %w",
			layoutSlug, domain.ErrNoPublishedSnapshot, domain.ErrNotFound,
		)
	}
	return snap, err
}

// GetSnapshotByVersion retrieves an exact immutable snapshot version.
func (s *SQLStore) GetSnapshotByVersion(ctx context.Context, tenantID, layoutSlug string, version int) (domain.PublishedSnapshot, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, tenant_id, layout_id, layout_slug, version, title,
		       page_config, slots, resolved_widgets, published_by, comment, published_at
		FROM published_snapshots
		WHERE tenant_id = ? AND layout_slug = ? AND version = ?`), tenantID, layoutSlug, version)
	snap, err := scanSnapshot(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PublishedSnapshot{}, fmt.Errorf("persistence: snapshot %s v%d: %w", layoutSlug, version, domain.ErrNotFound)
	}
	return snap, err
}

// ListSnapshots returns all published snapshots for (tenantID, layoutSlug), newest version first.
func (s *SQLStore) ListSnapshots(ctx context.Context, tenantID, layoutSlug string) ([]domain.PublishedSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, layout_id, layout_slug, version, title,
		       page_config, slots, resolved_widgets, published_by, comment, published_at
		FROM published_snapshots
		WHERE tenant_id = ? AND layout_slug = ?
		ORDER BY version DESC`), tenantID, layoutSlug)
	if err != nil {
		return nil, fmt.Errorf("persistence: list snapshots: %w", err)
	}
	return collectRows(rows, scanSnapshot)
}

func scanSnapshot(sc rowScanner) (domain.PublishedSnapshot, error) {
	var (
		snap                              domain.PublishedSnapshot
		pageCfgJSON, slotsJSON, widgetStr string
		rawPublished                      any
	)
	if err := sc.Scan(
		&snap.ID, &snap.TenantID, &snap.LayoutID, &snap.LayoutSlug, &snap.Version, &snap.Title,
		&pageCfgJSON, &slotsJSON, &widgetStr, &snap.PublishedBy, &snap.Comment, &rawPublished,
	); err != nil {
		return domain.PublishedSnapshot{}, err
	}
	if err := unmarshalJSON(pageCfgJSON, &snap.PageConfig); err != nil {
		return domain.PublishedSnapshot{}, fmt.Errorf("persistence: decode snapshot page_config: %w", err)
	}
	if err := unmarshalJSON(slotsJSON, &snap.Slots); err != nil {
		return domain.PublishedSnapshot{}, fmt.Errorf("persistence: decode snapshot slots: %w", err)
	}
	if snap.Slots == nil {
		snap.Slots = []domain.LayoutSlot{}
	}
	if err := unmarshalJSON(widgetStr, &snap.ResolvedWidgets); err != nil {
		return domain.PublishedSnapshot{}, fmt.Errorf("persistence: decode snapshot resolved_widgets: %w", err)
	}
	if snap.ResolvedWidgets == nil {
		snap.ResolvedWidgets = map[string]domain.Widget{}
	}
	var err error
	if snap.PublishedAt, err = parseTime(rawPublished); err != nil {
		return domain.PublishedSnapshot{}, err
	}
	return snap, nil
}

// ── FallbackStore ──────────────────────────────────────────────────────────────

// SaveFallback upserts the pre-validated fallback page for (tenant_id, layout_slug).
func (s *SQLStore) SaveFallback(ctx context.Context, fb domain.FallbackLayout) error {
	if strings.TrimSpace(fb.ID) == "" {
		fb.ID = "fb_" + uuid.NewString()[:8]
	}
	if fb.UpdatedAt.IsZero() {
		fb.UpdatedAt = time.Now().UTC()
	}

	pageJSON, err := marshalJSON(fb.Page, "{}")
	if err != nil {
		return fmt.Errorf("persistence: marshal fallback page: %w", err)
	}

	_, err = s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO fallback_layouts (
			id, tenant_id, layout_slug, snapshot_version, page_json, updated_by, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tenant_id, layout_slug) DO UPDATE SET
			snapshot_version = excluded.snapshot_version,
			page_json = excluded.page_json,
			updated_by = excluded.updated_by,
			updated_at = excluded.updated_at`),
		fb.ID, fb.TenantID, fb.LayoutSlug, fb.SnapshotVersion, pageJSON, fb.UpdatedBy, s.formatTime(fb.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("persistence: save fallback %q: %w", fb.LayoutSlug, err)
	}
	return nil
}

// GetFallback retrieves the pre-validated fallback page for (tenantID, layoutSlug).
func (s *SQLStore) GetFallback(ctx context.Context, tenantID, layoutSlug string) (domain.FallbackLayout, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, tenant_id, layout_slug, snapshot_version, page_json, updated_by, updated_at
		FROM fallback_layouts WHERE tenant_id = ? AND layout_slug = ?`), tenantID, layoutSlug)

	var (
		fb         domain.FallbackLayout
		pageJSON   string
		rawUpdated any
	)
	err := row.Scan(
		&fb.ID, &fb.TenantID, &fb.LayoutSlug, &fb.SnapshotVersion,
		&pageJSON, &fb.UpdatedBy, &rawUpdated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.FallbackLayout{}, fmt.Errorf("persistence: fallback %q: %w", layoutSlug, domain.ErrNotFound)
	}
	if err != nil {
		return domain.FallbackLayout{}, fmt.Errorf("persistence: get fallback %q: %w", layoutSlug, err)
	}
	if err := unmarshalJSON(pageJSON, &fb.Page); err != nil {
		return domain.FallbackLayout{}, fmt.Errorf("persistence: decode fallback page: %w", err)
	}
	if fb.UpdatedAt, err = parseTime(rawUpdated); err != nil {
		return domain.FallbackLayout{}, err
	}
	return fb, nil
}
