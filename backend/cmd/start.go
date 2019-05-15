package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/path"
	stringsutil "github.com/sensu/sensu-go/util/strings"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// Flag constants
	flagConfigFile            = "config-file"
	flagAgentHost             = "agent-host"
	flagAgentPort             = "agent-port"
	deprecatedFlagAPIHost     = "api-host"
	deprecatedFlagAPIPort     = "api-port"
	flagAPIListenAddress      = "api-listen-address"
	flagAPIURL                = "api-url"
	flagDashboardHost         = "dashboard-host"
	flagDashboardPort         = "dashboard-port"
	flagDashboardCertFile     = "dashboard-cert-file"
	flagDashboardKeyFile      = "dashboard-key-file"
	flagDeregistrationHandler = "deregistration-handler"
	flagCacheDir              = "cache-dir"
	flagStateDir              = "state-dir"
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"
	flagDebug                 = "debug"
	flagLogLevel              = "log-level"

	// Etcd flag constants
	deprecatedFlagEtcdClientURLs               = "listen-client-urls"
	flagEtcdClientURLs                         = "etcd-listen-client-urls"
	deprecatedFlagEtcdPeerURLs                 = "listen-peer-urls"
	flagEtcdPeerURLs                           = "etcd-listen-peer-urls"
	deprecatedFlagEtcdInitialCluster           = "initial-cluster"
	flagEtcdInitialCluster                     = "etcd-initial-cluster"
	deprecatedFlagEtcdInitialAdvertisePeerURLs = "initial-advertise-peer-urls"
	flagEtcdInitialAdvertisePeerURLs           = "etcd-initial-advertise-peer-urls"
	deprecatedFlagEtcdInitialClusterState      = "initial-cluster-state"
	flagEtcdInitialClusterState                = "etcd-initial-cluster-state"
	deprecatedFlagEtcdInitialClusterToken      = "initial-cluster-token"
	flagEtcdInitialClusterToken                = "etcd-initial-cluster-token"
	deprecatedFlagEtcdNodeName                 = "name"
	flagEtcdNodeName                           = "etcd-name"
	flagNoEmbedEtcd                            = "no-embed-etcd"
	flagEtcdAdvertiseClientURLs                = "etcd-advertise-client-urls"

	// Etcd TLS flag constants
	flagEtcdCertFile           = "etcd-cert-file"
	flagEtcdKeyFile            = "etcd-key-file"
	flagEtcdClientCertAuth     = "etcd-client-cert-auth"
	flagEtcdTrustedCAFile      = "etcd-trusted-ca-file"
	flagEtcdPeerCertFile       = "etcd-peer-cert-file"
	flagEtcdPeerKeyFile        = "etcd-peer-key-file"
	flagEtcdPeerClientCertAuth = "etcd-peer-client-cert-auth"
	flagEtcdPeerTrustedCAFile  = "etcd-peer-trusted-ca-file"
	flagEtcdCipherSuites       = "etcd-cipher-suites"

	// Default values

	// defaultEtcdClientURL is the default URL to listen for Etcd clients
	defaultEtcdClientURL = "http://127.0.0.1:2379"
	// defaultEtcdName is the default etcd member node name (single-node cluster
	// only)
	defaultEtcdName = "default"
	// defaultEtcdPeerURL is the default URL to listen for Etcd peers (single-node
	// cluster only)
	defaultEtcdPeerURL = "http://127.0.0.1:2380"

	// defaultEtcdAdvertiseClientURL is the default list of this member's client
	// URLs to advertise to the rest of the cluster
	defaultEtcdAdvertiseClientURL = "http://localhost:2379"

	// Start command usage template
	startUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
	{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
	{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
	{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

General Flags:
{{ $flags := categoryFlags "" .LocalFlags }}{{ $flags.FlagUsages | trimTrailingWhitespaces}}

Store Flags:
{{ $storeFlags := categoryFlags "store" .LocalFlags }}{{ $storeFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
	{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

func init() {
	rootCmd.AddCommand(newStartCommand())
}

func newStartCommand() *cobra.Command {
	var setupErr error

	cmd := &cobra.Command{
		Use:           "start",
		Short:         "start the sensu backend",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())
			if setupErr != nil {
				return setupErr
			}

			// Make sure the deprecated API flags are no longer used
			if host := viper.GetString(deprecatedFlagAPIHost); host != "[::]" {
				logger.Fatalf("Flag --%s has been deprecated, please use --%s instead", deprecatedFlagAPIHost, flagAPIListenAddress)
			}
			if port := viper.GetInt(deprecatedFlagAPIPort); port != 8080 {
				logger.Fatalf("Flag --%s has been deprecated, please use --%s instead", deprecatedFlagAPIPort, flagAPIListenAddress)
			}

			level, err := logrus.ParseLevel(viper.GetString(flagLogLevel))
			if err != nil {
				return err
			}
			logrus.SetLevel(level)

			cfg := &backend.Config{
				AgentHost:             viper.GetString(flagAgentHost),
				AgentPort:             viper.GetInt(flagAgentPort),
				APIListenAddress:      viper.GetString(flagAPIListenAddress),
				APIURL:                viper.GetString(flagAPIURL),
				DashboardHost:         viper.GetString(flagDashboardHost),
				DashboardPort:         viper.GetInt(flagDashboardPort),
				DashboardTLSCertFile:  viper.GetString(flagDashboardCertFile),
				DashboardTLSKeyFile:   viper.GetString(flagDashboardKeyFile),
				DeregistrationHandler: viper.GetString(flagDeregistrationHandler),
				CacheDir:              viper.GetString(flagCacheDir),
				StateDir:              viper.GetString(flagStateDir),

				EtcdAdvertiseClientURLs:      viper.GetStringSlice(flagEtcdAdvertiseClientURLs),
				EtcdListenClientURLs:         viper.GetStringSlice(flagEtcdClientURLs),
				EtcdListenPeerURLs:           viper.GetStringSlice(flagEtcdPeerURLs),
				EtcdInitialCluster:           viper.GetString(flagEtcdInitialCluster),
				EtcdInitialClusterState:      viper.GetString(flagEtcdInitialClusterState),
				EtcdInitialAdvertisePeerURLs: viper.GetStringSlice(flagEtcdInitialAdvertisePeerURLs),
				EtcdInitialClusterToken:      viper.GetString(flagEtcdInitialClusterToken),
				EtcdName:                     viper.GetString(flagEtcdNodeName),
				EtcdCipherSuites:             viper.GetStringSlice(flagEtcdCipherSuites),
				NoEmbedEtcd:                  viper.GetBool(flagNoEmbedEtcd),
			}

			// Sensu APIs TLS config
			certFile := viper.GetString(flagCertFile)
			keyFile := viper.GetString(flagKeyFile)
			insecureSkipTLSVerify := viper.GetBool(flagInsecureSkipTLSVerify)
			// TODO(ccressent gbolo): issue #2548
			// Eventually this should be changed: --insecure-skip-tls-verify --etcd-insecure-skip-tls-verify
			trustedCAFile := viper.GetString(flagTrustedCAFile)

			if certFile != "" && keyFile != "" {
				cfg.TLS = &types.TLSOptions{
					CertFile:           certFile,
					KeyFile:            keyFile,
					TrustedCAFile:      trustedCAFile,
					InsecureSkipVerify: insecureSkipTLSVerify,
				}
			} else if certFile == "" && keyFile != "" {
				return fmt.Errorf("tls configuration error, missing flag: --%s", flagCertFile)
			} else if certFile != "" && keyFile == "" {
				return fmt.Errorf("tls configuration error, missing flag: --%s", flagKeyFile)
			}

			// Etcd TLS config
			cfg.EtcdClientTLSInfo = etcd.TLSInfo{
				CertFile:       viper.GetString(flagEtcdCertFile),
				KeyFile:        viper.GetString(flagEtcdKeyFile),
				TrustedCAFile:  viper.GetString(flagEtcdTrustedCAFile),
				ClientCertAuth: viper.GetBool(flagEtcdClientCertAuth),
			}
			cfg.EtcdPeerTLSInfo = etcd.TLSInfo{
				CertFile:       viper.GetString(flagEtcdPeerCertFile),
				KeyFile:        viper.GetString(flagEtcdPeerKeyFile),
				TrustedCAFile:  viper.GetString(flagEtcdPeerTrustedCAFile),
				ClientCertAuth: viper.GetBool(flagEtcdPeerClientCertAuth),
			}

			sensuBackend, err := initialize(cfg)
			if err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				logger.Info("signal received: ", sig)
				sensuBackend.Stop()
			}()

			if viper.GetBool(flagDebug) {
				go func() {
					log.Println(http.ListenAndServe(":6060", nil))
				}()
			}

			return sensuBackend.Run()
		},
	}

	// Set up distinct flagset for handling config file
	configFlagSet := pflag.NewFlagSet("sensu", pflag.ContinueOnError)
	configFlagSet.StringP(flagConfigFile, "c", "", "path to sensu-backend config file")
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(os.Args[1:])

	// Get the given config file path
	configFile, _ := configFlagSet.GetString(flagConfigFile)
	configFilePath := configFile

	// use the default config path if flagConfigFile was used
	if configFile == "" {
		configFilePath = filepath.Join(path.SystemConfigDir(), "backend.yml")
	}

	// Configure location of backend configuration
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFilePath)

	// Flag defaults
	viper.SetDefault(flagAgentHost, "[::]")
	viper.SetDefault(flagAgentPort, 8081)
	viper.SetDefault(deprecatedFlagAPIHost, "[::]")
	viper.SetDefault(deprecatedFlagAPIPort, 8080)
	viper.SetDefault(flagAPIListenAddress, "[::]:8080")
	viper.SetDefault(flagAPIURL, "http://localhost:8080")
	viper.SetDefault(flagDashboardHost, "[::]")
	viper.SetDefault(flagDashboardPort, 3000)
	viper.SetDefault(flagDashboardCertFile, "")
	viper.SetDefault(flagDashboardKeyFile, "")
	viper.SetDefault(flagDeregistrationHandler, "")
	viper.SetDefault(flagCacheDir, path.SystemCacheDir("sensu-backend"))
	viper.SetDefault(flagStateDir, path.SystemDataDir("sensu-backend"))
	viper.SetDefault(flagCertFile, "")
	viper.SetDefault(flagKeyFile, "")
	viper.SetDefault(flagTrustedCAFile, "")
	viper.SetDefault(flagInsecureSkipTLSVerify, false)
	viper.SetDefault(flagLogLevel, "warn")

	// Etcd defaults
	viper.SetDefault(flagEtcdAdvertiseClientURLs, defaultEtcdAdvertiseClientURL)
	viper.SetDefault(flagEtcdClientURLs, defaultEtcdClientURL)
	viper.SetDefault(flagEtcdPeerURLs, defaultEtcdPeerURL)
	viper.SetDefault(flagEtcdInitialCluster,
		fmt.Sprintf("%s=%s", defaultEtcdName, defaultEtcdPeerURL))
	viper.SetDefault(flagEtcdInitialAdvertisePeerURLs, defaultEtcdPeerURL)
	viper.SetDefault(flagEtcdInitialClusterState, etcd.ClusterStateNew)
	viper.SetDefault(flagEtcdInitialClusterToken, "")
	viper.SetDefault(flagEtcdNodeName, defaultEtcdName)
	viper.SetDefault(flagNoEmbedEtcd, false)

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Main Flags
	cmd.Flags().String(flagAgentHost, viper.GetString(flagAgentHost), "agent listener host")
	cmd.Flags().Int(flagAgentPort, viper.GetInt(flagAgentPort), "agent listener port")
	cmd.Flags().String(flagAPIListenAddress, viper.GetString(flagAPIListenAddress), "address to listen on for api traffic")
	cmd.Flags().String(flagAPIURL, viper.GetString(flagAPIURL), "url of the api to connect to")
	cmd.Flags().String(flagDashboardHost, viper.GetString(flagDashboardHost), "dashboard listener host")
	cmd.Flags().Int(flagDashboardPort, viper.GetInt(flagDashboardPort), "dashboard listener port")
	cmd.Flags().String(flagDashboardCertFile, viper.GetString(flagDashboardCertFile), "dashboard TLS certificate in PEM format")
	cmd.Flags().String(flagDashboardKeyFile, viper.GetString(flagDashboardKeyFile), "dashboard TLS certificate key in PEM format")
	cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "default deregistration handler")
	cmd.Flags().String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
	cmd.Flags().StringP(flagStateDir, "d", viper.GetString(flagStateDir), "path to sensu state storage")
	cmd.Flags().String(flagCertFile, viper.GetString(flagCertFile), "TLS certificate in PEM format")
	cmd.Flags().String(flagKeyFile, viper.GetString(flagKeyFile), "TLS certificate key in PEM format")
	cmd.Flags().String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "TLS CA certificate bundle in PEM format used for etcd client (mutual TLS)")
	cmd.Flags().Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip TLS verification (not recommended!)")
	cmd.Flags().Bool(flagDebug, false, "enable debugging and profiling features")
	cmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")

	// Etcd flags
	cmd.Flags().StringSlice(flagEtcdAdvertiseClientURLs, viper.GetStringSlice(flagEtcdAdvertiseClientURLs), "list of this member's client URLs to advertise to the rest of the cluster.")
	_ = cmd.Flags().SetAnnotation(flagEtcdAdvertiseClientURLs, "categories", []string{"store"})
	cmd.Flags().StringSlice(flagEtcdClientURLs, viper.GetStringSlice(flagEtcdClientURLs), "list of URLs to listen on for client traffic")
	_ = cmd.Flags().SetAnnotation(flagEtcdClientURLs, "categories", []string{"store"})
	cmd.Flags().StringSlice(flagEtcdPeerURLs, viper.GetStringSlice(flagEtcdPeerURLs), "list of URLs to listen on for peer traffic")
	_ = cmd.Flags().SetAnnotation(flagEtcdPeerURLs, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdInitialCluster, viper.GetString(flagEtcdInitialCluster), "initial cluster configuration for bootstrapping")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialCluster, "categories", []string{"store"})
	cmd.Flags().StringSlice(flagEtcdInitialAdvertisePeerURLs, viper.GetStringSlice(flagEtcdInitialAdvertisePeerURLs), "list of this member's peer URLs to advertise to the rest of the cluster")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialAdvertisePeerURLs, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdInitialClusterState, viper.GetString(flagEtcdInitialClusterState), "initial cluster state (\"new\" or \"existing\")")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialClusterState, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdInitialClusterToken, viper.GetString(flagEtcdInitialClusterToken), "initial cluster token for the etcd cluster during bootstrap")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialClusterToken, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdNodeName, viper.GetString(flagEtcdNodeName), "human-readable name for this member")
	_ = cmd.Flags().SetAnnotation(flagEtcdNodeName, "categories", []string{"store"})
	cmd.Flags().Bool(flagNoEmbedEtcd, viper.GetBool(flagNoEmbedEtcd), "don't embed etcd, use external etcd instead")
	_ = cmd.Flags().SetAnnotation(flagNoEmbedEtcd, "categories", []string{"store"})
	cmd.Flags().StringSlice(flagEtcdCipherSuites, nil, "list of ciphers to use for etcd TLS configuration")
	_ = cmd.Flags().SetAnnotation(flagEtcdCipherSuites, "categories", []string{"store"})

	// Etcd TLS flags
	cmd.Flags().String(flagEtcdCertFile, viper.GetString(flagEtcdCertFile), "path to the client server TLS cert file")
	_ = cmd.Flags().SetAnnotation(flagEtcdCertFile, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdKeyFile, viper.GetString(flagEtcdKeyFile), "path to the client server TLS key file")
	_ = cmd.Flags().SetAnnotation(flagEtcdKeyFile, "categories", []string{"store"})
	cmd.Flags().Bool(flagEtcdClientCertAuth, viper.GetBool(flagEtcdClientCertAuth), "enable client cert authentication")
	_ = cmd.Flags().SetAnnotation(flagEtcdClientCertAuth, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdTrustedCAFile, viper.GetString(flagEtcdTrustedCAFile), "path to the client server TLS trusted CA cert file")
	_ = cmd.Flags().SetAnnotation(flagEtcdTrustedCAFile, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdPeerCertFile, viper.GetString(flagEtcdPeerCertFile), "path to the peer server TLS cert file")
	_ = cmd.Flags().SetAnnotation(flagEtcdPeerCertFile, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdPeerKeyFile, viper.GetString(flagEtcdPeerKeyFile), "path to the peer server TLS key file")
	_ = cmd.Flags().SetAnnotation(flagEtcdPeerKeyFile, "categories", []string{"store"})
	cmd.Flags().Bool(flagEtcdPeerClientCertAuth, viper.GetBool(flagEtcdPeerClientCertAuth), "enable peer client cert authentication")
	_ = cmd.Flags().SetAnnotation(flagEtcdPeerClientCertAuth, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdPeerTrustedCAFile, viper.GetString(flagEtcdPeerTrustedCAFile), "path to the peer server TLS trusted CA file")
	_ = cmd.Flags().SetAnnotation(flagEtcdPeerTrustedCAFile, "categories", []string{"store"})

	// Make sure some deprecated flags are no longer used
	cmd.Flags().String(deprecatedFlagAPIHost, viper.GetString(deprecatedFlagAPIHost), "http api listener host")
	cmd.Flags().Int(deprecatedFlagAPIPort, viper.GetInt(deprecatedFlagAPIPort), "http api port")
	_ = cmd.Flags().MarkHidden(deprecatedFlagAPIHost)
	_ = cmd.Flags().MarkHidden(deprecatedFlagAPIPort)

	_ = cmd.Flags().MarkDeprecated(deprecatedFlagAPIHost, fmt.Sprintf("please use --%s instead", flagAPIListenAddress))
	_ = cmd.Flags().MarkDeprecated(deprecatedFlagAPIPort, fmt.Sprintf("please use --%s instead", flagAPIListenAddress))

	// Mark the old etcd flags as deprecated and maintain backward compability
	cmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)

	// Load the configuration file but only error out if flagConfigFile is used
	if err := viper.ReadInConfig(); err != nil && configFile != "" {
		setupErr = err
	}

	// Mark the old etcd keys as deprecated in the config file and then register
	// aliases for older etcd attributes in config file to maintain backward
	// compatiblity
	deprecatedConfigAttributes()
	viper.RegisterAlias(deprecatedFlagEtcdClientURLs, flagEtcdClientURLs)
	viper.RegisterAlias(deprecatedFlagEtcdInitialAdvertisePeerURLs, flagEtcdInitialAdvertisePeerURLs)
	viper.RegisterAlias(deprecatedFlagEtcdInitialCluster, flagEtcdInitialCluster)
	viper.RegisterAlias(deprecatedFlagEtcdInitialClusterState, flagEtcdInitialClusterState)
	viper.RegisterAlias(deprecatedFlagEtcdInitialClusterToken, flagEtcdInitialClusterToken)
	viper.RegisterAlias(deprecatedFlagEtcdNodeName, flagEtcdNodeName)
	viper.RegisterAlias(deprecatedFlagEtcdPeerURLs, flagEtcdPeerURLs)

	// Use our custom template for the start command
	cobra.AddTemplateFunc("categoryFlags", categoryFlags)
	cmd.SetUsageTemplate(startUsageTemplate)

	return cmd
}

