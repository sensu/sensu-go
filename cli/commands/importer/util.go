package importer

import "fmt"

func unsupportedAttr(resource, attr string) string {
	return fmt.Sprintf(
		"%s with '%s' attribute are not supported at this time",
		resource,
		attr,
	)
}
