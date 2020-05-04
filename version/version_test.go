package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintln(t *testing.T) {
	goVersion := "go1.14.2"
	tests := []struct {
		name      string
		component string
		version   string
		edition   string
		buildDate string
		buildSHA  string
		goVersion string
		want      string
	}{
		{
			name: "component & version specified",
			component: "testing",
			version:   "1.2.3",
			goVersion: goVersion,
			want: "testing version 1.2.3+ce, community edition, built with go1.14.2",
		},
		{
			name: "component, version & buildSHA specified, nil edition",
			component: "testing",
			version:   "1.2.3",
			buildSHA:  "387c20615518f1325528705e0ef09e4d30d80378",
			goVersion: goVersion,
			want: "testing version 1.2.3+ce, community edition, build 387c20615518f1325528705e0ef09e4d30d80378, built with go1.14.2",
		},
		{
			name: "component, version, buildDate, & buildSHA specified",
			component: "testing",
			version:   "1.2.3",
			buildDate: "2019-10-09",
			buildSHA:  "387c20615518f1325528705e0ef09e4d30d80378",
			goVersion: goVersion,
			want: "testing version 1.2.3+ce, community edition, build 387c20615518f1325528705e0ef09e4d30d80378, built 2019-10-09, built with go1.14.2",
		},
		{
			name: "component, version, and enterprise edition specified",
			component: "testing",
			version:   "1.2.3",
			edition:   "enterprise",
			goVersion: goVersion,
			want: "testing version 1.2.3+ee, enterprise edition, built with go1.14.2",
		},
		{
			name: "component, version, and invalid edition specified",
			component: "testing",
			version:   "1.2.3",
			edition:   "fake",
			goVersion: goVersion,
			want: "testing version 1.2.3+invalid, built with an invalid \"edition\" ldflag, built with go1.14.2",
		},
		{
			name: "component, version, and invalid edition specified",
			component: "testing",
			version:   "1.2.3",
			edition:   "fake",
			goVersion: goVersion,
			want: "testing version 1.2.3+invalid, built with an invalid \"edition\" ldflag, built with go1.14.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			if tt.edition != "" {
				Edition = tt.edition
			}
			BuildDate = tt.buildDate
			BuildSHA = tt.buildSHA
			GoVersion = tt.goVersion
			assert.EqualValues(t, tt.want, FormattedOutput(tt.component))
		})
	}
}
