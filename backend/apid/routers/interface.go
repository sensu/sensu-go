package routers

import "github.com/gorilla/mux"

// Router mounts new subrouters on parent routers
type Router interface {
	Mount(*mux.Router)
}
