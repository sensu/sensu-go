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
	TLS  *types.TLSOptions
}

// Dashboardd represents the dashboard daemon
type Dashboardd struct {
	stopping      chan struct{}
	running       *atomic.Value
	wg            *sync.WaitGroup
	errChan       chan error
	BackendStatus func() types.StatusMap
	httpServer    *http.Server

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

	d.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", d.Host, d.Port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("starting dashboardd on address: ", d.httpServer.Addr)
	d.wg.Add(1)

	go func() {
		defer d.wg.Done()
		var err error
		TLS := d.Config.TLS
		if TLS != nil {
			err = d.httpServer.ListenAndServeTLS(TLS.CertFile, TLS.KeyFile)
		} else {
			err = d.httpServer.ListenAndServe()
		}
		// TODO (JK): need a way to handle closing things like errChan, etc.
		// in cases where there's a failure to start the daemon
		if err != nil && err != http.ErrServerClosed {
			logger.Errorf("failed to start http/https server %s", err.Error())
		}
	}()

	return nil
}

// Stop dashboardd.
func (d *Dashboardd) Stop() error {
	if err := d.httpServer.Shutdown(nil); err != nil {
		// failure/timeout shutting down the server gracefully
		logger.Error("failed to shutdown http server gracefully - forcing shutdown")
		if closeErr := d.httpServer.Close(); closeErr != nil {
			logger.Error("failed to shutdown http server forcefully")
		}
	}

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
	if d.Dir != "" {
		r.PathPrefix("/").Handler(noCacheHandler(http.FileServer(http.Dir(d.Dir))))
	} else {
		r.PathPrefix("/").Handler(noCacheHandler(http.FileServer(HTTP)))
	}

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
