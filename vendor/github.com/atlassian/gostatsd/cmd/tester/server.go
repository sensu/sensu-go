package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// Server encapsulates all of the parameters necessary for starting up
// the server. These can either be set via command line or directly.
type Server struct {
	Concurrency   int
	MaxPacketSize int

	FlushInterval time.Duration

	MetricsAddr string
	Namespace   string
	WebAddr     string

	start chan bool
	stop  chan bool
	stats chan Stats

	Stats Stats

	Benchmark  int
	Started    int32
	Load       bool
	Verbose    bool
	Version    bool
	CPUProfile bool
}

// Stats reprensents the stats for the session.
type Stats struct {
	Duration         string    `json:"duration,omitempty"`
	MetricsPerSecond float64   `json:"metricsPerSecond,omitempty"`
	NumPackets       int64     `json:"numPackets,omitempty"`
	NumMetrics       int64     `json:"numMetrics,omitempty"`
	PacketsPerSecond float64   `json:"packetsPerSecond,omitempty"`
	StartTime        time.Time `json:"startTime,omitempty"`
	StopTime         time.Time `json:"stopTime,omitempty"`
}

// NewServer will create a new Server with default values.
func newServer() *Server {
	return &Server{
		Concurrency:   10,
		FlushInterval: 1 * time.Second,
		MaxPacketSize: 1400,
		MetricsAddr:   ":8125",
		WebAddr:       ":8080",
	}
}

// AddFlags adds flags for a specific Server to the specified FlagSet.
func (s *Server) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&s.Concurrency, "concurrency", s.Concurrency, "How much concurrency for load testing")
	fs.IntVar(&s.MaxPacketSize, "max-packet-size", s.MaxPacketSize, "Max size of the packet sent to statsd")
	fs.DurationVar(&s.FlushInterval, "flush-interval", s.FlushInterval, "How often to flush metrics to the backends")
	fs.BoolVar(&s.Load, "load", false, "Trigger load testing")
	fs.IntVar(&s.Benchmark, "benchmark", 0, "Time in seconds to run benchmark (0 - disabled)")
	fs.BoolVar(&s.CPUProfile, "cpu-profile", false, "Enable CPU profiler for benchmark")
	fs.StringVar(&s.MetricsAddr, "metrics-addr", s.MetricsAddr, "Address on which to send metrics")
	fs.StringVar(&s.Namespace, "namespace", s.Namespace, "Namespace all metrics")
	fs.BoolVar(&s.Verbose, "verbose", false, "Verbose")
	fs.BoolVar(&s.Version, "version", false, "Print the version and exit")
	fs.StringVar(&s.WebAddr, "web-addr", s.WebAddr, "Address on which to listen for request")
}

// Run runs the specified Server.
func (s *Server) Run() error {
	if s.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	if s.Namespace == "" {
		s.Namespace = os.Getenv("MICROS_ENV")
	}
	if s.Namespace != "" {
		s.Namespace += "."
	}

	if s.Concurrency <= 0 || s.Concurrency > 65536 {
		return fmt.Errorf("concurrency needs to be an integer between 1 and 65,536")
	}

	s.start = make(chan bool)
	s.stop = make(chan bool)
	s.stats = make(chan Stats)

	if s.Load {
		go s.load()
	} else {
		go s.send()
	}

	http.HandleFunc("/heartbeat", s.heartbeatHandler)
	http.HandleFunc("/start", s.startHandler)
	http.HandleFunc("/stop", s.stopHandler)
	http.HandleFunc("/exit", s.exitHandler)
	http.ListenAndServe(s.WebAddr, nil)

	return nil
}

func (s *Server) heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("starting process\n")
	s.start <- true
	w.Write([]byte("process started"))
}

func (s *Server) stopHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&s.Started) != 0 {
		log.Info("stopping process")
		s.stop <- true
		stats := <-s.stats
		w.Header().Set("Content-Type", "application/json")
		bytes, err := json.Marshal(stats)
		if err != nil {
			log.Errorf("unable to marshal TimeSeries, %s\n", err.Error())
			w.Write([]byte("process stopped"))
			return
		}
		w.Write(bytes)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("process not started"))
	}

}

func (s *Server) exitHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("exiting process\n")
	s.stop <- true
	os.Exit(0)
}

func (s *Server) gatherStats() {
	s.Stats.StopTime = time.Now()
	duration := s.Stats.StopTime.Sub(s.Stats.StartTime)
	s.Stats.Duration = duration.String()
	s.Stats.MetricsPerSecond = float64(atomic.LoadInt64(&s.Stats.NumMetrics)) / duration.Seconds()
	s.Stats.PacketsPerSecond = float64(atomic.LoadInt64(&s.Stats.NumPackets)) / duration.Seconds()
	s.stats <- s.Stats
}
