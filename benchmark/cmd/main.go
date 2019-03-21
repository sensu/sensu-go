package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/path"
)

func main() {
	count := flag.Int("agent-count", 1000, "number of concurrent simulated agents")
	backendHost := flag.String("backend-host", "localhost", "backend hostnames, comma separated")

	flag.Parse()

	var (
		wg     sync.WaitGroup
		agents []*agent.Agent
	)

	backends := strings.Split(*backendHost, ",")

	agents = make([]*agent.Agent, *count)
	i := 0

	for i < *count {
		name := uuid.New().String()
		backend := backends[i%len(backends)]

		cfg := agent.NewConfig()
		cfg.API.Host = agent.DefaultAPIHost
		cfg.API.Port = agent.DefaultAPIPort
		cfg.CacheDir = path.SystemCacheDir(fmt.Sprintf("sensu-agent-%s", name))
		cfg.Deregister = true
		cfg.DeregistrationHandler = ""
		cfg.KeepaliveInterval = agent.DefaultKeepaliveInterval
		cfg.KeepaliveTimeout = types.DefaultKeepaliveTimeout
		cfg.Namespace = agent.DefaultNamespace
		cfg.Password = agent.DefaultPassword
		cfg.Socket.Host = agent.DefaultAPIHost
		cfg.Socket.Port = agent.DefaultAPIPort
		cfg.User = agent.DefaultUser
		cfg.Subscriptions = []string{"default"}
		cfg.AgentName = name
		cfg.BackendURLs = []string{fmt.Sprintf("ws://%s:%d", backend, 8081)}

		agent, err := agent.NewAgent(cfg)
		if err != nil {
			log.Fatal(err)
		}
		if err := agent.Run(); err != nil {
			log.Fatal(err)
			continue
		}
		agents[i] = agent
		wg.Add(1)
		i++
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		for _, agent := range agents {
			agent.Stop()
			wg.Done()
		}
	}()

	wg.Wait()
}
