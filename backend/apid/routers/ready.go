package routers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type ReadyRouter struct{}

func (r *ReadyRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/ready", r.ready).Methods(http.MethodGet)
}

func (r *ReadyRouter) ready(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ready")
}
