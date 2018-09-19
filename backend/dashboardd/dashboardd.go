package dashboardd

import (
	"compress/gzip"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/dashboard"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

const (
	// APIHost represents the Sensu API's host
	APIHost = "127.0.0.1"
)

// Config represents the dashboard configuration
type Config struct {
	Host string
	Port int
	TLS  *types.TLSOptions

	APIPort int
}

// Dashboardd represents the dashboard daemon
type Dashboardd struct {
	stopping   chan struct{}
	running    *atomic.Value
	wg         *sync.WaitGroup
	errChan    chan error
	httpServer *http.Server

	Config
}

// Option is a functional option.
type Option func(*Dashboardd) error

// New creates a new Dashboardd.
func New(cfg Config, opts ...Option) (*Dashboardd, error) {
	d := &Dashboardd{
		Config:   cfg,
		stopping: make(chan struct{}, 1),
		running:  &atomic.Value{},
		wg:       &sync.WaitGroup{},
		errChan:  make(chan error, 1),
	}
	d.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", d.Host, d.Port),
		Handler:      httpRouter(d),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	for _, o := range opts {
		if err := o(d); err != nil {
			return nil, err
		}
	}

	return d, nil
}

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "dashboard",
	})
}

// Start dashboardd
func (d *Dashboardd) Start() error {
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
		if err != nil && err != http.ErrServerClosed {
			d.errChan <- fmt.Errorf("failed to start http/https server %s", err.Error())
		}
	}()

	return nil
}

// Stop dashboardd.
func (d *Dashboardd) Stop() error {
	if err := d.httpServer.Shutdown(nil); err != nil {
		// failure/timeout shutting down the server gracefully
		logger.WithError(err).Error("failed to shutdown http server gracefully - forcing shutdown")
		if closeErr := d.httpServer.Close(); closeErr != nil {
			logger.WithError(closeErr).Error("failed to shutdown http server forcefully")
		}
	}

	close(d.stopping)
	d.wg.Wait()
	close(d.errChan)

	return nil
}

// Err returns a channel to listen for terminal errors on.
func (d *Dashboardd) Err() <-chan error {
	return d.errChan
}

// Name returns the daemon name
func (d *Dashboardd) Name() string {
	return "dashboardd"
}

func httpRouter(d *Dashboardd) *mux.Router {
	r := mux.NewRouter()

	backendProxy, err := newBackendProxy(d.Config.APIPort, d.Config.TLS)
	if err != nil {
		d.errChan <- err
	}

	// Proxy endpoints
	r.PathPrefix("/auth").Handler(backendProxy)
	r.PathPrefix("/graphql").Handler(backendProxy)

	// Serve assets
	r.PathPrefix("/").Handler(assetsHandler())

	return r
}

func assetsHandler() http.Handler {
	fs := dashboard.Assets
	handler := http.FileServer(fs)

	// Gzip content
	gziphandler, err := gziphandler.NewGzipLevelAndMinSize(
		gzip.DefaultCompression,
		gziphandler.DefaultMinSize,
	)
	if err != nil {
		panic(err)
	}
	handler = gziphandler(handler)

	// Set proper headers
	immutableHandler := immutableHandler(handler)
	noCacheHandler := noCacheHandler(handler)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fallback to index if path didn't match an asset
		if f, _ := fs.Open(r.URL.Path); f == nil {
			r.URL.Path = "/"
		}

		// wrap all static assets in a the immutable handler so that they are not
		// needless revalidated when the client refreshes.
		if strings.HasPrefix(r.URL.Path, "/static") {
			immutableHandler.ServeHTTP(w, r)
		} else {
			noCacheHandler.ServeHTTP(w, r)
		}
	})
}

// immutableHandler sets the proper headers to allow client to cache file
// indefinitely.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#Freshness
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Revalidation_and_reloading
func immutableHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", "max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

// noCacheHandler sets the proper headers to prevent any sort of caching for the
// index.html file, served as /
func noCacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")
		w.Header().Set("pragma", "no-cache")
		w.Header().Set("expires", "0")
		next.ServeHTTP(w, r)
	})
}

func newBackendProxy(port int, TLS *types.TLSOptions) (*httputil.ReverseProxy, error) {
	// API gateway to Sensu API
	target := &url.URL{
		Host:   fmt.Sprintf("%s:%d", APIHost, port),
		Scheme: "http",
	}

	// Copy of values from http.DefaultTransport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Configure TLS
	if TLS != nil {
		target.Scheme = "https"

		cfg, err := TLS.ToTLSConfig()
		if err != nil {
			return nil, err
		}
		cfg.InsecureSkipVerify = true // skip host verification on loopback interface
		transport.TLSClientConfig = cfg
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	return proxy, nil
}
