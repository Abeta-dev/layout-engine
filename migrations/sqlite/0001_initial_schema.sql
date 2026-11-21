-- Abeta Dynamic Layout Engine — Initial SQLite Schema (v0001)
-- Uses TEXT columns for JSON payloads and fixed-width UTC RFC3339Nano timestamps.

CREATE TABLE IF NOT EXISTS tenants (
    id          TEXT PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    active      INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS widget_templates (
    id            TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    template_key  TEXT NOT NULL,
    name          TEXT NOT NULL,
    snippet_type  TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    default_props TEXT NOT NULL DEFAULT '{}',
    schema_json   TEXT NOT NULL DEFAULT '{}',
    tags          TEXT NOT NULL DEFAULT '[]',
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL,
    UNIQUE (tenant_id, template_key)
);

CREATE INDEX IF NOT EXISTS idx_widget_templates_tenant_type
    ON widget_templates (tenant_id, snippet_type);

CREATE TABLE IF NOT EXISTS widgets (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    slug         TEXT NOT NULL,
    name         TEXT NOT NULL,
    template_key TEXT NOT NULL DEFAULT '',
    snippet_type TEXT NOT NULL,
    props        TEXT NOT NULL DEFAULT '{}',
    tracking     TEXT,
    motion       TEXT,
    tags         TEXT NOT NULL DEFAULT '[]',
    active       INTEGER NOT NULL DEFAULT 1,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL,
    UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_widgets_tenant_type
    ON widgets (tenant_id, snippet_type);

CREATE INDEX IF NOT EXISTS idx_widgets_tenant_active
    ON widgets (tenant_id, active);

CREATE TABLE IF NOT EXISTS targeting_rules (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rule_key    TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    match_mode  TEXT NOT NULL DEFAULT 'all',
    conditions  TEXT NOT NULL DEFAULT '[]',
    schedule    TEXT,
    priority    INTEGER NOT NULL DEFAULT 0,
    enabled     INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE (tenant_id, rule_key)
);

CREATE INDEX IF NOT EXISTS idx_targeting_rules_tenant_priority
    ON targeting_rules (tenant_id, priority DESC);

CREATE TABLE IF NOT EXISTS layouts (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    slug              TEXT NOT NULL,
    title             TEXT NOT NULL,
    description       TEXT NOT NULL DEFAULT '',
    status            TEXT NOT NULL DEFAULT 'draft',
    published_version INTEGER NOT NULL DEFAULT 0,
    draft_version     INTEGER NOT NULL DEFAULT 1,
    page_config       TEXT NOT NULL DEFAULT '{}',
    slots             TEXT NOT NULL DEFAULT '[]',
    tags              TEXT NOT NULL DEFAULT '[]',
    default_rule_id   TEXT NOT NULL DEFAULT '',
    created_at        TEXT NOT NULL,
    updated_at        TEXT NOT NULL,
    UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_layouts_tenant_status
    ON layouts (tenant_id, status);

CREATE TABLE IF NOT EXISTS slotting_rules (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    layout_slug     TEXT NOT NULL,
    name            TEXT NOT NULL,
    action          TEXT NOT NULL,
    target_position INTEGER NOT NULL DEFAULT 0,
    widget_id       TEXT NOT NULL,
    rule_id         TEXT NOT NULL DEFAULT '',
    schedule        TEXT,
    priority        INTEGER NOT NULL DEFAULT 0,
    enabled         INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_slotting_rules_tenant_layout
    ON slotting_rules (tenant_id, layout_slug, priority DESC);

CREATE TABLE IF NOT EXISTS published_snapshots (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    layout_id        TEXT NOT NULL,
    layout_slug      TEXT NOT NULL,
    version          INTEGER NOT NULL,
    title            TEXT NOT NULL,
    page_config      TEXT NOT NULL DEFAULT '{}',
    slots            TEXT NOT NULL DEFAULT '[]',
    resolved_widgets TEXT NOT NULL DEFAULT '{}',
    published_by     TEXT NOT NULL DEFAULT '',
    comment          TEXT NOT NULL DEFAULT '',
    published_at     TEXT NOT NULL,
    UNIQUE (tenant_id, layout_slug, version)
);

CREATE INDEX IF NOT EXISTS idx_published_snapshots_lookup
    ON published_snapshots (tenant_id, layout_slug, version DESC);

CREATE TABLE IF NOT EXISTS fallback_layouts (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    layout_slug      TEXT NOT NULL,
    snapshot_version INTEGER NOT NULL DEFAULT 0,
    page_json        TEXT NOT NULL,
    updated_by       TEXT NOT NULL DEFAULT '',
    updated_at       TEXT NOT NULL,
    UNIQUE (tenant_id, layout_slug)
);

CREATE TABLE IF NOT EXISTS splash_configs (
    id                   TEXT PRIMARY KEY,
    tenant_id            TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    config_key           TEXT NOT NULL,
    image_url            TEXT NOT NULL DEFAULT '',
    background_color_hex TEXT NOT NULL DEFAULT '#FFFFFF',
    duration_ms          INTEGER NOT NULL DEFAULT 1500,
    lottie_url           TEXT NOT NULL DEFAULT '',
    rule_id              TEXT NOT NULL DEFAULT '',
    schedule             TEXT,
    is_default           INTEGER NOT NULL DEFAULT 0,
    enabled              INTEGER NOT NULL DEFAULT 1,
    updated_at           TEXT NOT NULL,
    UNIQUE (tenant_id, config_key)
);

CREATE INDEX IF NOT EXISTS idx_splash_configs_tenant
    ON splash_configs (tenant_id, enabled);

CREATE TABLE IF NOT EXISTS token_overrides (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    config_key   TEXT NOT NULL,
    theme_preset TEXT NOT NULL DEFAULT 'default',
    tokens       TEXT NOT NULL DEFAULT '{}',
    rule_id      TEXT NOT NULL DEFAULT '',
    enabled      INTEGER NOT NULL DEFAULT 1,
    updated_at   TEXT NOT NULL,
    UNIQUE (tenant_id, config_key)
);

CREATE INDEX IF NOT EXISTS idx_token_overrides_tenant
    ON token_overrides (tenant_id, enabled);

CREATE TABLE IF NOT EXISTS change_logs (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id   TEXT NOT NULL,
    action      TEXT NOT NULL,
    actor       TEXT NOT NULL DEFAULT '',
    summary     TEXT NOT NULL DEFAULT '',
    diff        TEXT NOT NULL DEFAULT '[]',
    created_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_change_logs_tenant_created
    ON change_logs (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_change_logs_entity
    ON change_logs (tenant_id, entity_type, entity_id);
