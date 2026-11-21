package domain

import (
	"context"
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
)

// TenantStore persists multi-tenant workspaces.
type TenantStore interface {
	ListTenants(ctx context.Context) ([]Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (Tenant, error)
	GetTenantByID(ctx context.Context, id string) (Tenant, error)
	UpsertTenant(ctx context.Context, tenant Tenant) (Tenant, error)
}

// TemplateStore persists reusable widget templates.
type TemplateStore interface {
	ListTemplates(ctx context.Context, tenantID string) ([]WidgetTemplate, error)
	GetTemplateByKey(ctx context.Context, tenantID, key string) (WidgetTemplate, error)
	UpsertTemplate(ctx context.Context, tmpl WidgetTemplate) (WidgetTemplate, error)
}

// WidgetStore persists configured widgets.
type WidgetStore interface {
	ListWidgets(ctx context.Context, tenantID string) ([]Widget, error)
	GetWidgetByID(ctx context.Context, tenantID, id string) (Widget, error)
	GetWidgetsByIDs(ctx context.Context, tenantID string, ids []string) (map[string]Widget, error)
	UpsertWidget(ctx context.Context, widget Widget) (Widget, error)
	DeleteWidget(ctx context.Context, tenantID, id string) error
}

// LayoutStore persists editable layout drafts and slot orderings.
type LayoutStore interface {
	ListLayouts(ctx context.Context, tenantID string) ([]Layout, error)
	GetLayoutBySlug(ctx context.Context, tenantID, slug string) (Layout, error)
	UpsertLayout(ctx context.Context, layout Layout) (Layout, error)
	DeleteLayout(ctx context.Context, tenantID, slug string) error
}

// RuleStore persists targeting and scheduling rules.
type RuleStore interface {
	ListRules(ctx context.Context, tenantID string) ([]TargetingRule, error)
	GetRuleByID(ctx context.Context, tenantID, id string) (TargetingRule, error)
	UpsertRule(ctx context.Context, rule TargetingRule) (TargetingRule, error)
	DeleteRule(ctx context.Context, tenantID, id string) error
}

// SlottingStore persists dynamic slot-injection rules.
type SlottingStore interface {
	ListSlottingRules(ctx context.Context, tenantID, layoutSlug string) ([]SlottingRule, error)
	UpsertSlottingRule(ctx context.Context, rule SlottingRule) (SlottingRule, error)
	DeleteSlottingRule(ctx context.Context, tenantID, id string) error
}

// SnapshotStore persists immutable published layout versions.
type SnapshotStore interface {
	SaveSnapshot(ctx context.Context, snapshot PublishedSnapshot) error
	GetLatestSnapshot(ctx context.Context, tenantID, layoutSlug string) (PublishedSnapshot, error)
	GetSnapshotByVersion(ctx context.Context, tenantID, layoutSlug string, version int) (PublishedSnapshot, error)
	ListSnapshots(ctx context.Context, tenantID, layoutSlug string) ([]PublishedSnapshot, error)
}

// FallbackStore persists pre-validated fallback pages per tenant and layout slug.
type FallbackStore interface {
	SaveFallback(ctx context.Context, fb FallbackLayout) error
	GetFallback(ctx context.Context, tenantID, layoutSlug string) (FallbackLayout, error)
}

// ConfigStore persists rule-targeted splash screen and token override configs.
type ConfigStore interface {
	ListSplashConfigs(ctx context.Context, tenantID string) ([]SplashConfig, error)
	UpsertSplashConfig(ctx context.Context, cfg SplashConfig) (SplashConfig, error)
	ListTokenOverrides(ctx context.Context, tenantID string) ([]TokenOverride, error)
	UpsertTokenOverride(ctx context.Context, ov TokenOverride) (TokenOverride, error)
}

// ChangeLogStore persists immutable audit logs with JSON diffs.
type ChangeLogStore interface {
	AppendChangeLog(ctx context.Context, entry ChangeLogEntry) error
	ListChangeLogs(ctx context.Context, tenantID string, limit int) ([]ChangeLogEntry, error)
}

// Store aggregates all Clean Architecture persistence ports behind one closeable adapter.
type Store interface {
	TenantStore
	TemplateStore
	WidgetStore
	LayoutStore
	RuleStore
	SlottingStore
	SnapshotStore
	FallbackStore
	ConfigStore
	ChangeLogStore
	Ping(ctx context.Context) error
	Close() error
}

// RuleEvaluator evaluates targeting rules and schedule windows deterministically.
type RuleEvaluator interface {
	MatchesRule(rule TargetingRule, ctx RequestContext) bool
	MatchesSchedule(window *ScheduleWindow, now time.Time) bool
}

// SlottingEngine merges dynamic slotting rules onto a layout's base slot list.
type SlottingEngine interface {
	Apply(
		base []LayoutSlot,
		rules []SlottingRule,
		ruleLookup map[string]TargetingRule,
		ctx RequestContext,
	) (merged []LayoutSlot, appliedRuleIDs []string)
}

// PageComposer transforms resolved slots and widgets into a v1 KMP PageResponse.
type PageComposer interface {
	Compose(
		layoutSlug string,
		title string,
		pageConfig LayoutPageConfig,
		slots []LayoutSlot,
		widgets map[string]Widget,
		ctx RequestContext,
	) (snippets.PageResponse, error)
}

// LayoutValidator validates a Layout draft and its widget bindings before publication.
type LayoutValidator interface {
	ValidateLayout(layout Layout, widgets map[string]Widget) []ValidationIssue
}
