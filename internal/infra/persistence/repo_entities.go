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

// ── TenantStore ────────────────────────────────────────────────────────────────

// ListTenants returns all tenants ordered by slug.
func (s *SQLStore) ListTenants(ctx context.Context) ([]domain.Tenant, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, slug, name, description, active, created_at, updated_at
		FROM tenants ORDER BY slug ASC`))
	if err != nil {
		return nil, fmt.Errorf("persistence: list tenants: %w", err)
	}
	return collectRows(rows, scanTenant)
}

// GetTenantBySlug retrieves a tenant by its unique slug.
func (s *SQLStore) GetTenantBySlug(ctx context.Context, slug string) (domain.Tenant, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, slug, name, description, active, created_at, updated_at
		FROM tenants WHERE slug = ?`), slug)
	t, err := scanTenant(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tenant{}, fmt.Errorf("persistence: tenant slug %q: %w", slug, domain.ErrNotFound)
	}
	return t, err
}

// GetTenantByID retrieves a tenant by its primary ID.
func (s *SQLStore) GetTenantByID(ctx context.Context, id string) (domain.Tenant, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, slug, name, description, active, created_at, updated_at
		FROM tenants WHERE id = ?`), id)
	t, err := scanTenant(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tenant{}, fmt.Errorf("persistence: tenant id %q: %w", id, domain.ErrNotFound)
	}
	return t, err
}

// UpsertTenant creates or updates a tenant while preserving its original creation time.
func (s *SQLStore) UpsertTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(tenant.ID) == "" {
		if existing, err := s.GetTenantBySlug(ctx, tenant.Slug); err == nil {
			tenant.ID = existing.ID
			if tenant.CreatedAt.IsZero() {
				tenant.CreatedAt = existing.CreatedAt
			}
		} else {
			tenant.ID = "ten_" + uuid.NewString()[:8]
		}
	}
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = now
	}
	if tenant.UpdatedAt.IsZero() {
		tenant.UpdatedAt = now
	}

	var rawCreated any
	err := s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO tenants (id, slug, name, description, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			slug = excluded.slug, name = excluded.name, description = excluded.description,
			active = excluded.active, updated_at = excluded.updated_at
		RETURNING created_at`),
		tenant.ID, tenant.Slug, tenant.Name, tenant.Description, tenant.Active,
		s.formatTime(tenant.CreatedAt), s.formatTime(tenant.UpdatedAt),
	).Scan(&rawCreated)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Tenant{}, fmt.Errorf("persistence: tenant slug %q: %w", tenant.Slug, domain.ErrConflict)
		}
		return domain.Tenant{}, fmt.Errorf("persistence: upsert tenant %q: %w", tenant.Slug, err)
	}
	if parsed, err := parseTime(rawCreated); err == nil && !parsed.IsZero() {
		tenant.CreatedAt = parsed
	}
	tenant.UpdatedAt = tenant.UpdatedAt.UTC()
	return tenant, nil
}

func scanTenant(sc rowScanner) (domain.Tenant, error) {
	var (
		t                      domain.Tenant
		rawCreated, rawUpdated any
	)
	if err := sc.Scan(&t.ID, &t.Slug, &t.Name, &t.Description, &t.Active, &rawCreated, &rawUpdated); err != nil {
		return domain.Tenant{}, err
	}
	var err error
	if t.CreatedAt, t.UpdatedAt, err = parseTimestamps(rawCreated, rawUpdated); err != nil {
		return domain.Tenant{}, err
	}
	return t, nil
}

// ── TemplateStore ──────────────────────────────────────────────────────────────

