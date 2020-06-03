package token

import (
	"fmt"
	"text/template"

	"github.com/sensu/sensu-go/util/environment"
)

// funcMap defines the available custom functions in templates
func funcMap() template.FuncMap {
	return template.FuncMap{
		"default":   defaultFunc,
		"assetPath": assetPath,
	}
}

// defaultFunc receives v, a slice of interfaces, which length range between one
// and two arguments, depending on whether the token has a corresponding field.
// The first argument always represents the default value, while the optional
// second argument represent the value of the token if it was properly
// substitued, in which case we should return that value instead of the default
func defaultFunc(v ...interface{}) interface{} {
	if len(v) == 1 {
		return v[0]
	} else if len(v) == 2 {
		if v[1] == nil {
			return v[0]
		}
		return v[1]
	}
	return nil
}

func assetPath(name string) string {
	return fmt.Sprintf("%s_PATH", environment.Key(name))
}
