package version

import (
	"strings"

	"github.com/Masterminds/semver/v3"
	"k8s.io/klog/v2"
)

const (
	// minFormalVersion is the lower bound (inclusive) for formal releases
	// (versions without any prerelease suffix).
	minFormalVersion = "1.12.6"

	// minPrereleaseVersion is the lower bound (inclusive) for general
	// prerelease builds (e.g. dated build numbers).
	minPrereleaseVersion = "1.12.7-20260507"
)

// allowedBaselinePrereleaseTags lists prerelease prefixes that are accepted at
// the minFormalVersion baseline. e.g. "1.12.6-rc1" and "1.12.6-beta2" are
// considered supported even though they fall below minPrereleaseVersion.
var allowedBaselinePrereleaseTags = []string{"rc", "beta", "alpha"}

// IsSupported reports whether version satisfies the supported range:
//
//   - formal release (no prerelease suffix): >= 1.12.6
//   - prerelease at the 1.12.6 baseline tagged "rc*" or "beta*"
//     (e.g. 1.12.6-rc1, 1.12.6-beta2)
//   - any other prerelease build: >= 1.12.7-20260507
//
// Returns false (without error) when the input cannot be parsed as semver.
func IsSupported(ver string) bool {
	v, err := semver.NewVersion(ver)
	if err != nil {
		klog.V(4).Infof("version.IsSupported: invalid semver %q: %v", ver, err)
		return false
	}

	if v.Prerelease() == "" {
		return v.GreaterThanEqual(semver.MustParse(minFormalVersion))
	}

	if isAllowedBaselinePrerelease(v) {
		return true
	}

	return v.GreaterThanEqual(semver.MustParse(minPrereleaseVersion))
}

// isAllowedBaselinePrerelease reports whether v sits exactly on the
// minFormalVersion (e.g. 1.12.6) and carries a prerelease tag listed in
// allowedBaselinePrereleaseTags.
func isAllowedBaselinePrerelease(v *semver.Version) bool {
	baseline := semver.MustParse(minFormalVersion)
	if v.Major() != baseline.Major() ||
		v.Minor() != baseline.Minor() ||
		v.Patch() != baseline.Patch() {
		return false
	}
	pre := v.Prerelease()
	for _, tag := range allowedBaselinePrereleaseTags {
		if strings.HasPrefix(pre, tag) {
			return true
		}
	}
	return false
}
