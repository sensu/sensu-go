package globals

import (
	"fmt"
	"strings"

	corev2 "github.com/sensu/core/v2"
)

// FormatHookLists formats the Check Hook struct into a string mapping
func FormatHookLists(hookLists []corev2.HookList) string {
	hooksString := []string{}
	for _, hookList := range hookLists {
		hookString := fmt.Sprintf("%s: [", hookList.Type)
		hookString += fmt.Sprintf("%s]", strings.Join(hookList.Hooks, ", "))
		hooksString = append(hooksString, hookString)
	}
	return strings.Join(hooksString, ", ")
}