func categoryFlags(category string, flags *pflag.FlagSet) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(category, pflag.ContinueOnError)

	flags.VisitAll(func(flag *pflag.Flag) {
		if categories, ok := flag.Annotations["categories"]; ok {
			if stringsutil.InArray(category, categories) {
				flagSet.AddFlag(flag)
			}
		} else if category == "" {
			// If no category was specified, return all flags without a category
			flagSet.AddFlag(flag)
		}
	})

	return flagSet
}

func aliasNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	// Wait until the command-line flags have been parsed
	if !f.Parsed() {
		return pflag.NormalizedName(name)
	}

	switch name {
	case deprecatedFlagEtcdClientURLs:
		deprecatedFlagMessage(name, flagEtcdClientURLs)
		name = flagEtcdClientURLs
	case deprecatedFlagEtcdInitialAdvertisePeerURLs:
		deprecatedFlagMessage(name, flagEtcdInitialAdvertisePeerURLs)
		name = flagEtcdInitialAdvertisePeerURLs
	case deprecatedFlagEtcdInitialCluster:
		deprecatedFlagMessage(name, flagEtcdInitialCluster)
		name = flagEtcdInitialCluster
	case deprecatedFlagEtcdInitialClusterState:
		deprecatedFlagMessage(name, flagEtcdInitialCluster)
		name = flagEtcdInitialClusterState
	case deprecatedFlagEtcdInitialClusterToken:
		deprecatedFlagMessage(name, flagEtcdInitialClusterToken)
		name = flagEtcdInitialClusterToken
	case deprecatedFlagEtcdNodeName:
		deprecatedFlagMessage(name, flagEtcdNodeName)
		name = flagEtcdNodeName
	case deprecatedFlagEtcdPeerURLs:
		deprecatedFlagMessage(name, flagEtcdPeerURLs)
		name = flagEtcdPeerURLs
	}
	return pflag.NormalizedName(name)
}

// Look up the deprecated attributes in our config file and print a warning
// message if set
func deprecatedConfigAttributes() {
	attributes := map[string]string{
		deprecatedFlagEtcdClientURLs:               flagEtcdClientURLs,
		deprecatedFlagEtcdInitialAdvertisePeerURLs: flagEtcdInitialAdvertisePeerURLs,
		deprecatedFlagEtcdInitialCluster:           flagEtcdInitialCluster,
		deprecatedFlagEtcdInitialClusterState:      flagEtcdInitialClusterState,
		deprecatedFlagEtcdInitialClusterToken:      flagEtcdInitialClusterToken,
		deprecatedFlagEtcdNodeName:                 flagEtcdNodeName,
		deprecatedFlagEtcdPeerURLs:                 flagEtcdPeerURLs,
	}

	for old, new := range attributes {
		if viper.IsSet(old) {
			logger.Warningf(
				"config attribute %s has been deprecated, please use %s instead",
				old, new,
			)
		}
	}
}

func deprecatedFlagMessage(oldFlag, newFlag string) {
	logger.Warningf("flag --%s has been deprecated, please use --%s instead",
		oldFlag, newFlag)
}
