package version

import (
	"fmt"
	"os"
	"testing"
)

type tagTest struct {
	tag string
	bt  BuildType
	exp string
	env map[string]string
	err bool
}

// Mock out the build environment so we can test environment-dependent version info.
type mockBuildEnv struct {
	ci, nightly bool
	tag         string
}

func (m *mockBuildEnv) IsCI() bool {
	return m.ci
}

func (m *mockBuildEnv) IsNightly() bool {
	return m.nightly
}

func (m *mockBuildEnv) GetMostRecentTag() string {
	return m.tag
}

func TestBuildTypeFromEnv(t *testing.T) {
	tests := []struct {
		buildEnv BuildEnv
		exp      BuildType
	}{
		// if building outside of CI, this is a dev build
		{
			buildEnv: &mockBuildEnv{tag: "2.0.0-alpha.1-1"},
			exp:      Dev,
		},
		// if tag is not exact match for current HEAD, this is a nightly build
		{
			buildEnv: &mockBuildEnv{tag: "2.0.0-beta.1-1", ci: true, nightly: true},
			exp:      Nightly,
		},
		// the following tests use string matching to determine build type
		{
			buildEnv: &mockBuildEnv{tag: "2.0.0-alpha.17-1", ci: true},
			exp:      Alpha,
		},
		{
			buildEnv: &mockBuildEnv{tag: "2.0.0-beta.18-1", ci: true},
			exp:      Beta,
		},
		{
			buildEnv: &mockBuildEnv{tag: "v2.0.0-rc-2", ci: true},
			exp:      RC,
		},
		// if no string can be matched, fall back to stable
		{
			buildEnv: &mockBuildEnv{tag: "2.0.1", ci: true},
			exp:      Stable,
		},
	}

	for _, test := range tests {
		_, got := ParseBuildEnv(test.buildEnv)
		if got != test.exp {
			t.Errorf("bad build type: got %q, want %q", got, test.exp)
		}
	}
}

func TestIteration(t *testing.T) {
	tests := []tagTest{
		{
			tag: "v2.0.0-rc-2",
			bt:  RC,
			exp: "2",
		},
		{
			tag: "2.0.0-alpha.17-1",
			bt:  Alpha,
			exp: "1",
		},
		{
			tag: "2.0.0-beta.18-2",
			bt:  Beta,
			exp: "2",
		},
		{
			tag: "2.0.0-alpha.2-1",
			bt:  Nightly,
			err: true, // nightly build needs env vars
		},
		{
			tag: "2.0.0-alpha.2-1",
			bt:  Nightly,
			env: map[string]string{
				"SENSU_BUILD_ITERATION": "7",
			},
			exp: "7",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for k, v := range test.env {
				if err := os.Setenv(k, v); err != nil {
					t.Fatal(err)
				}
			}
			defer os.Clearenv()
			it, err := Iteration(test.tag, test.bt)
			if test.err && err == nil {
				t.Fatalf("expected error, got iteration %q", it)
			}
			if !test.err && err != nil {
				t.Fatal(err)
			}
			if got, want := it, test.exp; got != want {
				t.Errorf("bad iteration: got %q, want %q", got, want)
			}
		})
	}
}

func TestGetPrereleaseVersion(t *testing.T) {
	tests := []tagTest{
		{
			tag: "v2.0.0-rc-2",
			bt:  RC,
			err: true, // no iteration to parse
		},
		{
			tag: "2.0.0-alpha.17-1",
			bt:  Alpha,
			exp: "17",
		},
		{
			tag: "2.0.0-beta.18-2",
			bt:  Beta,
			exp: "18",
		},
		// prerelease version does not apply to Dev or Nightly builds
		{
			bt:  Dev,
			exp: "",
		},
		{
			bt:  Nightly,
			exp: "",
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ver, err := GetPrereleaseVersion(test.tag, test.bt)
			if test.err && err == nil {
				t.Fatal("expected error")
			}
			if !test.err && err != nil {
				t.Fatal(err)
			}
			if got, want := ver, test.exp; got != want {
				t.Errorf("bad prelease version: got %q, want %q", got, want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	tests := []tagTest{
		{
			tag: "",
			bt:  Dev,
			err: true, // could not determine base version from tag
		},
		{
			tag: "v2.0.0-rc-2",
			bt:  RC,
			err: true, // could not parse prerelease version
		},
		{
			tag: "2.0.0-alpha.17-1",
			bt:  Alpha,
			exp: "2.0.0-alpha.17",
		},
		{
			tag: "2.0.0-beta.18-2",
			bt:  Beta,
			exp: "2.0.0-beta.18",
		},
		{
			tag: "2.0.0-beta.20-1",
			bt:  Dev,
			exp: "2.0.0-dev",
		},
		{
			tag: "2.0.0-beta.20-1",
			bt:  Nightly,
			exp: "2.0.0-nightly",
		},
		{
			// nightly build with env var tag
			tag: "2.0.0",
			bt:  Stable,
			// TODO the env is not being parsed anywhere
			env: map[string]string{
				"TRAVIS":              "true",
				"TRAVIS_BUILD_NUMBER": "42",
			},
			exp: "2.0.0",
		},
		{
			tag: "2.0.0",
			bt:  Stable,
			// TODO the env is not being parsed anywhere
			env: map[string]string{
				"APPVEYOR":              "true",
				"APPVEYOR_BUILD_NUMBER": "666",
			},
			exp: "2.0.0",
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ver, err := GetVersion(test.tag, test.bt)
			if test.err && err == nil {
				t.Fatal("expected error")
			}
			if !test.err && err != nil {
				t.Fatal(err)
			}
			if got, want := ver, test.exp; got != want {
				t.Errorf("bad version: got %q, want %q", got, want)
			}
		})
	}
}

func TestFullVersion(t *testing.T) {
	tests := []tagTest{
		{
			tag: "v2.0.0-rc-2",
			bt:  RC,
			err: true, // could not parse base version
		},
		{
			tag: "2.0.0-alpha.17-1",
			bt:  Alpha,
			exp: "2.0.0-alpha.17-1",
		},
		{
			tag: "2.0.0-beta.18-2",
			bt:  Beta,
			exp: "2.0.0-beta.18-2",
		},
		{
			tag: "2.0.0-rc.17-1",
			bt:  Nightly,
			// nightly build with env var tag
			env: map[string]string{
				"SENSU_BUILD_ITERATION": "7",
			},
			exp: "2.0.0-nightly-7",
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for k, v := range test.env {
				if err := os.Setenv(k, v); err != nil {
					t.Fatal(err)
				}
			}
			defer os.Clearenv()
			ver, err := FullVersion(test.tag, test.bt)
			if test.err && err == nil {
				t.Fatal("expected error")
			}
			if !test.err && err != nil {
				t.Fatal(err)
			}
			if got, want := ver, test.exp; got != want {
				t.Errorf("bad fullversion: got %q, want %q", got, want)
			}
		})
	}
}
