package dashboardd

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

type Config struct {
	Dir  string
	Host string
	Port int
}

type Dashboardd struct {
	stopping      chan struct{}
	running       *atomic.Value
	wg            *sync.WaitGroup
	errChan       chan error
	BackendStatus func() types.StatusMap

	Config
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

	log.Println("starting dashboardd on address: ", server.Addr)

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
	// eventChan is closed by MessageBus Stop()

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

	// Serve static content
	r.Handle("/", noCacheHandler(http.FileServer(http.Dir(d.Dir))))

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
