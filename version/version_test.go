package version

import (
	"fmt"
	"os"
	"testing"
)

type tagTest struct {
	tag string
	exp string
	env map[string]string
	err bool
}

func TestBuildTypeFromTag(t *testing.T) {
	tests := []tagTest{
		{
			tag: "v2.0.0-rc-2",
			exp: string(RC),
		},
		{
			tag: "2.0.0-alpha.17-1",
			exp: string(Alpha),
		},
		{
			tag: "2.0.0-beta.18-1",
			exp: string(Beta),
		},
		{
			tag: "",
			exp: string(Nightly),
		},
		{
			tag: "2.0.1",
			exp: string(Stable),
		},
	}

	for _, test := range tests {
		got := BuildTypeFromTag(test.tag)
		if string(got) != test.exp {
			t.Errorf("bad build type: got %q, want %q", got, test.exp)
		}
	}
}

func TestIteration(t *testing.T) {
	tests := []tagTest{
		{
			tag: "v2.0.0-rc-2",
			exp: "2",
		},
		{
			tag: "2.0.0-alpha.17-1",
			exp: "1",
		},
		{
			tag: "2.0.0-beta.18-2",
			exp: "2",
		},
		{
			tag: "",
			err: true, // nightly build needs env vars
		},
		{
			// nightly build with env var tag
			env: map[string]string{
				"TRAVIS":              "true",
				"TRAVIS_BUILD_NUMBER": "42",
			},
			exp: "42",
		},
		{
			env: map[string]string{
				"APPVEYOR":              "true",
				"APPVEYOR_BUILD_NUMBER": "666",
			},
			exp: "666",
		},
		{
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
			it, err := Iteration(test.tag)
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
			err: true,
		},
		{
			tag: "2.0.0-alpha.17-1",
			exp: "17",
		},
		{
			tag: "2.0.0-beta.18-2",
			exp: "18",
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ver, err := GetPrereleaseVersion(test.tag)
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
			tag: "v2.0.0-rc-2",
			err: true,
		},
		{
			tag: "2.0.0-alpha.17-1",
			exp: "2.0.0-alpha.17",
		},
		{
			tag: "2.0.0-beta.18-2",
			exp: "2.0.0-beta.18",
		},
		{
			exp: "dev-nightly",
		},
		{
			// nightly build with env var tag
			tag: "2.0.0",
			env: map[string]string{
				"TRAVIS":              "true",
				"TRAVIS_BUILD_NUMBER": "42",
			},
			exp: "2.0.0",
		},
		{
			tag: "2.0.0",
			env: map[string]string{
				"APPVEYOR":              "true",
				"APPVEYOR_BUILD_NUMBER": "666",
			},
			exp: "2.0.0",
		},
		{
			tag: "2.0.0",
			env: map[string]string{
				"SENSU_BUILD_ITERATION": "7",
			},
			exp: "2.0.0",
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ver, err := GetVersion(test.tag)
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
			err: true,
		},
		{
			tag: "2.0.0-alpha.17-1",
			exp: "2.0.0-alpha.17-1",
		},
		{
			tag: "2.0.0-beta.18-2",
			exp: "2.0.0-beta.18-2",
		},
		{
			err: true,
		},
		{
			// nightly build with env var tag
			env: map[string]string{
				"TRAVIS":              "true",
				"TRAVIS_BUILD_NUMBER": "42",
			},
			exp: "dev-nightly-42",
		},
		{
			env: map[string]string{
				"APPVEYOR":              "true",
				"APPVEYOR_BUILD_NUMBER": "666",
			},
			exp: "dev-nightly-666",
		},
		{
			env: map[string]string{
				"SENSU_BUILD_ITERATION": "7",
			},
			exp: "dev-nightly-7",
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
			ver, err := FullVersion(test.tag)
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
