package routers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
)

const defaultTimeout = 3

// ClusterController represents the controller needs of the ClusterRouter.
type ClusterController interface {
	// MemberList lists the current cluster membership.
	MemberList(ctx context.Context) (*clientv3.MemberListResponse, error)

	// MemberAdd adds a new member into the cluster.
	MemberAdd(ctx context.Context, peerAddrs []string) (*clientv3.MemberAddResponse, error)

	// MemberRemove removes an existing member from the cluster.
	MemberRemove(ctx context.Context, id uint64) (*clientv3.MemberRemoveResponse, error)

	// MemberUpdate updates the peer addresses of the member.
	MemberUpdate(ctx context.Context, id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error)

	// ClusterID gets the sensu cluster id.
	ClusterID(ctx context.Context) (string, error)
}

// ClusterRouter handles requests for /cluster
type ClusterRouter struct {
	controller ClusterController
}

// NewClusterRouter creates a new ClusterRouter.
func NewClusterRouter(ctrl ClusterController) *ClusterRouter {
	return &ClusterRouter{
		controller: ctrl,
	}
}

// Mount mounts the ClusterRouter to a parent Router.
func (r *ClusterRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/cluster/members", r.list).Methods(http.MethodGet)
	parent.HandleFunc("/cluster/members", r.memberAdd).Methods(http.MethodPost)
	parent.HandleFunc("/cluster/members/{id}", r.memberRemove).Methods(http.MethodDelete)
	parent.HandleFunc("/cluster/members/{id}", r.memberUpdate).Methods(http.MethodPut)
	parent.HandleFunc("/cluster/id", r.clusterID).Methods(http.MethodGet)
}

func parseID(req *http.Request) (uint64, error) {
	paramID := mux.Vars(req)["id"]
	id, err := strconv.ParseUint(paramID, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("bad id (%s): %s", paramID, err)
	}
	return id, nil
}

func parsePeerAddrs(req *http.Request) ([]string, error) {
	peerAddrsValue := req.FormValue("peer-addrs")
	if len(peerAddrsValue) == 0 {
		return nil, errors.New("missing peer-addrs form value")
	}
	return strings.Split(peerAddrsValue, ","), nil
}

func parseTimeout(req *http.Request) (int, error) {
	val := req.FormValue("timeout")
	if len(val) == 0 {
		return defaultTimeout, nil
	}
	return strconv.Atoi(val)
}

func (r *ClusterRouter) list(w http.ResponseWriter, req *http.Request) {
	timeout, err := parseTimeout(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if timeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		ctx = tctx
	}
	resp, err := r.controller.MemberList(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (r *ClusterRouter) memberAdd(w http.ResponseWriter, req *http.Request) {
	timeout, err := parseTimeout(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if timeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		ctx = tctx
	}
	peerAddrs, err := parsePeerAddrs(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := r.controller.MemberAdd(ctx, peerAddrs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (r *ClusterRouter) memberRemove(w http.ResponseWriter, req *http.Request) {
	timeout, err := parseTimeout(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if timeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		ctx = tctx
	}
	id, err := parseID(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := r.controller.MemberRemove(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (r *ClusterRouter) memberUpdate(w http.ResponseWriter, req *http.Request) {
	timeout, err := parseTimeout(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if timeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		ctx = tctx
	}
	id, err := parseID(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	peerAddrs, err := parsePeerAddrs(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := r.controller.MemberUpdate(ctx, id, peerAddrs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (r *ClusterRouter) clusterID(w http.ResponseWriter, req *http.Request) {
	timeout, err := parseTimeout(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if timeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		ctx = tctx
	}
	resp, err := r.controller.ClusterID(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}
