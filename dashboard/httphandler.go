package dashboard

import "net/http"

// HTTPHandler returns new handler returning compiled dashboard assets.
var HTTPHandler http.Handler =
// On builds without the release tag we simply return a empty handler.
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
