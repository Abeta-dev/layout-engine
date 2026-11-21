package engine_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/engine"
)

func intPtr(v int) *int { return &v }

func TestRuleEvaluator_MatchesSchedule(t *testing.T) {
	t.Parallel()

	eval := engine.NewRuleEvaluator()
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name   string
		window *domain.ScheduleWindow
		now    time.Time
		want   bool
	}{
		{
			name:   "nil window always matches",
			window: nil,
			now:    time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "within start and end dates",
			window: &domain.ScheduleWindow{StartAt: &start, EndAt: &end},
			now:    time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "before start date",
			window: &domain.ScheduleWindow{StartAt: &start, EndAt: &end},
			now:    time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC),
			want:   false,
		},
		{
			name:   "after end date",
			window: &domain.ScheduleWindow{StartAt: &start, EndAt: &end},
			now:    time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name: "matching weekday (Monday=1)",
			// 2026-06-15 is Monday
			window: &domain.ScheduleWindow{DaysOfWeek: []int{1, 3, 5}},
			now:    time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "non-matching weekday (Tuesday=2)",
			window: &domain.ScheduleWindow{DaysOfWeek: []int{0, 6}},
			now:    time.Date(2026, 6, 16, 9, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "normal hour window 6..11 at start hour 6",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(6), EndHourUTC: intPtr(11)},
			now:    time.Date(2026, 6, 15, 6, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "normal hour window 6..11 at hour 10",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(6), EndHourUTC: intPtr(11)},
			now:    time.Date(2026, 6, 15, 10, 45, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "normal hour window 6..11 excludes hour 11",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(6), EndHourUTC: intPtr(11)},
			now:    time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "normal hour window 6..11 excludes hour 5",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(6), EndHourUTC: intPtr(11)},
			now:    time.Date(2026, 6, 15, 5, 59, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "overnight wrap window 22..4 matches hour 22",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(22), EndHourUTC: intPtr(4)},
			now:    time.Date(2026, 6, 15, 22, 30, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "overnight wrap window 22..4 matches hour 0",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(22), EndHourUTC: intPtr(4)},
			now:    time.Date(2026, 6, 16, 0, 15, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "overnight wrap window 22..4 matches hour 3",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(22), EndHourUTC: intPtr(4)},
			now:    time.Date(2026, 6, 16, 3, 59, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "overnight wrap window 22..4 excludes hour 4",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(22), EndHourUTC: intPtr(4)},
			now:    time.Date(2026, 6, 16, 4, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "overnight wrap window 22..4 excludes hour 21",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(22), EndHourUTC: intPtr(4)},
			now:    time.Date(2026, 6, 15, 21, 59, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "single hour window 12..12 matches hour 12",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(12), EndHourUTC: intPtr(12)},
			now:    time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "single hour window 12..12 excludes hour 13",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(12), EndHourUTC: intPtr(12)},
			now:    time.Date(2026, 6, 15, 13, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "start hour only",
			window: &domain.ScheduleWindow{StartHourUTC: intPtr(18)},
			now:    time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "end hour only",
			window: &domain.ScheduleWindow{EndHourUTC: intPtr(10)},
			now:    time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := eval.MatchesSchedule(tc.window, tc.now)
			if got != tc.want {
				t.Fatalf("MatchesSchedule() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRuleEvaluator_OperatorsAndModes(t *testing.T) {
	t.Parallel()

	eval := engine.NewRuleEvaluator()
	fixedNow := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)

	baseCtx := domain.RequestContext{
		TenantSlug:  "abeta-fresh",
		Platform:    "android",
		AppVersion:  "1.4.2",
		City:        "Bengaluru",
		Zone:        "koramangala-01",
		Segment:     "vip",
		UserID:      "usr_4242",
		WindowClass: "expanded",
		DarkMode:    true,
		Attributes: map[string]string{
			"order_count": "12",
			"tier_score":  "85.5",
			"promo_tag":   "summer-express-2026",
		},
		Now: fixedNow,
	}

	tests := []struct {
		name string
		rule domain.TargetingRule
		ctx  domain.RequestContext
		want bool
	}{
		{
			name: "disabled rule returns false",
			rule: domain.TargetingRule{Enabled: false},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "enabled rule with empty conditions and valid schedule returns true",
			rule: domain.TargetingRule{
				Enabled:  true,
				Schedule: &domain.ScheduleWindow{StartHourUTC: intPtr(6), EndHourUTC: intPtr(12)},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "eq case-insensitive string on city",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "city", Operator: "eq", Value: "bengaluru"}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "eq boolean on dark_mode with bool value",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "dark_mode", Operator: "eq", Value: true}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "eq boolean on dark_mode with string value",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "dark_mode", Operator: "eq", Value: "false"}},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "neq on platform",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "platform", Operator: "neq", Value: "ios"}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "in with []any list",
			rule: domain.TargetingRule{
				Enabled: true,
				Conditions: []domain.RuleCondition{
					{Field: "city", Operator: "in", Value: []any{"mumbai", "bengaluru", "delhi"}},
				},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "in with []string list",
			rule: domain.TargetingRule{
				Enabled: true,
				Conditions: []domain.RuleCondition{
					{Field: "window_class", Operator: "in", Value: []string{"medium", "expanded"}},
				},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "not_in matching",
			rule: domain.TargetingRule{
				Enabled: true,
				Conditions: []domain.RuleCondition{
					{Field: "platform", Operator: "not_in", Value: []string{"watchos", "wearos"}},
				},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "not_in failing",
			rule: domain.TargetingRule{
				Enabled: true,
				Conditions: []domain.RuleCondition{
					{Field: "segment", Operator: "not_in", Value: []string{"vip", "staff"}},
				},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "gt numeric on attribute",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "order_count", Operator: "gt", Value: 10}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "gte numeric float on attribute",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "tier_score", Operator: "gte", Value: 85.5}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "lt numeric on attribute",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "order_count", Operator: "lt", Value: 12}},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "lte numeric on attribute",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "order_count", Operator: "lte", Value: 12}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "semver_gte equal version with v prefix",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_gte", Value: "v1.4.2"}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "semver_gte higher required minor fails",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_gte", Value: "1.5.0"}},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "semver_gte release beats prerelease",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_gte", Value: "1.4.2-rc.1"}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "semver_lt passes when app version is lower",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_lt", Value: "2.0.0"}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "semver_lt fails on invalid version",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_lt", Value: "not-a-semver"}},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "contains case-insensitive on zone",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "zone", Operator: "contains", Value: "KORAMANGALA"}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "regex matches attribute",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "promo_tag", Operator: "regex", Value: `^summer-.*-2026$`}},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "invalid regex returns false safely",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "zone", Operator: "regex", Value: `[invalid(`}},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "match_mode any (OR) passes when one matches",
			rule: domain.TargetingRule{
				Enabled:   true,
				MatchMode: "any",
				Conditions: []domain.RuleCondition{
					{Field: "city", Operator: "eq", Value: "mumbai"},
					{Field: "tenant_slug", Operator: "eq", Value: "abeta-fresh"},
				},
			},
			ctx:  baseCtx,
			want: true,
		},
		{
			name: "match_mode all (AND) fails when one fails",
			rule: domain.TargetingRule{
				Enabled:   true,
				MatchMode: "all",
				Conditions: []domain.RuleCondition{
					{Field: "city", Operator: "eq", Value: "bengaluru"},
					{Field: "segment", Operator: "eq", Value: "new_user"},
				},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "unknown operator returns false",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "city", Operator: "approx", Value: "bengaluru"}},
			},
			ctx:  baseCtx,
			want: false,
		},
		{
			name: "missing attribute fails eq and passes neq",
			rule: domain.TargetingRule{
				Enabled: true,
				Conditions: []domain.RuleCondition{
					{Field: "unknown_attr", Operator: "neq", Value: "vip"},
				},
			},
			ctx:  baseCtx,
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := eval.MatchesRule(tc.rule, tc.ctx); got != tc.want {
				t.Fatalf("MatchesRule() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRuleEvaluator_BucketPctDeterminismAndDistribution(t *testing.T) {
	t.Parallel()

	eval := engine.NewRuleEvaluator()
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)

	rule0 := domain.TargetingRule{
		Key:        "rollout_zero",
		Enabled:    true,
		Conditions: []domain.RuleCondition{{Field: "user_id", Operator: "bucket_pct", Value: 0}},
	}
	rule100 := domain.TargetingRule{
		Key:        "rollout_full",
		Enabled:    true,
		Conditions: []domain.RuleCondition{{Field: "user_id", Operator: "bucket_pct", Value: 100}},
	}
	rule30 := domain.TargetingRule{
		Key:        "express_banner_exp",
		Enabled:    true,
		Conditions: []domain.RuleCondition{{Field: "user_id", Operator: "bucket_pct", Value: 30}},
	}

	matched := 0
	const total = 1000
	for i := 0; i < total; i++ {
		ctx := domain.RequestContext{
			UserID: fmt.Sprintf("user_%04d", i),
			Now:    now,
		}
		if eval.MatchesRule(rule0, ctx) {
			t.Fatalf("0%% bucket matched user %s", ctx.UserID)
		}
		if !eval.MatchesRule(rule100, ctx) {
			t.Fatalf("100%% bucket did not match user %s", ctx.UserID)
		}
		first := eval.MatchesRule(rule30, ctx)
		second := eval.MatchesRule(rule30, ctx)
		if first != second {
			t.Fatalf("bucket_pct non-deterministic for %s", ctx.UserID)
		}
		if first {
			matched++
		}
	}

	if matched < 220 || matched > 380 {
		t.Fatalf("30%% bucket distribution out of expected range: got %d/%d", matched, total)
	}
}

func TestSlottingEngine_Apply(t *testing.T) {
	t.Parallel()

	eval := engine.NewRuleEvaluator()
	slotting := engine.NewSlottingEngine(eval)
	now := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)

	ctx := domain.RequestContext{
		City:    "bengaluru",
		Segment: "vip",
		Now:     now,
	}

	ruleLookup := map[string]domain.TargetingRule{
		"rule-vip": {
			ID:      "rule-vip",
			Key:     "vip_users",
			Enabled: true,
			Conditions: []domain.RuleCondition{
				{Field: "segment", Operator: "eq", Value: "vip"},
			},
		},
		"rule-mumbai": {
			ID:      "rule-mumbai",
			Key:     "mumbai_only",
			Enabled: true,
			Conditions: []domain.RuleCondition{
				{Field: "city", Operator: "eq", Value: "mumbai"},
			},
		},
	}

	base := []domain.LayoutSlot{
		{SlotID: "slot_c", WidgetID: "w_c", Position: 20, Enabled: true},
		{SlotID: "slot_a", WidgetID: "w_a", Position: 0, Pinned: true, Enabled: true},
		{SlotID: "slot_b", WidgetID: "w_b", Position: 10, Enabled: true},
	}

	rules := []domain.SlottingRule{
		{
			ID:       "sr_disabled",
			Action:   domain.SlottingActionPrepend,
			WidgetID: "w_disabled",
			Priority: 999,
			Enabled:  false,
		},
		{
			ID:       "sr_mumbai_skip",
			Action:   domain.SlottingActionPrepend,
			WidgetID: "w_mumbai",
			RuleID:   "rule-mumbai",
			Priority: 500,
			Enabled:  true,
		},
		{
			ID:             "sr_replace_pinned",
			Action:         domain.SlottingActionReplace,
			TargetPosition: 0, // slot 0 is w_a (Pinned: true) -> must insert instead of overwrite!
			WidgetID:       "w_flash",
			RuleID:         "rule-vip",
			Priority:       100,
			Enabled:        true,
		},
		{
			ID:             "sr_replace_unpinned",
			Action:         domain.SlottingActionReplace,
			TargetPosition: 2, // now points to w_b (unpinned) -> replaces w_b
			WidgetID:       "w_vip_banner",
			Priority:       90,
			Enabled:        true,
		},
		{
			ID:       "sr_prepend_promote",
			Action:   domain.SlottingActionPrepend,
			WidgetID: "w_c", // promotes w_c from base to top, deduplicating base w_c
			Priority: 80,
			Enabled:  true,
		},
		{
			ID:       "sr_low_prio_dup",
			Action:   domain.SlottingActionAppend,
			WidgetID: "w_flash", // already slotted by sr_replace_pinned -> skipped
			Priority: 10,
			Enabled:  true,
		},
		{
			ID:       "sr_append",
			Action:   domain.SlottingActionAppend,
			WidgetID: "w_footer",
			Priority: 5,
			Enabled:  true,
		},
	}

	merged, applied := slotting.Apply(base, rules, ruleLookup, ctx)

	wantWidgets := []string{"w_c", "w_a", "w_flash", "w_vip_banner", "w_footer"}
	gotWidgets := make([]string, len(merged))
	for i, s := range merged {
		gotWidgets[i] = s.WidgetID
		if s.Position != i {
			t.Errorf("merged[%d].Position = %d, want sequential %d", i, s.Position, i)
		}
	}
	if !slices.Equal(gotWidgets, wantWidgets) {
		t.Fatalf("merged widgets = %v, want %v", gotWidgets, wantWidgets)
	}

	wantApplied := []string{"sr_replace_pinned", "sr_replace_unpinned", "sr_prepend_promote", "sr_append"}
	if !slices.Equal(applied, wantApplied) {
		t.Fatalf("applied rule IDs = %v, want %v", applied, wantApplied)
	}

	// Verify base slice was not mutated.
	if base[0].SlotID != "slot_c" || base[0].Position != 20 {
		t.Fatalf("base slice was unexpectedly mutated: %+v", base[0])
	}
}