// ListTemplates returns all widget templates for tenantID ordered by template key.
func (s *SQLStore) ListTemplates(ctx context.Context, tenantID string) ([]domain.WidgetTemplate, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, template_key, name, snippet_type, description,
		       default_props, schema_json, tags, created_at, updated_at
		FROM widget_templates WHERE tenant_id = ? ORDER BY template_key ASC`), tenantID)
	if err != nil {
		return nil, fmt.Errorf("persistence: list templates: %w", err)
	}
	return collectRows(rows, scanTemplate)
}

// GetTemplateByKey retrieves a widget template by tenant and unique key.
func (s *SQLStore) GetTemplateByKey(ctx context.Context, tenantID, key string) (domain.WidgetTemplate, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, tenant_id, template_key, name, snippet_type, description,
		       default_props, schema_json, tags, created_at, updated_at
		FROM widget_templates WHERE tenant_id = ? AND template_key = ?`), tenantID, key)
	tmpl, err := scanTemplate(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: template %q: %w", key, domain.ErrNotFound)
	}
	return tmpl, err
}

// UpsertTemplate creates or updates a reusable widget template keyed by (tenant_id, key).
func (s *SQLStore) UpsertTemplate(ctx context.Context, tmpl domain.WidgetTemplate) (domain.WidgetTemplate, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(tmpl.ID) == "" {
		tmpl.ID = "tmpl_" + uuid.NewString()[:8]
	}
	if tmpl.CreatedAt.IsZero() {
		tmpl.CreatedAt = now
	}
	if tmpl.UpdatedAt.IsZero() {
		tmpl.UpdatedAt = now
	}

	propsJSON, err := marshalJSON(tmpl.DefaultProps, "{}")
	if err != nil {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: marshal template default_props: %w", err)
	}
	schemaJSON, err := marshalJSON(tmpl.Schema, "{}")
	if err != nil {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: marshal template schema: %w", err)
	}
	tagsJSON, err := marshalJSON(tmpl.Tags, "[]")
	if err != nil {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: marshal template tags: %w", err)
	}

	var rawCreated any
	err = s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO widget_templates (
			id, tenant_id, template_key, name, snippet_type, description,
			default_props, schema_json, tags, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tenant_id, template_key) DO UPDATE SET
			name = excluded.name, snippet_type = excluded.snippet_type, description = excluded.description,
			default_props = excluded.default_props, schema_json = excluded.schema_json,
			tags = excluded.tags, updated_at = excluded.updated_at
		RETURNING id, created_at`),
		tmpl.ID, tmpl.TenantID, tmpl.Key, tmpl.Name, tmpl.SnippetType, tmpl.Description,
		propsJSON, schemaJSON, tagsJSON, s.formatTime(tmpl.CreatedAt), s.formatTime(tmpl.UpdatedAt),
	).Scan(&tmpl.ID, &rawCreated)
	if err != nil {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: upsert template %q: %w", tmpl.Key, err)
	}
	if parsed, err := parseTime(rawCreated); err == nil && !parsed.IsZero() {
		tmpl.CreatedAt = parsed
	}
	tmpl.UpdatedAt = tmpl.UpdatedAt.UTC()
	if tmpl.DefaultProps == nil {
		tmpl.DefaultProps = map[string]any{}
	}
	return tmpl, nil
}

func scanTemplate(sc rowScanner) (domain.WidgetTemplate, error) {
	var (
		tmpl                            domain.WidgetTemplate
		propsJSON, schemaJSON, tagsJSON string
		rawCreated, rawUpdated          any
	)
	if err := sc.Scan(
		&tmpl.ID, &tmpl.TenantID, &tmpl.Key, &tmpl.Name, &tmpl.SnippetType, &tmpl.Description,
		&propsJSON, &schemaJSON, &tagsJSON, &rawCreated, &rawUpdated,
	); err != nil {
		return domain.WidgetTemplate{}, err
	}
	if err := unmarshalJSON(propsJSON, &tmpl.DefaultProps); err != nil {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: decode template default_props: %w", err)
	}
	if tmpl.DefaultProps == nil {
		tmpl.DefaultProps = map[string]any{}
	}
	if err := unmarshalJSON(schemaJSON, &tmpl.Schema); err != nil {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: decode template schema: %w", err)
	}
	if len(tmpl.Schema) == 0 {
		tmpl.Schema = nil
	}
	if err := unmarshalJSON(tagsJSON, &tmpl.Tags); err != nil {
		return domain.WidgetTemplate{}, fmt.Errorf("persistence: decode template tags: %w", err)
	}
	var err error
	if tmpl.CreatedAt, tmpl.UpdatedAt, err = parseTimestamps(rawCreated, rawUpdated); err != nil {
		return domain.WidgetTemplate{}, err
	}
	return tmpl, nil
}

// ── WidgetStore ────────────────────────────────────────────────────────────────

// ListWidgets returns all widgets belonging to tenantID ordered by slug.
func (s *SQLStore) ListWidgets(ctx context.Context, tenantID string) ([]domain.Widget, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, slug, name, template_key, snippet_type,
		       props, tracking, motion, tags, active, created_at, updated_at
		FROM widgets WHERE tenant_id = ? ORDER BY slug ASC`), tenantID)
	if err != nil {
		return nil, fmt.Errorf("persistence: list widgets: %w", err)
	}
	return collectRows(rows, scanWidget)
}

