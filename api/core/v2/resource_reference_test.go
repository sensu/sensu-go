package v2

import testing "testing"

func TestStringRef(t *testing.T) {
	testCases := []struct {
		Input           string
		ExpectError     bool
		ExpectedVersion string
		ExpectedType    string
		ExpectedName    string
	}{
		{
			Input:           "core/v2.Pipeline.testcase",
			ExpectedVersion: "core/v2",
			ExpectedType:    "Pipeline",
			ExpectedName:    "testcase",
		}, {
			Input:           "core/v2.Pipeline  testcase",
			ExpectedVersion: "core/v2",
			ExpectedType:    "Pipeline",
			ExpectedName:    "testcase",
		}, {
			Input:           "core/v2000.X  testcase",
			ExpectedVersion: "core/v2000",
			ExpectedType:    "X",
			ExpectedName:    "testcase",
		}, {
			Input:           "core/v2.X  test-case_123:A.b.C",
			ExpectedVersion: "core/v2",
			ExpectedType:    "X",
			ExpectedName:    "test-case_123:A.b.C",
		}, {
			Input:           "api_version/v0.type.testcase",
			ExpectedVersion: "api_version/v0",
			ExpectedType:    "type",
			ExpectedName:    "testcase",
		}, {
			Input:       "api_version.type.testcase",
			ExpectError: true,
		}, {
			Input:       "a/v1..",
			ExpectError: true,
		}, {
			Input:       "foo",
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		actual, err := FromStringRef(tc.Input)
		if err != nil {
			if !tc.ExpectError {
				t.Errorf("unexpected error for input %v: %v", tc.Input, err)
			}
			continue
		} else if tc.ExpectError {
			t.Errorf("expected error for input %v", tc.Input)
			continue
		}

		if actual.GetAPIVersion() != tc.ExpectedVersion {
			t.Errorf("expected %v, got %v", tc.ExpectedVersion, actual.GetAPIVersion())
		}
		if actual.GetType() != tc.ExpectedType {
			t.Errorf("expected %v, got %v", tc.ExpectedType, actual.GetType())
		}
		if actual.GetName() != tc.ExpectedName {
			t.Errorf("expected %v, got %v", tc.ExpectedName, actual.GetName())
		}
	}
}
