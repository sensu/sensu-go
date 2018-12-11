package environment

import (
	"os"
	"sort"
	"strings"
)

var pathSeparator = string(os.PathListSeparator)

// MergeEnvironments merges one or more sets of environment variables,
// overwriting any existing variable in the preceding set, except for the
// "special" variables PATH, CPATH and LD_LIBRARY_PATH.
//
// The "special" variables PATH, CPATH and LD_LIBRARY_PATH are merged by
// prepending the value from env2 to the value in env1, effectively giving
// priority to the value from env2.
//
// The expected format for an environment variable definition is VAR=VALUE. Any
// malformed environment variable definition will be discarded by the merge.
func MergeEnvironments(ea []string, es ...[]string) []string {
	envs := toMap(ea)

	for i := range es {
		env := toMap(es[i])
		for k, v := range env {
			switch k {
			case "PATH", "CPATH", "LD_LIBRARY_PATH":
				envs[k] = strings.Join([]string{v, envs[k]}, pathSeparator)
			default:
				envs[k] = v
			}
		}
	}

	return fromMap(envs)
}

func toMap(s []string) map[string]string {
	m := map[string]string{}

	for _, v := range s {
		// Try to split the variable definition into exactly 2 substrings:
		// what's left of the first '=' (the variable name) and what's right
		// of it (the variable value)
		split := strings.SplitN(v, "=", 2)

		switch len(split) {
		case 1:
			if split[0] == v {
				// There is no '=' in the input, consider it malformed
				break
			} else {
				// We came across VAR=, which is equivalent to VAR=""
				m[split[0]] = ""
			}
		case 2:
			// On Windows environment variables can be mixed case, eg. Path or PATH.
			key := strings.ToUpper(split[0])
			// A proper VAR=VALUE definiton
			m[key] = split[1]
		default:
			// Anything else is considered malformed and ignored
			break
		}
	}

	return m
}

func fromMap(m map[string]string) []string {
	s := []string{}

	for k, v := range m {
		s = append(s, k+"="+v)
	}
	sort.StringSlice(s).Sort()

	return s
}