// GetWidgetByID retrieves a widget by tenant and ID.
func (s *SQLStore) GetWidgetByID(ctx context.Context, tenantID, id string) (domain.Widget, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, tenant_id, slug, name, template_key, snippet_type,
		       props, tracking, motion, tags, active, created_at, updated_at
		FROM widgets WHERE tenant_id = ? AND id = ?`), tenantID, id)
	w, err := scanWidget(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Widget{}, fmt.Errorf("persistence: widget %q: %w", id, domain.ErrNotFound)
	}
	return w, err
}

// GetWidgetsByIDs batch-loads widgets by IDs for a tenant, returning a map keyed by widget ID.
func (s *SQLStore) GetWidgetsByIDs(ctx context.Context, tenantID string, ids []string) (map[string]domain.Widget, error) {
	result := make(map[string]domain.Widget, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			unique = append(unique, id)
		}
	}
	if len(unique) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(unique))
	args := make([]any, 0, len(unique)+1)
	args = append(args, tenantID)
	for i, id := range unique {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, slug, name, template_key, snippet_type,
		       props, tracking, motion, tags, active, created_at, updated_at
		FROM widgets WHERE tenant_id = ? AND id IN (%s)`, strings.Join(placeholders, ", "))

	rows, err := s.db.QueryContext(ctx, s.rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("persistence: get widgets by ids: %w", err)
	}
	items, err := collectRows(rows, scanWidget)
	if err != nil {
		return nil, err
	}
	for _, w := range items {
		result[w.ID] = w
	}
	return result, nil
}