func TestSlottingEngine_InsertClampingAndBaseDedup(t *testing.T) {
	t.Parallel()

	slotting := engine.NewSlottingEngine(nil)
	base := []domain.LayoutSlot{
		{SlotID: "s1", WidgetID: "w1", Position: 0, Enabled: true},
		{SlotID: "s2", WidgetID: "w1", Position: 1, Enabled: true}, // duplicate in base
	}
	rules := []domain.SlottingRule{
		{ID: "r_neg", Action: domain.SlottingActionInsert, TargetPosition: -5, WidgetID: "w0", Priority: 20, Enabled: true},
		{ID: "r_high", Action: domain.SlottingActionInsert, TargetPosition: 99, WidgetID: "w9", Priority: 10, Enabled: true},
	}

	merged, applied := slotting.Apply(base, rules, nil, domain.RequestContext{Now: time.Now().UTC()})
	gotWidgets := make([]string, len(merged))
	for i, s := range merged {
		gotWidgets[i] = s.WidgetID
	}
	wantWidgets := []string{"w0", "w1", "w9"}
	if !slices.Equal(gotWidgets, wantWidgets) {
		t.Fatalf("merged widgets = %v, want %v", gotWidgets, wantWidgets)
	}
	if !slices.Equal(applied, []string{"r_neg", "r_high"}) {
		t.Fatalf("applied = %v, want [r_neg r_high]", applied)
	}
}

