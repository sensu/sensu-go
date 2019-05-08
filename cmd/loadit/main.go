package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

var (
	flagCount         = flag.Int("count", 1000, "number of concurrent simulated agents")
	flagBackends      = flag.String("backends", "ws://localhost:8081", "comma separated list of backend URLs")
	flagSubscriptions = flag.String("subscriptions", "default", "comma separated list of subscriptions")
)

func main() {
	flag.Parse()

	logrus.SetLevel(logrus.WarnLevel)

	subscriptions := strings.Split(*flagSubscriptions, ",")
	backends := strings.Split(*flagBackends, ",")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	cacheDir := filepath.Join(homedir, "sensu-benchmark-cache")

	for i := 0; i < *flagCount; i++ {
		name := uuid.New().String()

		cfg := agent.NewConfig()
		cfg.API.Host = agent.DefaultAPIHost
		cfg.API.Port = agent.DefaultAPIPort
		cfg.CacheDir = filepath.Join(cacheDir, fmt.Sprintf("sensu-agent-%d", i))
		cfg.Deregister = true
		cfg.DeregistrationHandler = ""
		cfg.DisableAPI = true
		cfg.DisableSockets = true
		cfg.StatsdServer = &agent.StatsdServerConfig{
			Disable:       true,
			FlushInterval: 10,
		}
		cfg.KeepaliveInterval = agent.DefaultKeepaliveInterval
		cfg.KeepaliveTimeout = types.DefaultKeepaliveTimeout
		cfg.Namespace = agent.DefaultNamespace
		cfg.Password = agent.DefaultPassword
		cfg.Socket.Host = agent.DefaultAPIHost
		cfg.Socket.Port = agent.DefaultAPIPort
		cfg.User = agent.DefaultUser
		cfg.Subscriptions = subscriptions
		cfg.AgentName = name
		cfg.BackendURLs = backends

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
