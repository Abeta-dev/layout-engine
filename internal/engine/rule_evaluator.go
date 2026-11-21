// Package engine implements pure, deterministic domain engines for targeting rule
// evaluation, dynamic widget slotting, JSON diffing, and layout validation.
package engine

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/abeta-dev/layout-engine/internal/domain"
)

// Compile-time interface verification.
var _ domain.RuleEvaluator = (*DefaultRuleEvaluator)(nil)

// DefaultRuleEvaluator deterministically evaluates targeting rules and schedule windows.
type DefaultRuleEvaluator struct{}

// NewRuleEvaluator returns a new DefaultRuleEvaluator.
func NewRuleEvaluator() *DefaultRuleEvaluator {
	return &DefaultRuleEvaluator{}
}

// MatchesSchedule evaluates whether now falls within window.
// A nil window always matches. DaysOfWeek uses 0=Sunday..6=Saturday in UTC.
// StartHourUTC..EndHourUTC uses half-open [start, end) semantics when start != end,
// supporting overnight wrap-around ranges such as 22..4 (hours 22, 23, 0, 1, 2, 3),
// and single-hour equality when start == end.
func (e *DefaultRuleEvaluator) MatchesSchedule(window *domain.ScheduleWindow, now time.Time) bool {
	if window == nil {
		return true
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	nowUTC := now.UTC()

	if window.StartAt != nil && nowUTC.Before(window.StartAt.UTC()) {
		return false
	}
	if window.EndAt != nil && nowUTC.After(window.EndAt.UTC()) {
		return false
	}
	if len(window.DaysOfWeek) > 0 {
		today := int(nowUTC.Weekday())
		dayMatched := false
		for _, d := range window.DaysOfWeek {
			if d == today {
				dayMatched = true
				break
			}
		}
		if !dayMatched {
			return false
		}
	}

	if window.StartHourUTC != nil || window.EndHourUTC != nil {
		hour := nowUTC.Hour()
		switch {
		case window.StartHourUTC != nil && window.EndHourUTC != nil:
			start := *window.StartHourUTC
			end := *window.EndHourUTC
			if start == end {
				if hour != start {
					return false
				}
			} else if start < end {
				if hour < start || hour >= end {
					return false
				}
			} else {
				// Overnight wrap-around window, e.g., 22..4 matches hours 22, 23, 0, 1, 2, 3.
				if hour < start && hour >= end {
					return false
				}
			}
		case window.StartHourUTC != nil:
			if hour < *window.StartHourUTC {
				return false
			}
		case window.EndHourUTC != nil:
			if hour >= *window.EndHourUTC {
				return false
			}
		}
	}

	return true
}

// MatchesRule evaluates whether rule is enabled, inside its schedule window, and satisfies
// its conditions according to MatchMode ("any" for OR, default "all" for AND).
func (e *DefaultRuleEvaluator) MatchesRule(rule domain.TargetingRule, ctx domain.RequestContext) bool {
	if !rule.Enabled {
		return false
	}

	now := ctx.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if !e.MatchesSchedule(rule.Schedule, now) {
		return false
	}

	if len(rule.Conditions) == 0 {
		return true
	}

	matchAny := strings.EqualFold(strings.TrimSpace(rule.MatchMode), "any")
	for _, cond := range rule.Conditions {
		matched := evalCondition(rule, cond, ctx)
		if matchAny && matched {
			return true
		}
		if !matchAny && !matched {
			return false
		}
	}
	return !matchAny
}

func evalCondition(rule domain.TargetingRule, cond domain.RuleCondition, ctx domain.RequestContext) bool {
	op := strings.ToLower(strings.TrimSpace(cond.Operator))
	actual, exists := extractContextField(cond.Field, ctx)

	switch op {
	case "eq":
		if !exists {
			return false
		}
		return valuesEqual(actual, cond.Value)
	case "neq":
		if !exists {
			return !valuesEqual("", cond.Value)
		}
		return !valuesEqual(actual, cond.Value)
	case "in":
		if !exists {
			return false
		}
		return valueInList(actual, cond.Value)
	case "not_in":
		if !exists {
			return !valueInList("", cond.Value)
		}
		return !valueInList(actual, cond.Value)
	case "gt", "gte", "lt", "lte":
		if !exists {
			return false
		}
		return compareNumeric(op, actual, cond.Value)
	case "semver_gte", "semver_lt":
		if !exists {
			return false
		}
		return compareSemverOp(op, valueToString(actual), valueToString(cond.Value))
	case "contains":
		if !exists {
			return false
		}
		return strings.Contains(
			strings.ToLower(valueToString(actual)),
			strings.ToLower(valueToString(cond.Value)),
		)
	case "regex":
		if !exists {
			return false
		}
		re, err := regexp.Compile(valueToString(cond.Value))
		if err != nil {
			return false
		}
		return re.MatchString(valueToString(actual))
	case "bucket_pct":
		return evalBucketPct(rule, cond, actual, exists, ctx)
	default:
		return false
	}
}

func extractContextField(field string, ctx domain.RequestContext) (any, bool) {
	trimmed := strings.TrimSpace(field)
	switch strings.ToLower(trimmed) {
	case "platform":
		return ctx.Platform, true
	case "app_version":
		return ctx.AppVersion, true
	case "city":
		return ctx.City, true
	case "zone":
		return ctx.Zone, true
	case "segment":
		return ctx.Segment, true
	case "user_id":
		return ctx.UserID, true
	case "window_class":
		return ctx.WindowClass, true
	case "dark_mode":
		return ctx.DarkMode, true
	case "tenant_slug":
		return ctx.TenantSlug, true
	}

	if len(ctx.Attributes) == 0 || trimmed == "" {
		return "", false
	}
	if v, ok := ctx.Attributes[trimmed]; ok {
		return v, true
	}
	for k, v := range ctx.Attributes {
		if strings.EqualFold(k, trimmed) {
			return v, true
		}
	}
	return "", false
}

func evalBucketPct(
	rule domain.TargetingRule,
	cond domain.RuleCondition,
	actual any,
	exists bool,
	ctx domain.RequestContext,
) bool {
	targetPct, ok := toFloat64(cond.Value)
	if !ok || math.IsNaN(targetPct) || targetPct <= 0 {
		return false
	}
	if targetPct >= 100 {
		return true
	}

	seed := ""
	if exists {
		seed = strings.TrimSpace(valueToString(actual))
	}
	if seed == "" {
		seed = strings.TrimSpace(ctx.UserID)
	}
	if seed == "" {
		seed = strings.Join([]string{
			ctx.TenantSlug,
			ctx.Platform,
			ctx.City,
			ctx.Zone,
			ctx.Segment,
		}, ":")
	}

	ruleKey := strings.TrimSpace(rule.Key)
	if ruleKey == "" {
		ruleKey = rule.ID
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(ruleKey + ":" + seed))
	bucket := h.Sum32() % 100
	return float64(bucket) < targetPct
}

func valuesEqual(actual, expected any) bool {
	if ab, ok := actual.(bool); ok {
		eb, ok := toBool(expected)
		return ok && ab == eb
	}
	if eb, ok := expected.(bool); ok {
		ab, ok := toBool(actual)
		return ok && ab == eb
	}
	if isNumericValue(actual) || isNumericValue(expected) {
		af, aOk := toFloat64(actual)
		ef, eOk := toFloat64(expected)
		if aOk && eOk && !math.IsNaN(af) && !math.IsNaN(ef) {
			return af == ef
		}
	}
	return strings.EqualFold(
		strings.TrimSpace(valueToString(actual)),
		strings.TrimSpace(valueToString(expected)),
	)
}

func valueInList(actual, listVal any) bool {
	switch list := listVal.(type) {
	case []string:
		for _, item := range list {
			if valuesEqual(actual, item) {
				return true
			}
		}
		return false
	case []any:
		for _, item := range list {
			if valuesEqual(actual, item) {
				return true
			}
		}
		return false
	case []int:
		for _, item := range list {
			if valuesEqual(actual, item) {
				return true
			}
		}
		return false
	case []float64:
		for _, item := range list {
			if valuesEqual(actual, item) {
				return true
			}
		}
		return false
	case string:
		for _, part := range strings.Split(list, ",") {
			if valuesEqual(actual, strings.TrimSpace(part)) {
				return true
			}
		}
		return false
	default:
		return valuesEqual(actual, listVal)
	}
}

func compareNumeric(op string, actual, expected any) bool {
	a, aOk := toFloat64(actual)
	b, bOk := toFloat64(expected)
	if !aOk || !bOk || math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	switch op {
	case "gt":
		return a > b
	case "gte":
		return a >= b
	case "lt":
		return a < b
	case "lte":
		return a <= b
	default:
		return false
	}
}

func isNumericValue(v any) bool {
	switch v.(type) {
	case int, int64, float32, float64, json.Number:
		return true
	default:
		return false
	}
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	case string:
		s := strings.TrimSpace(val)
		if s == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func toBool(v any) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(val))
		return b, err == nil
	default:
		return false, false
	}
}

func valueToString(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case json.Number:
		return val.String()
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprint(val)
	}
}