func TestDiffJSON(t *testing.T) {
	t.Parallel()

	before := domain.Layout{
		Slug:  "store-home",
		Title: "Abeta Fresh",
		Slots: []domain.LayoutSlot{
			{SlotID: "s1", WidgetID: "w_hero", Position: 0, Enabled: true},
			{SlotID: "s2", WidgetID: "w_grid", Position: 1, Enabled: true},
		},
		Tags: []string{"home", "v1"},
	}

	after := domain.Layout{
		Slug:  "store-home",
		Title: "Abeta Fresh Express",
		Slots: []domain.LayoutSlot{
			{SlotID: "s1", WidgetID: "w_promo", Position: 0, Enabled: true},
		},
		Tags:        []string{"home", "v1", "express"},
		Description: "Updated storefront",
	}

	diffs := engine.DiffJSON(before, after)
	if len(diffs) == 0 {
		t.Fatal("expected non-empty diffs")
	}

	// Verify deterministic sorted order by path.
	for i := 1; i < len(diffs); i++ {
		if diffs[i-1].Path > diffs[i].Path {
			t.Fatalf("diffs not sorted by path: %q > %q", diffs[i-1].Path, diffs[i].Path)
		}
	}

	byPath := make(map[string]domain.DiffItem, len(diffs))
	for _, d := range diffs {
		byPath[d.Path] = d
	}

	if d, ok := byPath["/title"]; !ok || d.Op != "replace" || d.Before != "Abeta Fresh" || d.After != "Abeta Fresh Express" {
		t.Errorf("unexpected /title diff: %+v", d)
	}
	if d, ok := byPath["/description"]; !ok || d.Op != "add" || d.After != "Updated storefront" {
		t.Errorf("unexpected /description diff: %+v", d)
	}
	if d, ok := byPath["/slots/0/widget_id"]; !ok || d.Op != "replace" || d.Before != "w_hero" || d.After != "w_promo" {
		t.Errorf("unexpected /slots/0/widget_id diff: %+v", d)
	}
	if d, ok := byPath["/slots/1"]; !ok || d.Op != "remove" {
		t.Errorf("unexpected /slots/1 diff: %+v", d)
	}
	if d, ok := byPath["/tags/2"]; !ok || d.Op != "add" || d.After != "express" {
		t.Errorf("unexpected /tags/2 diff: %+v", d)
	}

	if eq := engine.DiffJSON(before, before); len(eq) != 0 {
		t.Fatalf("identical values produced non-empty diff: %+v", eq)
	}
	if nilToVal := engine.DiffJSON(nil, map[string]any{"a": 1}); len(nilToVal) != 1 || nilToVal[0].Op != "add" {
		t.Fatalf("unexpected nil->val diff: %+v", nilToVal)
	}
	if valToNil := engine.DiffJSON(map[string]any{"a": 1}, nil); len(valToNil) != 1 || valToNil[0].Op != "remove" {
		t.Fatalf("unexpected val->nil diff: %+v", valToNil)
	}
}

