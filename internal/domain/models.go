// Package domain defines pure business entities, value objects, and storage/engine ports
// for the Abeta Dynamic Layout Engine. It has zero HTTP, Gin, or database driver imports.
package domain

import (
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
)

// LayoutStatus represents the authoring lifecycle state of a dynamic layout.
type LayoutStatus string

const (
	LayoutStatusDraft     LayoutStatus = "draft"
	LayoutStatusPublished LayoutStatus = "published"
	LayoutStatusArchived  LayoutStatus = "archived"
)

// SlottingAction describes how a dynamic slotting rule modifies a layout's slot list.
type SlottingAction string

const (
	SlottingActionInsert  SlottingAction = "insert"
	SlottingActionPrepend SlottingAction = "prepend"
	SlottingActionAppend  SlottingAction = "append"
	SlottingActionReplace SlottingAction = "replace"
)

// Tenant isolates layouts, widgets, templates, rules, tokens, and snapshots per app or brand.
type Tenant struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WidgetTemplate is a reusable blueprint mapping to one of the 21 KMP v1 SnippetTypes.
type WidgetTemplate struct {
	ID           string         `json:"id"`
	TenantID     string         `json:"tenant_id"`
	Key          string         `json:"key"`
	Name         string         `json:"name"`
	SnippetType  string         `json:"snippet_type"`
	Description  string         `json:"description,omitempty"`
	DefaultProps map[string]any `json:"default_props"`
	Schema       map[string]any `json:"schema,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Widget is a configured, reusable content/commerce/container component stored in the CMS.
type Widget struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	Slug        string         `json:"slug"`
	Name        string         `json:"name"`
	TemplateKey string         `json:"template_key,omitempty"`
	SnippetType string         `json:"snippet_type"`
	Props       map[string]any `json:"props"`
	Tracking    map[string]any `json:"tracking,omitempty"`
	Motion      map[string]any `json:"motion,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Active      bool           `json:"active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// RuleCondition evaluates a single predicate against the incoming RequestContext.
type RuleCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// ScheduleWindow bounds when a rule, slot, splash, or slotting override is active.
type ScheduleWindow struct {
	StartAt      *time.Time `json:"start_at,omitempty"`
	EndAt        *time.Time `json:"end_at,omitempty"`
	DaysOfWeek   []int      `json:"days_of_week,omitempty"`
	StartHourUTC *int       `json:"start_hour_utc,omitempty"`
	EndHourUTC   *int       `json:"end_hour_utc,omitempty"`
}

// TargetingRule combines conditions, rollout percentages, and an optional schedule window.
type TargetingRule struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	Key         string          `json:"key"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	MatchMode   string          `json:"match_mode"` // "all" (AND) or "any" (OR)
	Conditions  []RuleCondition `json:"conditions"`
	Schedule    *ScheduleWindow `json:"schedule,omitempty"`
	Priority    int             `json:"priority"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// LayoutSlot binds a Widget into an ordered position inside a Layout with optional targeting.
type LayoutSlot struct {
	SlotID    string          `json:"slot_id"`
	WidgetID  string          `json:"widget_id"`
	Position  int             `json:"position"`
	RuleID    string          `json:"rule_id,omitempty"`
	Schedule  *ScheduleWindow `json:"schedule,omitempty"`
	Overrides map[string]any  `json:"overrides,omitempty"`
	Pinned    bool            `json:"pinned,omitempty"`
	Sticky    bool            `json:"sticky,omitempty"`
	Enabled   bool            `json:"enabled"`
}

// LayoutPageConfig holds page-level SDUI chrome (toolbar, floating cart, adaptive split-pane, etc.).
type LayoutPageConfig struct {
	Toolbar        *snippets.ToolbarSpec      `json:"toolbar,omitempty"`
	BottomBar      *snippets.BottomBarSpec    `json:"bottom_bar,omitempty"`
	FloatingCart   *snippets.FloatingCartSpec `json:"floating_cart,omitempty"`
	AdaptiveLayout *snippets.AdaptivePageSpec `json:"adaptive_layout,omitempty"`
	Pagination     *snippets.PaginationSpec   `json:"pagination,omitempty"`
	PullToRefresh  *actions.ActionTrigger     `json:"pull_to_refresh,omitempty"`
	PageActions    []actions.ActionTrigger    `json:"page_actions,omitempty"`
}

// Layout is the editable draft/published page definition for a given tenant and route slug.
type Layout struct {
	ID               string           `json:"id"`
	TenantID         string           `json:"tenant_id"`
	Slug             string           `json:"slug"`
	Title            string           `json:"title"`
	Description      string           `json:"description,omitempty"`
	Status           LayoutStatus     `json:"status"`
	PublishedVersion int              `json:"published_version"`
	DraftVersion     int              `json:"draft_version"`
	PageConfig       LayoutPageConfig `json:"page_config"`
	Slots            []LayoutSlot     `json:"slots"`
	Tags             []string         `json:"tags,omitempty"`
	DefaultRuleID    string           `json:"default_rule_id,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// SlottingRule dynamically injects, prepends, appends, or replaces widgets on a layout at runtime.
type SlottingRule struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	LayoutSlug     string          `json:"layout_slug"`
	Name           string          `json:"name"`
	Action         SlottingAction  `json:"action"`
	TargetPosition int             `json:"target_position"`
	WidgetID       string          `json:"widget_id"`
	RuleID         string          `json:"rule_id,omitempty"`
	Schedule       *ScheduleWindow `json:"schedule,omitempty"`
	Priority       int             `json:"priority"`
	Enabled        bool            `json:"enabled"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// PublishedSnapshot is an immutable versioned snapshot published from a validated Layout draft.
type PublishedSnapshot struct {
	ID              string            `json:"id"`
	TenantID        string            `json:"tenant_id"`
	LayoutID        string            `json:"layout_id"`
	LayoutSlug      string            `json:"layout_slug"`
	Version         int               `json:"version"`
	Title           string            `json:"title"`
	PageConfig      LayoutPageConfig  `json:"page_config"`
	Slots           []LayoutSlot      `json:"slots"`
	ResolvedWidgets map[string]Widget `json:"resolved_widgets"`
	PublishedBy     string            `json:"published_by"`
	Comment         string            `json:"comment,omitempty"`
	PublishedAt     time.Time         `json:"published_at"`
}

// FallbackLayout holds a pre-validated PageResponse served automatically if runtime resolution fails.
type FallbackLayout struct {
	ID              string                `json:"id"`
	TenantID        string                `json:"tenant_id"`
	LayoutSlug      string                `json:"layout_slug"`
	SnapshotVersion int                   `json:"snapshot_version"`
	Page            snippets.PageResponse `json:"page"`
	UpdatedBy       string                `json:"updated_by"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// SplashConfig controls rule-targeted app splash/launch branding dynamically.
type SplashConfig struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenant_id"`
	Key                string          `json:"key"`
	ImageURL           string          `json:"image_url"`
	BackgroundColorHex string          `json:"background_color_hex"`
	DurationMs         int             `json:"duration_ms"`
	LottieURL          string          `json:"lottie_url,omitempty"`
	RuleID             string          `json:"rule_id,omitempty"`
	Schedule           *ScheduleWindow `json:"schedule,omitempty"`
	IsDefault          bool            `json:"is_default"`
	Enabled            bool            `json:"enabled"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// TokenOverride publishes rule-targeted runtime color/theme token overrides.
type TokenOverride struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	Key         string            `json:"key"`
	ThemePreset string            `json:"theme_preset"`
	Tokens      map[string]string `json:"tokens"`
	RuleID      string            `json:"rule_id,omitempty"`
	Enabled     bool              `json:"enabled"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// DiffItem records a single field-level JSON diff operation between two revisions.
type DiffItem struct {
	Path   string `json:"path"`
	Op     string `json:"op"` // "add", "remove", "replace"
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

// ChangeLogEntry provides an immutable audit trail for Studio mutations, publishes, and rollbacks.
type ChangeLogEntry struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	Action     string     `json:"action"`
	Actor      string     `json:"actor"`
	Summary    string     `json:"summary"`
	Diff       []DiffItem `json:"diff,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// RequestContext captures caller targeting attributes (platform, version, city, segment, time).
type RequestContext struct {
	TenantSlug  string            `json:"tenant_slug"`
	Platform    string            `json:"platform"`     // android, ios, macos, watchos, wearos, web, desktop
	AppVersion  string            `json:"app_version"`  // e.g. 1.4.2
	City        string            `json:"city"`         // e.g. bengaluru, mumbai
	Zone        string            `json:"zone"`         // e.g. koramangala-01
	Segment     string            `json:"segment"`      // e.g. vip, new_user, returning
	UserID      string            `json:"user_id"`      // stable hash key for percentage rollouts
	WindowClass string            `json:"window_class"` // compact, medium, expanded
	DarkMode    bool              `json:"dark_mode"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Now         time.Time         `json:"now"`
}

// ResolutionTrace explains exactly how a dynamic page was composed for debugging & Studio preview.
type ResolutionTrace struct {
	TenantSlug         string    `json:"tenant_slug"`
	LayoutSlug         string    `json:"layout_slug"`
	SnapshotVersion    int       `json:"snapshot_version"`
	PreviewDraft       bool      `json:"preview_draft,omitempty"`
	UsedFallback       bool      `json:"used_fallback,omitempty"`
	FallbackReason     string    `json:"fallback_reason,omitempty"`
	MatchedRuleIDs     []string  `json:"matched_rule_ids,omitempty"`
	AppliedSlottingIDs []string  `json:"applied_slotting_ids,omitempty"`
	SkippedSlotIDs     []string  `json:"skipped_slot_ids,omitempty"`
	DurationMs         int64     `json:"duration_ms"`
	ResolvedAt         time.Time `json:"resolved_at"`
}

// ResolvedPageResult bundles the v1 KMP PageResponse with its ResolutionTrace.
type ResolvedPageResult struct {
	Page  snippets.PageResponse `json:"page"`
	Trace ResolutionTrace       `json:"trace"`
}

// ValidationIssue reports a single schema or semantic problem before a draft can be published.
type ValidationIssue struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