// UpsertWidget creates or updates a configured widget.
func (s *SQLStore) UpsertWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(widget.ID) == "" {
		widget.ID = "w_" + uuid.NewString()[:8]
	}
	if widget.CreatedAt.IsZero() {
		widget.CreatedAt = now
	}
	if widget.UpdatedAt.IsZero() {
		widget.UpdatedAt = now
	}

	propsJSON, err := marshalJSON(widget.Props, "{}")
	if err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: marshal widget props: %w", err)
	}
	trackingJSON, err := marshalNullableJSON(widget.Tracking)
	if err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: marshal widget tracking: %w", err)
	}
	motionJSON, err := marshalNullableJSON(widget.Motion)
	if err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: marshal widget motion: %w", err)
	}
	tagsJSON, err := marshalJSON(widget.Tags, "[]")
	if err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: marshal widget tags: %w", err)
	}

	var rawCreated any
	err = s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO widgets (
			id, tenant_id, slug, name, template_key, snippet_type,
			props, tracking, motion, tags, active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			slug = excluded.slug, name = excluded.name, template_key = excluded.template_key,
			snippet_type = excluded.snippet_type, props = excluded.props, tracking = excluded.tracking,
			motion = excluded.motion, tags = excluded.tags, active = excluded.active, updated_at = excluded.updated_at
		RETURNING created_at`),
		widget.ID, widget.TenantID, widget.Slug, widget.Name, widget.TemplateKey, widget.SnippetType,
		propsJSON, trackingJSON, motionJSON, tagsJSON, widget.Active,
		s.formatTime(widget.CreatedAt), s.formatTime(widget.UpdatedAt),
	).Scan(&rawCreated)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Widget{}, fmt.Errorf("persistence: widget slug %q: %w", widget.Slug, domain.ErrConflict)
		}
		return domain.Widget{}, fmt.Errorf("persistence: upsert widget %q: %w", widget.ID, err)
	}
	if parsed, err := parseTime(rawCreated); err == nil && !parsed.IsZero() {
		widget.CreatedAt = parsed
	}
	widget.UpdatedAt = widget.UpdatedAt.UTC()
	if widget.Props == nil {
		widget.Props = map[string]any{}
	}
	return widget, nil
}

// DeleteWidget deletes a widget by tenant and ID, returning domain.ErrNotFound if absent.
func (s *SQLStore) DeleteWidget(ctx context.Context, tenantID, id string) error {
	return s.execDelete(ctx, fmt.Sprintf("widget %q", id),
		`DELETE FROM widgets WHERE tenant_id = ? AND id = ?`, tenantID, id)
}

func scanWidget(sc rowScanner) (domain.Widget, error) {
	var (
		w                        domain.Widget
		propsJSON, tagsJSON      string
		trackingJSON, motionJSON sql.NullString
		rawCreated, rawUpdated   any
	)
	if err := sc.Scan(
		&w.ID, &w.TenantID, &w.Slug, &w.Name, &w.TemplateKey, &w.SnippetType,
		&propsJSON, &trackingJSON, &motionJSON, &tagsJSON, &w.Active, &rawCreated, &rawUpdated,
	); err != nil {
		return domain.Widget{}, err
	}
	if err := unmarshalJSON(propsJSON, &w.Props); err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: decode widget props: %w", err)
	}
	if w.Props == nil {
		w.Props = map[string]any{}
	}
	if err := unmarshalNullableJSON(trackingJSON, &w.Tracking); err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: decode widget tracking: %w", err)
	}
	if err := unmarshalNullableJSON(motionJSON, &w.Motion); err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: decode widget motion: %w", err)
	}
	if err := unmarshalJSON(tagsJSON, &w.Tags); err != nil {
		return domain.Widget{}, fmt.Errorf("persistence: decode widget tags: %w", err)
	}
	var err error
	if w.CreatedAt, w.UpdatedAt, err = parseTimestamps(rawCreated, rawUpdated); err != nil {
		return domain.Widget{}, err
	}
	return w, nil
}

// ── RuleStore ──────────────────────────────────────────────────────────────────

// ListRules returns all targeting rules for tenantID ordered by priority descending.
func (s *SQLStore) ListRules(ctx context.Context, tenantID string) ([]domain.TargetingRule, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, rule_key, name, description, match_mode,
		       conditions, schedule, priority, enabled, created_at, updated_at
		FROM targeting_rules WHERE tenant_id = ? ORDER BY priority DESC, rule_key ASC`), tenantID)
	if err != nil {
		return nil, fmt.Errorf("persistence: list rules: %w", err)
	}
	return collectRows(rows, scanRule)
}

