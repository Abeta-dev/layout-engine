package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/google/uuid"
)

// ── ConfigStore (SplashConfig & TokenOverride) ─────────────────────────────────

// ListSplashConfigs returns all splash configurations for tenantID,
// with rule-targeted configs ordered before default fallbacks.
func (s *SQLStore) ListSplashConfigs(ctx context.Context, tenantID string) ([]domain.SplashConfig, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, config_key, image_url, background_color_hex,
		       duration_ms, lottie_url, rule_id, schedule, is_default, enabled, updated_at
		FROM splash_configs
		WHERE tenant_id = ?
		ORDER BY is_default ASC, config_key ASC`), tenantID)
	if err != nil {
		return nil, fmt.Errorf("persistence: list splash configs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.SplashConfig, 0)
	for rows.Next() {
		cfg, err := scanSplashConfig(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cfg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("persistence: iterate splash configs: %w", err)
	}
	return out, nil
}

// UpsertSplashConfig creates or updates a splash configuration keyed by (tenant_id, key).
func (s *SQLStore) UpsertSplashConfig(ctx context.Context, cfg domain.SplashConfig) (domain.SplashConfig, error) {
	if strings.TrimSpace(cfg.ID) == "" {
		cfg.ID = "splash_" + uuid.NewString()[:8]
	}
	if cfg.DurationMs <= 0 {
		cfg.DurationMs = 1500
	}
	if cfg.UpdatedAt.IsZero() {
		cfg.UpdatedAt = time.Now().UTC()
	}

	schedJSON, err := marshalNullableJSON(cfg.Schedule)
	if err != nil {
		return domain.SplashConfig{}, fmt.Errorf("persistence: marshal splash schedule: %w", err)
	}

	err = s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO splash_configs (
			id, tenant_id, config_key, image_url, background_color_hex,
			duration_ms, lottie_url, rule_id, schedule, is_default, enabled, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tenant_id, config_key) DO UPDATE SET
			image_url            = excluded.image_url,
			background_color_hex = excluded.background_color_hex,
			duration_ms          = excluded.duration_ms,
			lottie_url           = excluded.lottie_url,
			rule_id              = excluded.rule_id,
			schedule             = excluded.schedule,
			is_default           = excluded.is_default,
			enabled              = excluded.enabled,
			updated_at           = excluded.updated_at
		RETURNING id`),
		cfg.ID,
		cfg.TenantID,
		cfg.Key,
		cfg.ImageURL,
		cfg.BackgroundColorHex,
		cfg.DurationMs,
		cfg.LottieURL,
		cfg.RuleID,
		schedJSON,
		cfg.IsDefault,
		cfg.Enabled,
		s.formatTime(cfg.UpdatedAt),
	).Scan(&cfg.ID)
	if err != nil {
		return domain.SplashConfig{}, fmt.Errorf("persistence: upsert splash config %q: %w", cfg.Key, err)
	}
	cfg.UpdatedAt = cfg.UpdatedAt.UTC()
	return cfg, nil
}

func scanSplashConfig(sc rowScanner) (domain.SplashConfig, error) {
	var (
		cfg        domain.SplashConfig
		schedJSON  sql.NullString
		rawUpdated any
	)
	if err := sc.Scan(
		&cfg.ID,
		&cfg.TenantID,
		&cfg.Key,
		&cfg.ImageURL,
		&cfg.BackgroundColorHex,
		&cfg.DurationMs,
		&cfg.LottieURL,
		&cfg.RuleID,
		&schedJSON,
		&cfg.IsDefault,
		&cfg.Enabled,
		&rawUpdated,
	); err != nil {
		return domain.SplashConfig{}, err
	}
	if schedJSON.Valid {
		var w domain.ScheduleWindow
		if err := unmarshalJSON(schedJSON.String, &w); err != nil {
			return domain.SplashConfig{}, fmt.Errorf("persistence: decode splash schedule: %w", err)
		}
		cfg.Schedule = &w
	}
	var err error
	if cfg.UpdatedAt, err = parseTime(rawUpdated); err != nil {
		return domain.SplashConfig{}, err
	}
	return cfg, nil
}

