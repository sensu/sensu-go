package dashboardd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/dashboardd/asset"
	"github.com/sensu/sensu-go/dashboard"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// Config represents the dashboard configuration
type Config struct {
	Host string
	Port int
	TLS  *types.TLSOptions

	APIURL string
}

// Dashboardd represents the dashboard daemon
type Dashboardd struct {
	stopping   chan struct{}
	running    *atomic.Value
	wg         *sync.WaitGroup
	errChan    chan error
	httpServer *http.Server
	logger     *logrus.Entry

	Config
	Assets *asset.Collection
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
		logger: logrus.WithField(
			"component", "dashboard",
		),
	}

	// load assets
	collection := asset.NewCollection()
	collection.Extend(dashboard.Assets)
	d.Assets = collection

	// prepare server TLS config
	tlsServerConfig, err := cfg.TLS.ToServerTLSConfig()
	if err != nil {
		return nil, err
	}

	d.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", d.Host, d.Port),
		Handler:      httpRouter(d),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		TLSConfig:    tlsServerConfig,
	}
	for _, o := range opts {
		if err := o(d); err != nil {
			return nil, err
		}
	}

	return d, nil
}

// Start dashboardd
func (d *Dashboardd) Start() error {
	d.logger.Info("starting dashboardd on address: ", d.httpServer.Addr)
	d.wg.Add(1)

	go func() {
		defer d.wg.Done()
		var err error
		TLS := d.Config.TLS
		if TLS != nil {
			// TLS configuration comes from ToServerTLSConfig
			err = d.httpServer.ListenAndServeTLS("", "")
		} else {
			err = d.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			d.errChan <- fmt.Errorf("failed to start http/https server %s", err)
		}
	}()

	return nil
}

// Stop dashboardd.
func (d *Dashboardd) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := d.httpServer.Shutdown(ctx)

	close(d.stopping)
	d.wg.Wait()
	close(d.errChan)

	return err
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

	backendProxy, err := newBackendProxy(d.Config.APIURL, d.Config.TLS)
	if err != nil {
		d.errChan <- err
	}

	// Gzip content
	gziphandler, err := gziphandler.NewGzipLevelAndMinSize(
		gzip.DefaultCompression,
		gziphandler.DefaultMinSize,
	)
	if err != nil {
		panic(err)
	}

	// Proxy endpoints
	r.PathPrefix("/auth").Handler(backendProxy)
	r.PathPrefix("/graphql").Handler(gziphandler(backendProxy))

	// Expose Asset Info
	r.PathPrefix("/asset-info.json").Handler(gziphandler(listAssetsHandler(d.Assets, d.logger)))

	// Serve assets
	r.PathPrefix("/static").Handler(staticHandler(d.Assets))
	r.PathPrefix("/").Handler(rootHandler(d.Assets))

	return r
}

func staticHandler(fs http.FileSystem) http.Handler {
	handler := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set cache-control header to allow client to cache file indefinitely.
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#Freshness
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Revalidation_and_reloading
		w.Header().Set("cache-control", "max-age=31536000, immutable")
		handler.ServeHTTP(w, r)
	})
}

func rootHandler(fs http.FileSystem) http.Handler {
	handler := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fallback to index if path didn't match an asset
		if f, _ := fs.Open(r.URL.Path); f == nil {
			r.URL.Path = "/"
		}

		// Sets the proper headers to prevent any sort of caching for the non-static
		// files such as the index and manifest.
		w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")
		w.Header().Set("pragma", "no-cache")
		w.Header().Set("expires", "0")
		handler.ServeHTTP(w, r)
	})
}

func newBackendProxy(aurl string, tlsOpts *types.TLSOptions) (*httputil.ReverseProxy, error) {
	// API gateway to Sensu API
	target, err := url.Parse(aurl)
	if err != nil {
		return nil, err
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
	if tlsOpts != nil {
		cfg, err := tlsOpts.ToServerTLSConfig()
		if err != nil {
			return nil, err
		}
		// TODO(palourde): We should avoid using the loopback interface
		cfg.InsecureSkipVerify = true
		transport.TLSClientConfig = cfg
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	return proxy, nil
}

type Path struct {
	Name    string
	IsDir   bool
	ModTime time.Time
	Size    int64
}

func listAssetsHandler(fs http.FileSystem, logger *logrus.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Derive contents of assets collection
		paths, err := asset.ListContents(fs, "/")
		if err != nil {
			logger.WithError(err).Error("unable to list assets")
			writeError(w)
			return
		}

		// Sanatize
		out := []Path{}
		for _, path := range paths {
			out = append(out, Path{
				Name:    path.Name,
				IsDir:   path.IsDir(),
				ModTime: path.ModTime(),
				Size:    path.Size(),
			})
		}

		// Write
		if err = writeJSON(w, out); err != nil {
			logger.WithError(err).Error("unable to write assets")
			writeError(w)
		}
	})
}

func writeJSON(w http.ResponseWriter, d interface{}) error {
	// Headers
	w.Header().Set("Content-Type", "application/json")

	// Marshal
	bytes, err := json.Marshal(d)
	if err != nil {
		return err
	}

	// Write
	_, err = w.Write(bytes)
	return err
}

func writeError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(500)
	_, _ = w.Write([]byte("{}"))
}
