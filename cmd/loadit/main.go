package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"

	"net/http"
	_ "net/http/pprof"
)

var (
	flagCount             = flag.Int("count", 1000, "number of concurrent simulated agents")
	flagBackends          = flag.String("backends", "ws://localhost:8081", "comma separated list of backend URLs")
	flagSubscriptions     = flag.String("subscriptions", "default", "comma separated list of subscriptions")
	flagKeepaliveInterval = flag.Int("keepalive-interval", agent.DefaultKeepaliveInterval, "Keepalive interval")
	flagKeepaliveTimeout  = flag.Int("keepalive-timeout", types.DefaultKeepaliveTimeout, "Keepalive timeout")
	flagProfilingPort     = flag.Int("pprof-port", 6060, "pprof port to bind to")
)

func main() {
	flag.Parse()

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
		cfg.KeepaliveWarningTimeout = uint32(*flagKeepaliveTimeout)
		cfg.Namespace = agent.DefaultNamespace
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

		agent, err := agent.NewAgent(cfg)
		if err != nil {
			log.Fatal(err)
		}

		agent.ProcessGetter = &procGetter{}
		if err := agent.RefreshSystemInfo(ctx); err != nil {
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

type procGetter struct{}

func (p *procGetter) Get(ctx context.Context) ([]*corev2.Process, error) {
	processList := []*corev2.Process{{Name: "systemd"}, {Name: "kthreadd"}, {Name: "ksoftirqd/0"}, {Name: "kworker/0:0H"}, {Name: "kworker/u12:0"}, {Name: "migration/0"}, {Name: "rcu_bh"}, {Name: "rcu_sched"}, {Name: "lru-add-drain"}, {Name: "watchdog/0"}, {Name: "watchdog/1"}, {Name: "migration/1"}, {Name: "ksoftirqd/1"}, {Name: "kworker/1:0H"}, {Name: "watchdog/2"}, {Name: "migration/2"}, {Name: "ksoftirqd/2"}, {Name: "kworker/2:0H"}, {Name: "watchdog/3"}, {Name: "migration/3"}, {Name: "ksoftirqd/3"}, {Name: "kworker/3:0"}, {Name: "kworker/3:0H"}, {Name: "watchdog/4"}, {Name: "migration/4"}, {Name: "ksoftirqd/4"}, {Name: "kworker/4:0"}, {Name: "kworker/4:0H"}, {Name: "watchdog/5"}, {Name: "migration/5"}, {Name: "ksoftirqd/5"}, {Name: "kworker/5:0"}, {Name: "kworker/5:0H"}, {Name: "kdevtmpfs"}, {Name: "netns"}, {Name: "khungtaskd"}, {Name: "writeback"}, {Name: "kintegrityd"}, {Name: "bioset"}, {Name: "bioset"}, {Name: "bioset"}, {Name: "kblockd"}, {Name: "md"}, {Name: "edac-poller"}, {Name: "watchdogd"}, {Name: "kworker/u12:1"}, {Name: "kswapd0"}, {Name: "ksmd"}, {Name: "khugepaged"}, {Name: "crypto"}, {Name: "kthrotld"}, {Name: "kmpath_rdacd"}, {Name: "kaluad"}, {Name: "kworker/1:1"}, {Name: "kpsmoused"}, {Name: "ipv6_addrconf"}, {Name: "deferwq"}, {Name: "kworker/4:1"}, {Name: "kauditd"}, {Name: "kworker/2:1"}, {Name: "ata_sff"}, {Name: "scsi_eh_0"}, {Name: "scsi_tmf_0"}, {Name: "scsi_eh_1"}, {Name: "scsi_tmf_1"}, {Name: "bioset"}, {Name: "xfsalloc"}, {Name: "xfs_mru_cache"}, {Name: "xfs-buf/sda1"}, {Name: "xfs-data/sda1"}, {Name: "xfs-conv/sda1"}, {Name: "xfs-cil/sda1"}, {Name: "xfs-reclaim/sda"}, {Name: "xfs-log/sda1"}, {Name: "xfs-eofblocks/s"}, {Name: "xfsaild/sda1"}, {Name: "kworker/0:1H"}, {Name: "kworker/5:1H"}, {Name: "kworker/2:2"}, {Name: "kworker/1:2"}, {Name: "systemd-journald"}, {Name: "kworker/1:1H"}, {Name: "systemd-udevd"}, {Name: "kworker/3:1H"}, {Name: "rpciod"}, {Name: "xprtiod"}, {Name: "auditd"}, {Name: "kworker/5:2"}, {Name: "dbus-daemon"}, {Name: "rpcbind"}, {Name: "polkitd"}, {Name: "irqbalance"}, {Name: "systemd-logind"}, {Name: "chronyd"}, {Name: "gssproxy"}, {Name: "kworker/2:1H"}, {Name: "agetty"}, {Name: "tuned"}, {Name: "sshd"}, {Name: "rsyslogd"}, {Name: "kworker/3:2"}, {Name: "master"}, {Name: "pickup"}, {Name: "qmgr"}, {Name: "kworker/4:1H"}, {Name: "crond"}, {Name: "kworker/0:3"}, {Name: "ttm_swap"}, {Name: "iprt-VBoxWQueue"}, {Name: "VBoxService"}, {Name: "NetworkManager"}, {Name: "dhclient"}, {Name: "dhclient"}, {Name: "sshd"}, {Name: "sshd"}, {Name: "bash"}, {Name: "sudo"}, {Name: "su"}, {Name: "bash"}, {Name: "kworker/0:1"}, {Name: "kworker/0:0"}, {Name: "sensu-agent"}}
	return processList, nil
}