func TestLayoutValidator(t *testing.T) {
	t.Parallel()

	validator := engine.NewLayoutValidator()

	supported := engine.SupportedSnippetTypes()
	if len(supported) != 17 {
		t.Fatalf("SupportedSnippetTypes() len = %d, want 17", len(supported))
	}
	for _, st := range supported {
		if !engine.IsValidSnippetType(st) {
			t.Errorf("IsValidSnippetType(%q) = false, want true", st)
		}
	}
	if engine.IsValidSnippetType("unknown_widget_v9") {
		t.Error("IsValidSnippetType(unknown_widget_v9) = true, want false")
	}

	widgets := map[string]domain.Widget{
		"w_header": {ID: "w_header", SnippetType: snippets.SnippetTypeHeader, Active: true},
		"w_prod":   {ID: "w_prod", SnippetType: snippets.SnippetTypeProductCard, Active: true},
		"w_grid": {
			ID:          "w_grid",
			SnippetType: snippets.SnippetTypeGrid,
			Active:      true,
			Props: map[string]any{
				"columns":          2,
				"child_widget_ids": []string{"w_prod"},
			},
		},
		"w_tabs": {
			ID:          "w_tabs",
			SnippetType: snippets.SnippetTypeTabs,
			Active:      true,
			Props: map[string]any{
				"tabs": []any{
					map[string]any{
						"id":         "t1",
						"title":      "Fresh",
						"widget_ids": []string{"w_prod"},
					},
				},
			},
		},
	}

	validLayout := domain.Layout{
		Slug:  "store-home",
		Title: "Abeta Fresh",
		PageConfig: domain.LayoutPageConfig{
			AdaptiveLayout: &snippets.AdaptivePageSpec{LayoutMode: snippets.LayoutModeSplitPane},
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "slot_1", WidgetID: "w_header", Position: 0, Enabled: true},
			{SlotID: "slot_2", WidgetID: "w_grid", Position: 1, Enabled: true},
			{SlotID: "slot_3", WidgetID: "w_tabs", Position: 2, Enabled: true},
			{SlotID: "slot_4_disabled", WidgetID: "does_not_exist", Position: 3, Enabled: false},
		},
	}

	if issues := validator.ValidateLayout(validLayout, widgets); len(issues) != 0 {
		t.Fatalf("expected valid layout to have 0 issues, got: %+v", issues)
	}

	// Test invalid cases including cyclic container references.
	badWidgets := map[string]domain.Widget{
		"w_bad_type": {ID: "w_bad_type", SnippetType: "custom_unregistered", Active: true},
		"w_cycle_a": {
			ID:          "w_cycle_a",
			SnippetType: snippets.SnippetTypeCarousel,
			Props:       map[string]any{"child_widget_ids": []string{"w_cycle_b"}},
		},
		"w_cycle_b": {
			ID:          "w_cycle_b",
			SnippetType: snippets.SnippetTypeGrid,
			Props:       map[string]any{"columns": 0, "child_widget_ids": []string{"w_cycle_a"}},
		},
		"w_bad_tabs": {
			ID:          "w_bad_tabs",
			SnippetType: snippets.SnippetTypeTabs,
			Props: map[string]any{
				"tabs": []any{
					map[string]any{
						"id":         "",
						"title":      "",
						"widget_ids": []string{"w_missing_child"},
					},
				},
			},
		},
	}

	badLayout := domain.Layout{
		Slug:  "Bad Slug!",
		Title: "",
		PageConfig: domain.LayoutPageConfig{
			AdaptiveLayout: &snippets.AdaptivePageSpec{LayoutMode: "floating_grid"},
		},
		Slots: []domain.LayoutSlot{
			{SlotID: "", WidgetID: "w_bad_type", Enabled: true},
			{SlotID: "dup", WidgetID: "w_missing", Enabled: true},
			{SlotID: "dup", WidgetID: "w_cycle_a", Enabled: true},
			{SlotID: "tabs_slot", WidgetID: "w_bad_tabs", Enabled: true},
		},
	}

	issues := validator.ValidateLayout(badLayout, badWidgets)
	codes := make(map[string]bool, len(issues))
	for _, iss := range issues {
		codes[iss.Code] = true
	}

	wantCodes := []string{
		"invalid_slug",
		"required",
		"invalid_layout_mode",
		"invalid_snippet_type",
		"missing_widget",
		"duplicate_slot_id",
		"invalid_columns",
		"cycle_detected",
		"missing_child_widget",
	}
	for _, code := range wantCodes {
		if !codes[code] {
			t.Errorf("expected validation issue code %q not found in issues: %+v", code, issues)
		}
	}
}

