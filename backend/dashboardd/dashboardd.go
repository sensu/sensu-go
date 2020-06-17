package dashboardd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid"
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

	APIDConfig apid.Config
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

	AuthenticationSubrouter *mux.Router
	CoreSubrouter           *mux.Router
	GraphQLSubrouter        *mux.Router
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
	collection.Extend(dashboard.OSS)
	d.Assets = collection

	// prepare server TLS config
	tlsServerConfig, err := cfg.TLS.ToServerTLSConfig()
	if err != nil {
		return nil, err
	}

	handler, err := httpRouter(cfg.APIDConfig, d)
	if err != nil {
		return nil, err
	}

	d.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", d.Host, d.Port),
		Handler:      handler,
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
	ln, err := net.Listen("tcp", d.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to start dashboardd: %s", err)
	}
	d.wg.Add(1)

	go func() {
		defer d.wg.Done()
		var err error
		TLS := d.Config.TLS
		if TLS != nil {
			// TLS configuration comes from ToServerTLSConfig
			err = d.httpServer.ServeTLS(ln, "", "")
		} else {
			err = d.httpServer.Serve(ln)
		}
		if err != nil && err != http.ErrServerClosed {
			d.errChan <- fmt.Errorf("dashboardd failed while serving: %s", err)
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

func httpRouter(c apid.Config, d *Dashboardd) (*mux.Router, error) {
	r := mux.NewRouter()

	router := apid.NewRouter()
	_ = apid.PublicSubrouter(router, c)
	d.GraphQLSubrouter = apid.GraphQLSubrouter(router, c)
	d.AuthenticationSubrouter = apid.AuthenticationSubrouter(router, c)
	d.CoreSubrouter = apid.CoreSubrouter(router, c)

	// Gzip content
	gziphandler, err := gziphandler.NewGzipLevelAndMinSize(
		gzip.DefaultCompression,
		gziphandler.DefaultMinSize,
	)
	if err != nil {
		return nil, err
	}

	// Don't let API or sensuctl users accidentally connect to the dashboard port
	r.PathPrefix("/").HeadersRegexp("User-Agent", "(curl|sensuctl)").HandlerFunc(handleAPIMistake)

	// Proxy endpoints
	r.PathPrefix("/auth").Handler(router)
	r.PathPrefix("/api").Handler(router)
	r.PathPrefix("/graphql").Handler(gziphandler(router))

	// Expose Asset Info
	r.PathPrefix("/index.json").Handler(gziphandler(listAssetsHandler(d.Assets, d.logger)))

	// Serve assets
	r.PathPrefix("/static").Handler(gziphandler(staticHandler(d.Assets)))
	r.PathPrefix("/").Handler(rootHandler(d.Assets))

	return r, nil
}

// handleAPIMistake sends errors to API clients that mistakenly connect to the
// dashboard port.
func handleAPIMistake(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "Client sent an API request to the web application port!", http.StatusBadRequest)
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

		// Sanitize
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

	// If we receive an err here it is most likely because the client has
	// disconnected and if not a disconnect, there isn't much we can do either.
	// https://github.com/sensu/sensu-go/pull/2786#discussion_r265258086
	_, _ = w.Write(bytes)
	return nil
}

func writeError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(500)
	_, _ = w.Write([]byte("{}"))
}