// ListTokenOverrides returns all design-token overrides for tenantID,
// placing unscoped defaults before rule-targeted overrides so matching rules win cleanly.
func (s *SQLStore) ListTokenOverrides(ctx context.Context, tenantID string) ([]domain.TokenOverride, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, config_key, theme_preset, tokens, rule_id, enabled, updated_at
		FROM token_overrides
		WHERE tenant_id = ?
		ORDER BY CASE WHEN rule_id = '' THEN 0 ELSE 1 END ASC, config_key ASC`), tenantID)
	if err != nil {
		return nil, fmt.Errorf("persistence: list token overrides: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.TokenOverride, 0)
	for rows.Next() {
		ov, err := scanTokenOverride(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ov)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("persistence: iterate token overrides: %w", err)
	}
	return out, nil
}

// UpsertTokenOverride creates or updates a token override keyed by (tenant_id, key).
func (s *SQLStore) UpsertTokenOverride(ctx context.Context, ov domain.TokenOverride) (domain.TokenOverride, error) {
	if strings.TrimSpace(ov.ID) == "" {
		ov.ID = "tok_" + uuid.NewString()[:8]
	}
	if strings.TrimSpace(ov.ThemePreset) == "" {
		ov.ThemePreset = "default"
	}
	if ov.UpdatedAt.IsZero() {
		ov.UpdatedAt = time.Now().UTC()
	}

	tokensJSON, err := marshalJSON(ov.Tokens, "{}")
	if err != nil {
		return domain.TokenOverride{}, fmt.Errorf("persistence: marshal token overrides: %w", err)
	}

	err = s.db.QueryRowContext(ctx, s.rebind(`
		INSERT INTO token_overrides (
			id, tenant_id, config_key, theme_preset, tokens, rule_id, enabled, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tenant_id, config_key) DO UPDATE SET
			theme_preset = excluded.theme_preset,
			tokens       = excluded.tokens,
			rule_id      = excluded.rule_id,
			enabled      = excluded.enabled,
			updated_at   = excluded.updated_at
		RETURNING id`),
		ov.ID,
		ov.TenantID,
		ov.Key,
		ov.ThemePreset,
		tokensJSON,
		ov.RuleID,
		ov.Enabled,
		s.formatTime(ov.UpdatedAt),
	).Scan(&ov.ID)
	if err != nil {
		return domain.TokenOverride{}, fmt.Errorf("persistence: upsert token override %q: %w", ov.Key, err)
	}
	if ov.Tokens == nil {
		ov.Tokens = map[string]string{}
	}
	ov.UpdatedAt = ov.UpdatedAt.UTC()
	return ov, nil
}

func scanTokenOverride(sc rowScanner) (domain.TokenOverride, error) {
	var (
		ov         domain.TokenOverride
		tokensJSON string
		rawUpdated any
	)
	if err := sc.Scan(
		&ov.ID,
		&ov.TenantID,
		&ov.Key,
		&ov.ThemePreset,
		&tokensJSON,
		&ov.RuleID,
		&ov.Enabled,
		&rawUpdated,
	); err != nil {
		return domain.TokenOverride{}, err
	}
	if err := unmarshalJSON(tokensJSON, &ov.Tokens); err != nil {
		return domain.TokenOverride{}, fmt.Errorf("persistence: decode token overrides: %w", err)
	}
	if ov.Tokens == nil {
		ov.Tokens = map[string]string{}
	}
	var err error
	if ov.UpdatedAt, err = parseTime(rawUpdated); err != nil {
		return domain.TokenOverride{}, err
	}
	return ov, nil
}

// ── ChangeLogStore ─────────────────────────────────────────────────────────────

// AppendChangeLog persists an immutable audit changelog entry.
func (s *SQLStore) AppendChangeLog(ctx context.Context, entry domain.ChangeLogEntry) error {
	if strings.TrimSpace(entry.ID) == "" {
		entry.ID = "chg_" + uuid.NewString()[:8]
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	diffJSON, err := marshalJSON(entry.Diff, "[]")
	if err != nil {
		return fmt.Errorf("persistence: marshal changelog diff: %w", err)
	}

	_, err = s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO change_logs (
			id, tenant_id, entity_type, entity_id, action, actor, summary, diff, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		entry.ID,
		entry.TenantID,
		entry.EntityType,
		entry.EntityID,
		entry.Action,
		entry.Actor,
		entry.Summary,
		diffJSON,
		s.formatTime(entry.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("persistence: append changelog: %w", err)
	}
	return nil
}

// ListChangeLogs returns the most recent audit changelog entries for tenantID, newest first.
func (s *SQLStore) ListChangeLogs(ctx context.Context, tenantID string, limit int) ([]domain.ChangeLogEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, tenant_id, entity_type, entity_id, action, actor, summary, diff, created_at
		FROM change_logs
		WHERE tenant_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?`), tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("persistence: list changelogs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.ChangeLogEntry, 0)
	for rows.Next() {
		var (
			e          domain.ChangeLogEntry
			diffJSON   string
			rawCreated any
		)
		if err := rows.Scan(
			&e.ID,
			&e.TenantID,
			&e.EntityType,
			&e.EntityID,
			&e.Action,
			&e.Actor,
			&e.Summary,
			&diffJSON,
			&rawCreated,
		); err != nil {
			return nil, err
		}
		if err := unmarshalJSON(diffJSON, &e.Diff); err != nil {
			return nil, fmt.Errorf("persistence: decode changelog diff: %w", err)
		}
		if len(e.Diff) == 0 {
			e.Diff = nil
		}
		if e.CreatedAt, err = parseTime(rawCreated); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("persistence: iterate changelogs: %w", err)
	}
	return out, nil
}
