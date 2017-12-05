package globals

import (
	"fmt"
	"strings"

	"github.com/sensu/sensu-go/types"
)

// FormatCheckHooks formats the Check Hook struct into a string mapping
func FormatCheckHooks(checkHooks []types.CheckHook) string {
	hooksString := []string{}
	for _, checkHook := range checkHooks {
		hookString := fmt.Sprintf("%s: [", checkHook.Type)
		hookString += fmt.Sprintf("%s]", strings.Join(checkHook.Hooks, ", "))
		hooksString = append(hooksString, hookString)
	}
	return strings.Join(hooksString, ", ")
}
