package middlewares

import (
	"net/http"
)

// Apply applies given middleware left to right returning new http.Handler.
//
//   // without apply
//   my_router := mux.Router{}
//   my_stack := Auth{}.Then(my_router)
//   my_stack = Logger{}.Then(my_stack)
//   my_stack = Instrumentation{}.Then(my_stack)
//
//   // with apply
//   my_router := mux.Router{}
//   my_stack := Apply(my_router, Instrumentation{}, Logger{}, Authentication{})
func Apply(handler http.Handler, ms ...HTTPMiddleware) http.Handler {
	var m HTTPMiddleware

	for len(ms) > 0 {
		m, ms = ms[len(ms)-1], ms[:len(ms)-1]
		handler = m.Then(handler)
	}

	return handler
}
