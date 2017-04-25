package dashboardd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

const (
	// API represents the Sensu API url
	API = "http://127.0.0.1:8080"
)

// Config represents the dashboard configuration
type Config struct {
	Dir  string
	Host string
	Port int
}

// Dashboardd represents the dashboard daemon
type Dashboardd struct {
	stopping      chan struct{}
	running       *atomic.Value
	wg            *sync.WaitGroup
	errChan       chan error
	BackendStatus func() types.StatusMap

	Config
}

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "dashboard",
	})
}

// Start dashboardd
func (d *Dashboardd) Start() error {
	d.stopping = make(chan struct{}, 1)
	d.running = &atomic.Value{}
	d.wg = &sync.WaitGroup{}

	d.errChan = make(chan error, 1)

	router := httpRouter(d)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", d.Host, d.Port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("starting dashboardd on address: ", server.Addr)

	go func() {
		defer d.wg.Done()
		server.ListenAndServe()
	}()

	return nil
}

// Stop dashboardd.
func (d *Dashboardd) Stop() error {
	close(d.stopping)
	d.wg.Wait()
	close(d.errChan)

	return nil
}

// Status returns an error if dashboardd is unhealthy.
func (d *Dashboardd) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (d *Dashboardd) Err() <-chan error {
	return d.errChan
}

func httpRouter(d *Dashboardd) *mux.Router {
	r := mux.NewRouter()

	// API gateway to Sensu API
	target, err := url.Parse(API)
	if err != nil {
		logger.Fatal(err)
	}

	r.Handle("/events", httputil.NewSingleHostReverseProxy(target))
	r.Handle("/entities", httputil.NewSingleHostReverseProxy(target))

	// Serve static content
	r.PathPrefix("/").Handler(noCacheHandler(http.FileServer(http.Dir(d.Dir))))

	return r
}

// noCacheHandler sets the proper headers to prevent any sort of caching for the
// index.html file, served as /
func noCacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")
			w.Header().Set("pragma", "no-cache")
			w.Header().Set("expires", "0")
		}
		next.ServeHTTP(w, r)
	})
}
