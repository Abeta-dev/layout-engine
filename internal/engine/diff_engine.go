package engine

import (
	"cmp"
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/abeta-dev/layout-engine/internal/domain"
)

// DiffJSON normalizes before and after into generic JSON structures and computes a
// deterministic, path-sorted slice of field-level DiffItems using JSON Pointer paths
// (for example "/title" or "/slots/0/widget_id").
func DiffJSON(before, after any) []domain.DiffItem {
	normBefore := normalizeForDiff(before)
	normAfter := normalizeForDiff(after)

	items := make([]domain.DiffItem, 0)
	if reflect.DeepEqual(normBefore, normAfter) {
		return items
	}

	if normBefore == nil && normAfter != nil {
		return []domain.DiffItem{{Path: "/", Op: "add", After: normAfter}}
	}
	if normBefore != nil && normAfter == nil {
		return []domain.DiffItem{{Path: "/", Op: "remove", Before: normBefore}}
	}

	diffRecursive("", normBefore, normAfter, &items)
	slices.SortFunc(items, func(a, b domain.DiffItem) int {
		if c := cmp.Compare(a.Path, b.Path); c != 0 {
			return c
		}
		return cmp.Compare(a.Op, b.Op)
	})
	return items
}

func normalizeForDiff(v any) any {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return v
	}
	return out
}

func diffRecursive(path string, before, after any, items *[]domain.DiffItem) {
	if reflect.DeepEqual(before, after) {
		return
	}

	bMap, bIsMap := before.(map[string]any)
	aMap, aIsMap := after.(map[string]any)
	if bIsMap && aIsMap {
		keys := make([]string, 0, len(bMap)+len(aMap))
		seen := make(map[string]struct{}, len(bMap)+len(aMap))
		for k := range bMap {
			keys = append(keys, k)
			seen[k] = struct{}{}
		}
		for k := range aMap {
			if _, ok := seen[k]; !ok {
				keys = append(keys, k)
			}
		}
		slices.Sort(keys)

		for _, k := range keys {
			childPath := joinJSONPointer(path, k)
			bv, bOk := bMap[k]
			av, aOk := aMap[k]
			switch {
			case !bOk && aOk:
				*items = append(*items, domain.DiffItem{
					Path:  childPath,
					Op:    "add",
					After: av,
				})
			case bOk && !aOk:
				*items = append(*items, domain.DiffItem{
					Path:   childPath,
					Op:     "remove",
					Before: bv,
				})
			default:
				diffRecursive(childPath, bv, av, items)
			}
		}
		return
	}

	bSlice, bIsSlice := before.([]any)
	aSlice, aIsSlice := after.([]any)
	if bIsSlice && aIsSlice {
		minLen := len(bSlice)
		if len(aSlice) < minLen {
			minLen = len(aSlice)
		}
		for i := 0; i < minLen; i++ {
			diffRecursive(joinJSONPointer(path, strconv.Itoa(i)), bSlice[i], aSlice[i], items)
		}
		for i := minLen; i < len(aSlice); i++ {
			*items = append(*items, domain.DiffItem{
				Path:  joinJSONPointer(path, strconv.Itoa(i)),
				Op:    "add",
				After: aSlice[i],
			})
		}
		for i := minLen; i < len(bSlice); i++ {
			*items = append(*items, domain.DiffItem{
				Path:   joinJSONPointer(path, strconv.Itoa(i)),
				Op:     "remove",
				Before: bSlice[i],
			})
		}
		return
	}

	targetPath := path
	if targetPath == "" {
		targetPath = "/"
	}
	*items = append(*items, domain.DiffItem{
		Path:   targetPath,
		Op:     "replace",
		Before: before,
		After:  after,
	})
}

func joinJSONPointer(base, segment string) string {
	escaped := strings.ReplaceAll(segment, "~", "~0")
	escaped = strings.ReplaceAll(escaped, "/", "~1")
	return base + "/" + escaped
}