// GetRuleByID retrieves a targeting rule by tenant and ID.
func (s *SQLStore) GetRuleByID(ctx context.Context, tenantID, id string) (domain.TargetingRule, error) {
	row := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, tenant_id, rule_key, name, description, match_mode,
		       conditions, schedule, priority, enabled, created_at, updated_at
		FROM targeting_rules WHERE tenant_id = ? AND id = ?`), tenantID, id)
	r, err := scanRule(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.TargetingRule{}, fmt.Errorf("persistence: rule %q: %w", id, domain.ErrNotFound)
	}
	return r, err
}

// UpsertRule creates or updates a targeting rule.
func (s *SQLStore) UpsertRule(ctx context.Context, rule domain.TargetingRule) (domain.TargetingRule, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(rule.ID) == "" {
		rule.ID = "rule_" + uuid.NewString()[:8]
	}
	if strings.TrimSpace(rule.MatchMode) == "" {
		rule.MatchMode = "all"
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = now
	}

	condsJSON, err := marshalJSON(rule.Conditions, "[]")
	if err != nil {
		return domain.TargetingRule{}, fmt.Errorf("persistence: marshal rule conditions: %w", err)
	}
	schedJSON, err := marshalNullableJSON(rule.Schedule)
	if err != nil {
		return domain.TargetingRule{}, fmt.Errorf("persistence: marshal rule schedule: %w", err)
	}

	var rawCreated any
	err = s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO targeting_rules (
			id, tenant_id, rule_key, name, description, match_mode,
			conditions, schedule, priority, enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			rule_key = excluded.rule_key, name = excluded.name, description = excluded.description,
			match_mode = excluded.match_mode, conditions = excluded.conditions, schedule = excluded.schedule,
			priority = excluded.priority, enabled = excluded.enabled, updated_at = excluded.updated_at
		RETURNING created_at`),
		rule.ID, rule.TenantID, rule.Key, rule.Name, rule.Description, rule.MatchMode,
		condsJSON, schedJSON, rule.Priority, rule.Enabled,
		s.formatTime(rule.CreatedAt), s.formatTime(rule.UpdatedAt),
	).Scan(&rawCreated)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.TargetingRule{}, fmt.Errorf("persistence: rule key %q: %w", rule.Key, domain.ErrConflict)
		}
		return domain.TargetingRule{}, fmt.Errorf("persistence: upsert rule %q: %w", rule.ID, err)
	}
	if parsed, err := parseTime(rawCreated); err == nil && !parsed.IsZero() {
		rule.CreatedAt = parsed
	}
	rule.UpdatedAt = rule.UpdatedAt.UTC()
	if rule.Conditions == nil {
		rule.Conditions = []domain.RuleCondition{}
	}
	return rule, nil
}

// DeleteRule removes a targeting rule by tenant and ID, returning domain.ErrNotFound if absent.
func (s *SQLStore) DeleteRule(ctx context.Context, tenantID, id string) error {
	return s.execDelete(ctx, fmt.Sprintf("rule %q", id),
		`DELETE FROM targeting_rules WHERE tenant_id = ? AND id = ?`, tenantID, id)
}

func scanRule(sc rowScanner) (domain.TargetingRule, error) {
	var (
		r                      domain.TargetingRule
		condsJSON              string
		schedJSON              sql.NullString
		rawCreated, rawUpdated any
	)
	if err := sc.Scan(
		&r.ID, &r.TenantID, &r.Key, &r.Name, &r.Description, &r.MatchMode,
		&condsJSON, &schedJSON, &r.Priority, &r.Enabled, &rawCreated, &rawUpdated,
	); err != nil {
		return domain.TargetingRule{}, err
	}
	if err := unmarshalJSON(condsJSON, &r.Conditions); err != nil {
		return domain.TargetingRule{}, fmt.Errorf("persistence: decode rule conditions: %w", err)
	}
	if r.Conditions == nil {
		r.Conditions = []domain.RuleCondition{}
	}
	if schedJSON.Valid {
		var w domain.ScheduleWindow
		if err := unmarshalJSON(schedJSON.String, &w); err != nil {
			return domain.TargetingRule{}, fmt.Errorf("persistence: decode rule schedule: %w", err)
		}
		r.Schedule = &w
	}
	var err error
	if r.CreatedAt, r.UpdatedAt, err = parseTimestamps(rawCreated, rawUpdated); err != nil {
		return domain.TargetingRule{}, err
	}
	return r, nil
}
