package agent

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var alFixture = []allowList{
	allowList{
		Exec: "my_script.sh",
		Args: []string{"-h foo.com", "-h bar.ca"},
	},
	allowList{
		Exec:      "executable",
		Args:      []string{"-h foo.com", "-h bar.ca"},
		Sha512:    "1two3four5",
		EnableEnv: true,
	},
}

func TestValidateAllowList(t *testing.T) {
	assert := assert.New(t)
	allowList := allowList{
		Exec:      "foo",
		Args:      []string{"bar"},
		Sha512:    "baz",
		EnableEnv: true,
	}

	// Given valid allowList, validation should pass
	assert.NoError(allowList.validate())

	// Given an allowList without an exec, validation should fail
	allowList.Exec = ""
	assert.Error(allowList.validate())

	// Given an allowList without args, validation should fail
	allowList.Exec = "foo"
	allowList.Args = []string{}
	assert.Error(allowList.validate())

	// Given valid allowList, validation should pass
	allowList.Args = []string{""}
	assert.NoError(allowList.validate())
}

func TestAllowListValidJSON(t *testing.T) {
	allowList, err := readAllowList("allow_list.json", func(string) ([]byte, error) {
		return []byte(`
		[
			{
				"exec": "my_script.sh",
				"args": ["-h foo.com", "-h bar.ca"]
			},
			{
				"exec": "executable",
				"args": ["-h foo.com", "-h bar.ca"],
				"sha512": "1two3four5",
				"enable_env": true
			}
		]
		`), nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(allowList))
	require.Equal(t, alFixture, allowList)
}

func TestAllowListValidYAML(t *testing.T) {
	allowList, err := readAllowList("allow_list.yaml", func(string) ([]byte, error) {
		return []byte(`
        - exec: my_script.sh
          args:
          - "-h foo.com"
          - "-h bar.ca"
        - exec: executable
          args:
          - "-h foo.com"
          - "-h bar.ca"
          sha512: 1two3four5
          enable_env: true
        `), nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(allowList))
	require.Equal(t, alFixture, allowList)
}

func TestAllowListValidateError(t *testing.T) {
	allowList, err := readAllowList("allow_list.json", func(string) ([]byte, error) {
		return []byte(`
		[
			{
				"exec": "my_script.sh"
			}
		]
		`), nil
	})
	require.Error(t, err)
	require.Nil(t, allowList)
}

func TestAllowListUnmarshalError(t *testing.T) {
	allowList, err := readAllowList("allow_list.json", func(string) ([]byte, error) {
		return []byte(`
		[
			{
				"exec": "my_script.sh"
		`), nil
	})
	require.Error(t, err)
	require.Nil(t, allowList)
}

func TestAllowListFileError(t *testing.T) {
	allowList, err := readAllowList("allow_list.json", func(string) ([]byte, error) {
		return nil, fmt.Errorf("foo")
	})
	require.Error(t, err)
	require.Nil(t, allowList)
}

func TestAllowListInvalidType(t *testing.T) {
	allowList, err := readAllowList("allow_list.foo", func(string) ([]byte, error) {
		return []byte(`
		[
			{
				"exec": "my_script.sh",
				"args": ["-h foo.com", "-h bar.ca"]
			}
		]
		`), nil
	})
	require.Error(t, err)
	require.Nil(t, allowList)
}

func TestAllowListNoFilePath(t *testing.T) {
	al, err := readAllowList("", func(string) ([]byte, error) {
		return []byte(`
		[
			{
				"exec": "my_script.sh",
				"args": ["-h foo.com", "-h bar.ca"]
			}
		]
		`), nil
	})

	require.NoError(t, err)
	require.Equal(t, 0, len(al))
	require.Equal(t, []allowList(nil), al)
}

func TestMatchAllowList(t *testing.T) {
	testCases := []struct {
		description string
		command     string
		allowList   []allowList
		match       bool
		matched     allowList
	}{
		{
			description: "Match: order of args is different",
			command:     "my_script.sh -h bar.ca -h foo.com",
			allowList: []allowList{
				allowList{
					Exec: "my_script.sh",
					Args: []string{"-h foo.com", "-h bar.ca"},
				},
			},
			match: true,
			matched: allowList{
				Exec: "my_script.sh",
				Args: []string{"-h foo.com", "-h bar.ca"},
			},
		},
		{
			description: "Match: multiple entries, with whitespace",
			command:     "my_script.sh -h bar.ca  -h foo.com  ",
			allowList: []allowList{
				allowList{
					Exec: "my_scripf.sh",
					Args: []string{"-h foo.com"},
				},
				allowList{
					Exec: "my_script.sh",
					Args: []string{"-h foo.com", "-h bar.ca"},
				},
			},
			match: true,
			matched: allowList{
				Exec: "my_script.sh",
				Args: []string{"-h foo.com", "-h bar.ca"},
			},
		},
		{
			description: "Match: Missing arg in check command",
			command:     "my_script.sh -h foo.com",
			allowList: []allowList{
				allowList{
					Exec: "my_script.sh",
					Args: []string{"-h foo.com", "-h bar.ca"},
				},
			},
			match: true,
			matched: allowList{
				Exec: "my_script.sh",
				Args: []string{"-h foo.com", "-h bar.ca"},
			},
		},
		{
			description: "Match: Empty args",
			command:     "my_script.sh",
			allowList: []allowList{
				allowList{
					Exec: "my_script.sh",
					Args: []string{""},
				},
			},
			match: true,
			matched: allowList{
				Exec: "my_script.sh",
				Args: []string{""},
			},
		},
		{
			description: "No match: missing arg in allow list",
			command:     "my_script.sh -h foo.com -h bar.ca",
			allowList: []allowList{
				allowList{
					Exec: "my_script.sh",
					Args: []string{"-h bar.ca"},
				},
			},
			match:   false,
			matched: allowList{},
		},
		{
			description: "No match: Wrong executable",
			command:     "my_script.sh -h bar.ca",
			allowList: []allowList{
				allowList{
					Exec: "my_scripf.sh",
					Args: []string{"-h bar.ca"},
				},
			},
			match:   false,
			matched: allowList{},
		},
		{
			description: "No match: No allow list",
			command:     "foo",
			allowList:   []allowList{},
			match:       false,
			matched:     allowList{},
		},
		{
			description: "No match: No command",
			command:     "",
			allowList: []allowList{
				allowList{
					Exec: "my_scripf.sh",
					Args: []string{"-h bar.ca"},
				},
			},
			match:   false,
			matched: allowList{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			agent := Agent{
				allowList: tc.allowList,
			}
			matched, match := agent.matchAllowList(tc.command)
			assert.Equal(t, tc.match, match)
			assert.Equal(t, tc.matched, matched)
		})
	}
}
