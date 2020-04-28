package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintln(t *testing.T) {
	tests := []struct {
		name      string
		component string
		version   string
		edition   string
		buildDate string
		buildSHA  string
		want      string
	}{
		{
			name: "component & version specified",
			component: "testing",
			version:   "1.2.3",
			want: "testing version 1.2.3+ce, community edition",
		},
		{
			name: "component, version & buildSHA specified, nil edition",
			component: "testing",
			version:   "1.2.3",
			buildSHA:  "387c20615518f1325528705e0ef09e4d30d80378",
			want: "testing version 1.2.3+ce, community edition, build 387c20615518f1325528705e0ef09e4d30d80378",
		},
		{
			name: "component, version, buildDate, & buildSHA specified",
			component: "testing",
			version:   "1.2.3",
			buildDate: "2019-10-09",
			buildSHA:  "387c20615518f1325528705e0ef09e4d30d80378",
			want: "testing version 1.2.3+ce, community edition, build 387c20615518f1325528705e0ef09e4d30d80378, built 2019-10-09",
		},
		{
			name: "component, version, and enterprise edition specified",
			component: "testing",
			version:   "1.2.3",
			edition:   "enterprise",
			want: "testing version 1.2.3+ee, enterprise edition",
		},
		{
			name: "component, version, and invalid edition specified",
			component: "testing",
			version:   "1.2.3",
			edition:   "fake",
			want: "testing version 1.2.3+invalid, built with an invalid \"edition\" ldflag",
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
			assert.EqualValues(t, tt.want, FormattedOutput(tt.component))
		})
	}
}
