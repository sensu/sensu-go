package middlewares

import (
	"fmt"
	"net/http"
)

func testHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Success")
	})
}
