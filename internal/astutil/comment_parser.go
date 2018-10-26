package astutil

import (
	"bufio"
	"go/doc"
	"strings"
)

// ParseCommentTags parses comment tags like this:
// +app:key=var -> map[string]string{"+app:key": "var"}
func ParseCommentTags(kindName string, pkg *doc.Package) map[string]string {
	var dtype *doc.Type
	for _, t := range pkg.Types {
		if t.Name == kindName {
			dtype = t
		}
	}
	result := make(map[string]string)
	if dtype == nil {
		return result
	}
	scanner := bufio.NewScanner(strings.NewReader(dtype.Doc))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		token := scanner.Text()
		if strings.HasPrefix(token, "+") && scanner.Scan() {
			result[token] = scanner.Text()
		}
	}
	return result
}
