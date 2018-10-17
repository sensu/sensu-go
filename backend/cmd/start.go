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
	flagAPIHost               = "api-host"
	flagAPIPort               = "api-port"
	flagDashboardHost         = "dashboard-host"
	flagDashboardPort         = "dashboard-port"
	flagDeregistrationHandler = "deregistration-handler"
	flagStateDir              = "state-dir"
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"
	flagDebug                 = "debug"
	flagLogLevel              = "log-level"

	// Etcd flag constants
	deprecatedFlagEtcdClientURL               = "listen-client-urls"
	flagEtcdClientURL                         = "etcd-listen-client-urls"
	deprecatedFlagEtcdPeerURL                 = "listen-peer-urls"
	flagEtcdPeerURL                           = "etcd-listen-peer-urls"
	deprecatedFlagEtcdInitialCluster          = "initial-cluster"
	flagEtcdInitialCluster                    = "etcd-initial-cluster"
	deprecatedFlagEtcdInitialAdvertisePeerURL = "initial-advertise-peer-urls"
	flagEtcdInitialAdvertisePeerURL           = "etcd-initial-advertise-peer-urls"
	deprecatedFlagEtcdInitialClusterState     = "initial-cluster-state"
	flagEtcdInitialClusterState               = "etcd-initial-cluster-state"
	deprecatedFlagEtcdInitialClusterToken     = "initial-cluster-token"
	flagEtcdInitialClusterToken               = "etcd-initial-cluster-token"
	deprecatedFlagEtcdNodeName                = "name"
	flagEtcdNodeName                          = "etcd-name"
	flagNoEmbedEtcd                           = "no-embed-etcd"

	// Etcd TLS flag constants
	flagEtcdCertFile           = "etcd-cert-file"
	flagEtcdKeyFile            = "etcd-key-file"
	flagEtcdClientCertAuth     = "etcd-client-cert-auth"
	flagEtcdTrustedCAFile      = "etcd-trusted-ca-file"
	flagEtcdPeerCertFile       = "etcd-peer-cert-file"
	flagEtcdPeerKeyFile        = "etcd-peer-key-file"
	flagEtcdPeerClientCertAuth = "etcd-peer-client-cert-auth"
	flagEtcdPeerTrustedCAFile  = "etcd-peer-trusted-ca-file"

	// Default values

	// defaultEtcdClientURL is the default URL to listen for Etcd clients
	defaultEtcdClientURL = "http://127.0.0.1:2379"
	// defaultEtcdName is the default etcd member node name (single-node cluster
	// only)
	defaultEtcdName = "default"
	// DefaultEtcdPeerURL is the default URL to listen for Etcd peers (single-node
	// cluster only)
	defaultEtcdPeerURL = "http://127.0.0.1:2380"

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

			level, err := logrus.ParseLevel(viper.GetString(flagLogLevel))
			if err != nil {
				return err
			}
			logrus.SetLevel(level)

			cfg := &backend.Config{
				AgentHost:             viper.GetString(flagAgentHost),
				AgentPort:             viper.GetInt(flagAgentPort),
				APIHost:               viper.GetString(flagAPIHost),
				APIPort:               viper.GetInt(flagAPIPort),
				DashboardHost:         viper.GetString(flagDashboardHost),
				DashboardPort:         viper.GetInt(flagDashboardPort),
				DeregistrationHandler: viper.GetString(flagDeregistrationHandler),
				StateDir:              viper.GetString(flagStateDir),

				EtcdListenClientURL:         viper.GetString(flagEtcdClientURL),
				EtcdListenPeerURL:           viper.GetString(flagEtcdPeerURL),
				EtcdInitialCluster:          viper.GetString(flagEtcdInitialCluster),
				EtcdInitialClusterState:     viper.GetString(flagEtcdInitialClusterState),
				EtcdInitialAdvertisePeerURL: viper.GetString(flagEtcdInitialAdvertisePeerURL),
				EtcdInitialClusterToken:     viper.GetString(flagEtcdInitialClusterToken),
				EtcdName:                    viper.GetString(flagEtcdNodeName),
				NoEmbedEtcd:                 viper.GetBool(flagNoEmbedEtcd),
			}

			// Sensu APIs TLS config
			certFile := viper.GetString(flagCertFile)
			keyFile := viper.GetString(flagKeyFile)
			trustedCAFile := viper.GetString(flagTrustedCAFile)
			insecureSkipTLSVerify := viper.GetBool(flagInsecureSkipTLSVerify)

			if certFile != "" && keyFile != "" && trustedCAFile != "" {
				cfg.TLS = &types.TLSOptions{
					CertFile:           certFile,
					KeyFile:            keyFile,
					TrustedCAFile:      trustedCAFile,
					InsecureSkipVerify: insecureSkipTLSVerify,
				}
			} else if certFile != "" || keyFile != "" || trustedCAFile != "" {
				emptyFlags := []string{}
				if certFile == "" {
					emptyFlags = append(emptyFlags, flagCertFile)
				}
				if keyFile == "" {
					emptyFlags = append(emptyFlags, flagKeyFile)
				}
				if trustedCAFile == "" {
					emptyFlags = append(emptyFlags, flagTrustedCAFile)
				}

				return fmt.Errorf("missing the following cert flags: %s", emptyFlags)
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
					log.Println(http.ListenAndServe("localhost:6060", nil))
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
	viper.SetDefault(flagAPIHost, "[::]")
	viper.SetDefault(flagAPIPort, 8080)
	viper.SetDefault(flagDashboardHost, "[::]")
	viper.SetDefault(flagDashboardPort, 3000)
	viper.SetDefault(flagDeregistrationHandler, "")
	viper.SetDefault(flagStateDir, path.SystemDataDir())
	viper.SetDefault(flagCertFile, "")
	viper.SetDefault(flagKeyFile, "")
	viper.SetDefault(flagTrustedCAFile, "")
	viper.SetDefault(flagInsecureSkipTLSVerify, false)
	viper.SetDefault(flagLogLevel, "warn")

	// Etcd defaults
	viper.SetDefault(flagEtcdClientURL, defaultEtcdClientURL)
	viper.SetDefault(flagEtcdPeerURL, defaultEtcdPeerURL)
	viper.SetDefault(flagEtcdInitialCluster,
		fmt.Sprintf("%s=%s", defaultEtcdName, defaultEtcdPeerURL))
	viper.SetDefault(flagEtcdInitialAdvertisePeerURL, defaultEtcdPeerURL)
	viper.SetDefault(flagEtcdInitialClusterState, etcd.ClusterStateNew)
	viper.SetDefault(flagEtcdInitialClusterToken, "")
	viper.SetDefault(flagEtcdNodeName, defaultEtcdName)
	viper.SetDefault(flagNoEmbedEtcd, false)

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Main Flags
	cmd.Flags().String(flagAgentHost, viper.GetString(flagAgentHost), "agent listener host")
	cmd.Flags().Int(flagAgentPort, viper.GetInt(flagAgentPort), "agent listener port")
	cmd.Flags().String(flagAPIHost, viper.GetString(flagAPIHost), "http api listener host")
	cmd.Flags().Int(flagAPIPort, viper.GetInt(flagAPIPort), "http api port")
	cmd.Flags().String(flagDashboardHost, viper.GetString(flagDashboardHost), "dashboard listener host")
	cmd.Flags().Int(flagDashboardPort, viper.GetInt(flagDashboardPort), "dashboard listener port")
	cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "default deregistration handler")
	cmd.Flags().StringP(flagStateDir, "d", viper.GetString(flagStateDir), "path to sensu state storage")
	cmd.Flags().String(flagCertFile, viper.GetString(flagCertFile), "tls certificate")
	cmd.Flags().String(flagKeyFile, viper.GetString(flagKeyFile), "tls certificate key")
	cmd.Flags().String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "tls certificate authority")
	cmd.Flags().Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip ssl verification")
	cmd.Flags().Bool(flagDebug, false, "enable debugging and profiling features")
	cmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")

	// Etcd flags
	cmd.Flags().String(flagEtcdClientURL, viper.GetString(flagEtcdClientURL), "list of URLs to listen on for client traffic")
	_ = cmd.Flags().SetAnnotation(flagEtcdClientURL, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdPeerURL, viper.GetString(flagEtcdPeerURL), "list of URLs to listen on for peer traffic")
	_ = cmd.Flags().SetAnnotation(flagEtcdPeerURL, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdInitialCluster, viper.GetString(flagEtcdInitialCluster), "initial cluster configuration for bootstrapping")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialCluster, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdInitialAdvertisePeerURL, viper.GetString(flagEtcdInitialAdvertisePeerURL), "list of this member's peer URLs to advertise to the rest of the cluster")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialAdvertisePeerURL, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdInitialClusterState, viper.GetString(flagEtcdInitialClusterState), "initial cluster state (\"new\" or \"existing\")")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialClusterState, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdInitialClusterToken, viper.GetString(flagEtcdInitialClusterToken), "initial cluster token for the etcd cluster during bootstrap")
	_ = cmd.Flags().SetAnnotation(flagEtcdInitialClusterToken, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdNodeName, viper.GetString(flagEtcdNodeName), "human-readable name for this member")
	_ = cmd.Flags().SetAnnotation(flagEtcdNodeName, "categories", []string{"store"})
	cmd.Flags().Bool(flagNoEmbedEtcd, viper.GetBool(flagNoEmbedEtcd), "don't embed etcd, use external etcd instead")
	_ = cmd.Flags().SetAnnotation(flagNoEmbedEtcd, "categories", []string{"store"})

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
	viper.RegisterAlias(deprecatedFlagEtcdClientURL, flagEtcdClientURL)
	viper.RegisterAlias(deprecatedFlagEtcdInitialAdvertisePeerURL, flagEtcdInitialAdvertisePeerURL)
	viper.RegisterAlias(deprecatedFlagEtcdInitialCluster, flagEtcdInitialCluster)
	viper.RegisterAlias(deprecatedFlagEtcdInitialClusterState, flagEtcdInitialClusterState)
	viper.RegisterAlias(deprecatedFlagEtcdInitialClusterToken, flagEtcdInitialClusterToken)
	viper.RegisterAlias(deprecatedFlagEtcdNodeName, flagEtcdNodeName)
	viper.RegisterAlias(deprecatedFlagEtcdPeerURL, flagEtcdPeerURL)

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
	case deprecatedFlagEtcdClientURL:
		deprecatedFlagMessage(name, flagEtcdClientURL)
		name = flagEtcdClientURL
	case deprecatedFlagEtcdInitialAdvertisePeerURL:
		deprecatedFlagMessage(name, flagEtcdInitialAdvertisePeerURL)
		name = flagEtcdInitialAdvertisePeerURL
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
	case deprecatedFlagEtcdPeerURL:
		deprecatedFlagMessage(name, flagEtcdPeerURL)
		name = flagEtcdPeerURL
	}
	return pflag.NormalizedName(name)
}

// Look up the deprecated attributes in our config file and print a warning
// message if set
func deprecatedConfigAttributes() {
	attributes := map[string]string{
		deprecatedFlagEtcdClientURL:               flagEtcdClientURL,
		deprecatedFlagEtcdInitialAdvertisePeerURL: flagEtcdInitialAdvertisePeerURL,
		deprecatedFlagEtcdInitialCluster:          flagEtcdInitialCluster,
		deprecatedFlagEtcdInitialClusterState:     flagEtcdInitialClusterState,
		deprecatedFlagEtcdInitialClusterToken:     flagEtcdInitialClusterToken,
		deprecatedFlagEtcdNodeName:                flagEtcdNodeName,
		deprecatedFlagEtcdPeerURL:                 flagEtcdPeerURL,
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
