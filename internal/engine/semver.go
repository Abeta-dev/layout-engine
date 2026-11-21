package engine

import (
	"strconv"
	"strings"
)

type semver struct {
	major      int
	minor      int
	patch      int
	prerelease []string
}

func parseSemver(raw string) (semver, bool) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "v"), "V")
	if s == "" {
		return semver{}, false
	}
	if idx := strings.IndexByte(s, '+'); idx >= 0 {
		s = s[:idx]
	}

	var prerelease []string
	if idx := strings.IndexByte(s, '-'); idx >= 0 {
		pre := s[idx+1:]
		if pre != "" {
			prerelease = strings.Split(pre, ".")
		}
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return semver{}, false
	}

	nums := [3]int{}
	for i, part := range parts {
		if part == "" {
			return semver{}, false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return semver{}, false
		}
		nums[i] = n
	}

	return semver{
		major:      nums[0],
		minor:      nums[1],
		patch:      nums[2],
		prerelease: prerelease,
	}, true
}

func compareSemver(a, b semver) int {
	if a.major != b.major {
		return compareInt(a.major, b.major)
	}
	if a.minor != b.minor {
		return compareInt(a.minor, b.minor)
	}
	if a.patch != b.patch {
		return compareInt(a.patch, b.patch)
	}

	aPre := len(a.prerelease) > 0
	bPre := len(b.prerelease) > 0
	if !aPre && !bPre {
		return 0
	}
	if !aPre && bPre {
		return 1
	}
	if aPre && !bPre {
		return -1
	}

	limit := len(a.prerelease)
	if len(b.prerelease) < limit {
		limit = len(b.prerelease)
	}
	for i := 0; i < limit; i++ {
		ai, aErr := strconv.Atoi(a.prerelease[i])
		bi, bErr := strconv.Atoi(b.prerelease[i])
		switch {
		case aErr == nil && bErr == nil:
			if ai != bi {
				return compareInt(ai, bi)
			}
		case aErr == nil && bErr != nil:
			return -1
		case aErr != nil && bErr == nil:
			return 1
		default:
			if a.prerelease[i] != b.prerelease[i] {
				if a.prerelease[i] < b.prerelease[i] {
					return -1
				}
				return 1
			}
		}
	}
	return compareInt(len(a.prerelease), len(b.prerelease))
}

func compareSemverOp(op, actualRaw, expectedRaw string) bool {
	a, aOk := parseSemver(actualRaw)
	b, bOk := parseSemver(expectedRaw)
	if !aOk || !bOk {
		return false
	}
	cmp := compareSemver(a, b)
	switch op {
	case "semver_gte":
		return cmp >= 0
	case "semver_lt":
		return cmp < 0
	default:
		return false
	}
}

func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
