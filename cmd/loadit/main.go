package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"

	"net/http"
	_ "net/http/pprof"
)

var (
	flagCount             = flag.Int("count", 1000, "number of concurrent simulated agents")
	flagBackends          = flag.String("backends", "ws://localhost:8081", "comma separated list of backend URLs")
	flagNamespace         = flag.String("namespace", agent.DefaultNamespace, "namespace to use for agents")
	flagSubscriptions     = flag.String("subscriptions", "default", "comma separated list of subscriptions")
	flagKeepaliveInterval = flag.Int("keepalive-interval", agent.DefaultKeepaliveInterval, "Keepalive interval")
	flagKeepaliveTimeout  = flag.Int("keepalive-timeout", types.DefaultKeepaliveTimeout, "Keepalive timeout")
	flagProfilingPort     = flag.Int("pprof-port", 6060, "pprof port to bind to")
	flagPromBinding       = flag.String("prom", ":8080", "binding for prometheus server")
	flagHugeEvents        = flag.Bool("huge-events", false, "send 1 MB events to the backend")
)

func main() {
	flag.Parse()

	if *flagHugeEvents {
		oneMegabyte := make([]byte, 1024*1024)
		for i := range oneMegabyte {
			oneMegabyte[i] = byte(rand.Uint32() % 255)
		}
		oneMegabyteString := base64.StdEncoding.EncodeToString(oneMegabyte)
		command.CannedResponse.Output = oneMegabyteString
	}

	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%d", *flagProfilingPort), nil))
	}()

	logrus.SetLevel(logrus.WarnLevel)

	subscriptions := strings.Split(*flagSubscriptions, ",")
	backends := strings.Split(*flagBackends, ",")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	for i := 0; i < *flagCount; i++ {
		uuidBytes, err := uuid.NewRandom()
		if err != nil {
			log.Fatal(err)
		}
		name := uuidBytes.String()

		cfg := agent.NewConfig()
		cfg.API.Host = agent.DefaultAPIHost
		cfg.API.Port = agent.DefaultAPIPort
		cfg.CacheDir = os.DevNull
		cfg.DisableAssets = true
		cfg.Deregister = true
		cfg.DeregistrationHandler = ""
		cfg.DisableAPI = true
		cfg.DisableSockets = true
		cfg.StatsdServer = &agent.StatsdServerConfig{
			Disable:       true,
			FlushInterval: 10,
		}
		cfg.KeepaliveInterval = uint32(*flagKeepaliveInterval)
		cfg.KeepaliveWarningTimeout = uint32(*flagKeepaliveTimeout)
		cfg.Namespace = *flagNamespace
		cfg.Password = agent.DefaultPassword
		cfg.Socket.Host = agent.DefaultAPIHost
		cfg.Socket.Port = agent.DefaultAPIPort
		cfg.User = agent.DefaultUser
		cfg.Subscriptions = subscriptions
		cfg.AgentName = name
		cfg.BackendURLs = backends
		cfg.MockSystemInfo = true
		cfg.BackendHeartbeatInterval = 30
		cfg.BackendHeartbeatTimeout = 300
		cfg.PrometheusBinding = *flagPromBinding

		agent, err := agent.NewAgent(cfg)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			if err := agent.Run(ctx); err != nil {
				log.Fatal(err)
			}
		}()
	}

	elapsed := time.Since(start)
	fmt.Printf("all agents have been connected in %s\n", elapsed)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}
