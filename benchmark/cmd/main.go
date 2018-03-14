package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/coreos/rkt/tests/testutils/logger"
	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/util/path"
)

func main() {
	count := flag.Int("agent-count", 1000, "number of concurrent simulated agents")
	backendHost := flag.String("backend-host", "localhost", "backend hostname")

	flag.Parse()

	var (
		wg     sync.WaitGroup
		agents []*agent.Agent
	)

	agents = make([]*agent.Agent, *count)
	i := 0

	for i < *count {
		id := uuid.New().String()

		cfg := agent.NewConfig()
		cfg.API.Host = agent.DefaultAPIHost
		cfg.API.Port = agent.DefaultAPIPort
		cfg.CacheDir = path.SystemCacheDir("sensu-agent")
		cfg.Deregister = true
		cfg.DeregistrationHandler = ""
		cfg.Environment = agent.DefaultEnvironment
		cfg.ExtendedAttributes = []byte{}
		cfg.KeepaliveInterval = agent.DefaultKeepaliveInterval
		cfg.KeepaliveTimeout = agent.DefaultKeepaliveTimeout
		cfg.Organization = agent.DefaultOrganization
		cfg.Password = agent.DefaultPassword
		cfg.Socket.Host = agent.DefaultAPIHost
		cfg.Socket.Port = agent.DefaultAPIPort
		cfg.User = agent.DefaultUser
		cfg.Subscriptions = []string{"default"}
		cfg.AgentID = id
		cfg.BackendURLs = []string{fmt.Sprintf("ws://%s:%d", *backendHost, 8081)}

		agent := agent.NewAgent(cfg)
		if err := agent.Run(); err != nil {
			logger.Error(err)
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
