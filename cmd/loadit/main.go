package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

var (
	flagCount             = flag.Int("count", 1000, "number of concurrent simulated agents")
	flagBackends          = flag.String("backends", "ws://localhost:8081", "comma separated list of backend URLs")
	flagSubscriptions     = flag.String("subscriptions", "default", "comma separated list of subscriptions")
	flagKeepaliveInterval = flag.Int("keepalive-interval", agent.DefaultKeepaliveInterval, "Keepalive interval")
	flagKeepaliveTimeout  = flag.Int("keepalive-timeout", types.DefaultKeepaliveTimeout, "Keepalive timeout")
)

func main() {
	flag.Parse()

	logrus.SetLevel(logrus.WarnLevel)

	subscriptions := strings.Split(*flagSubscriptions, ",")
	backends := strings.Split(*flagBackends, ",")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < *flagCount; i++ {
		name := uuid.New().String()

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
		cfg.KeepaliveTimeout = uint32(*flagKeepaliveTimeout)
		cfg.Namespace = agent.DefaultNamespace
		cfg.Password = agent.DefaultPassword
		cfg.Socket.Host = agent.DefaultAPIHost
		cfg.Socket.Port = agent.DefaultAPIPort
		cfg.User = agent.DefaultUser
		cfg.Subscriptions = subscriptions
		cfg.AgentName = name
		cfg.BackendURLs = backends
		cfg.BackendHandshakeTimeout = 60
		cfg.BackendHeartbeatInterval = 30
		cfg.BackendHeartbeatTimeout = 120
		cfg.TLS = &types.TLSOptions{}

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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}