func TestRuleEvaluator_AdditionalEdgeCases(t *testing.T) {
	t.Parallel()

	eval := engine.NewRuleEvaluator()
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		rule domain.TargetingRule
		ctx  domain.RequestContext
		want bool
	}{
		{
			name: "in with comma-separated string",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "city", Operator: "in", Value: "delhi, bengaluru, pune"}},
			},
			ctx:  domain.RequestContext{City: "Bengaluru", Now: now},
			want: true,
		},
		{
			name: "in with []int on numeric attribute",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "tier", Operator: "in", Value: []int{1, 2, 3}}},
			},
			ctx:  domain.RequestContext{Attributes: map[string]string{"tier": "2"}, Now: now},
			want: true,
		},
		{
			name: "in with []float64 on numeric attribute",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "score", Operator: "in", Value: []float64{4.5, 5.0}}},
			},
			ctx:  domain.RequestContext{Attributes: map[string]string{"score": "4.5"}, Now: now},
			want: true,
		},
		{
			name: "semver prerelease numeric vs alpha precedence",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_lt", Value: "1.4.2-beta.1"}},
			},
			ctx:  domain.RequestContext{AppVersion: "1.4.2-alpha.2+build.9", Now: now},
			want: true,
		},
		{
			name: "semver prerelease numeric comparison",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_gte", Value: "1.4.2-rc.2"}},
			},
			ctx:  domain.RequestContext{AppVersion: "1.4.2-rc.10", Now: now},
			want: true,
		},
		{
			name: "semver numeric prerelease lower than alphanumeric",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "app_version", Operator: "semver_lt", Value: "1.0.0-alpha"}},
			},
			ctx:  domain.RequestContext{AppVersion: "1.0.0-1", Now: now},
			want: true,
		},
		{
			name: "bucket_pct fallback seed when user_id empty",
			rule: domain.TargetingRule{
				ID:         "rule_fallback_seed",
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "user_id", Operator: "bucket_pct", Value: 100}},
			},
			ctx:  domain.RequestContext{TenantSlug: "abeta", Platform: "web", City: "pune", Now: now},
			want: true,
		},
		{
			name: "case-insensitive custom attribute lookup",
			rule: domain.TargetingRule{
				Enabled:    true,
				Conditions: []domain.RuleCondition{{Field: "LoyaltyTier", Operator: "eq", Value: "platinum"}},
			},
			ctx:  domain.RequestContext{Attributes: map[string]string{"loyaltytier": "Platinum"}, Now: now},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := eval.MatchesRule(tc.rule, tc.ctx); got != tc.want {
				t.Fatalf("MatchesRule() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLayoutValidator_InlineContainersAndTabs(t *testing.T) {
	t.Parallel()

	v := engine.NewLayoutValidator()
	widgets := map[string]domain.Widget{
		"w_carousel_inline": {
			ID:          "w_carousel_inline",
			SnippetType: snippets.SnippetTypeCarousel,
			Props: map[string]any{
				"items": []any{
					map[string]any{"snippet_id": "", "snippet_type": "invalid_type"},
					"not_an_object",
				},
			},
		},
		"w_accordion_inline": {
			ID:          "w_accordion_inline",
			SnippetType: snippets.SnippetTypeAccordion,
			Props: map[string]any{
				"snippets": []snippets.SnippetEnvelope{
					{SnippetID: "", SnippetType: "bad_snippet"},
				},
				"child_widget_ids": []any{123},
			},
		},
		"w_tabs_typed": {
			ID:          "w_tabs_typed",
			SnippetType: snippets.SnippetTypeTabs,
			Props: map[string]any{
				"tabs": []snippets.TabItemSpec{
					{
						ID:       "t1",
						Title:    "Valid Tab",
						Snippets: []snippets.SnippetEnvelope{{SnippetID: "s_ok", SnippetType: snippets.SnippetTypeHeader}},
					},
					{
						ID:    "",
						Title: "",
					},
				},
			},
		},
	}

	layout := domain.Layout{
		Slug:  "inline-check",
		Title: "Inline Container Check",
		Slots: []domain.LayoutSlot{
			{SlotID: "s1", WidgetID: "w_carousel_inline", Enabled: true},
			{SlotID: "s2", WidgetID: "w_accordion_inline", Enabled: true},
			{SlotID: "s3", WidgetID: "w_tabs_typed", Enabled: true},
		},
	}

	issues := v.ValidateLayout(layout, widgets)
	codes := make(map[string]bool, len(issues))
	for _, iss := range issues {
		codes[iss.Code] = true
	}

	for _, want := range []string{"required", "invalid_snippet_type", "invalid_container_item", "invalid_widget_refs"} {
		if !codes[want] {
			t.Errorf("expected issue code %q not found in %+v", want, issues)
		}
	}
}
