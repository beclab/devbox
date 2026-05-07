package version

import "testing"

func TestIsSupported(t *testing.T) {
	cases := []struct {
		name    string
		version string
		want    bool
	}{
		{"formal: greater than threshold", "1.12.7", true},
		{"formal: with v prefix", "v1.12.7", true},
		{"formal: minor bump", "1.13.0", true},
		{"formal: major bump", "2.0.0", true},

		{"formal: equal to threshold", "1.12.6", true},
		{"formal: less than threshold (patch)", "1.12.5", false},
		{"formal: less than threshold (minor)", "1.11.99", false},
		{"formal: less than threshold (major)", "0.99.99", false},

		{"prerelease: greater build number", "1.12.7-20260508", true},
		{"prerelease: future build number", "1.12.7-20270101", true},
		{"prerelease: greater base version", "1.13.0-rc1", true},
		{"prerelease: greater base version 0", "1.13.0-0", true},
		{
			// SemVer 2.0.0: numeric prerelease identifiers have lower
			// precedence than alphanumeric ones, so "rc1" > "20260507".
			name:    "prerelease: alphanumeric tag at same base",
			version: "1.12.7-rc1",
			want:    true,
		},

		{"prerelease: equal to threshold", "1.12.7-20260507", true},
		{"prerelease: smaller build number", "1.12.7-20260506", false},
		{"prerelease: smaller base version", "1.12.6-rc1", true},
		{"prerelease: smaller base version", "1.12.6-rc.1", true},
		{"prerelease: smaller base version", "1.12.6-beta.1", true},
		{"prerelease: smaller base version", "1.12.6-alpha.1", true},

		{"prerelease: smaller base version 0", "1.12.6-0", false},

		{"empty string", "", false},
		{"garbage", "not-a-version", false},
		{"incomplete: only major.minor", "1.12", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsSupported(c.version); got != c.want {
				t.Errorf("IsSupported(%q) = %v, want %v", c.version, got, c.want)
			}
		})
	}
}
