package cmd

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/seeds"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type seedConfig struct {
	backend.Config
}

// InitCommand is the 'sensu-backend init' subcommand.
func InitCommand() *cobra.Command {
	var setupErr error
	cmd := &cobra.Command{
		Use:           "init",
		Short:         "initialize a new sensu installation",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())
			if setupErr != nil {
				return setupErr
			}

			cfg := &backend.Config{
				EtcdAdvertiseClientURLs:      viper.GetStringSlice(flagEtcdAdvertiseClientURLs),
				EtcdListenClientURLs:         viper.GetStringSlice(flagEtcdClientURLs),
				EtcdListenPeerURLs:           viper.GetStringSlice(flagEtcdPeerURLs),
				EtcdInitialCluster:           viper.GetString(flagEtcdInitialCluster),
				EtcdInitialClusterState:      viper.GetString(flagEtcdInitialClusterState),
				EtcdInitialAdvertisePeerURLs: viper.GetStringSlice(flagEtcdInitialAdvertisePeerURLs),
				EtcdInitialClusterToken:      viper.GetString(flagEtcdInitialClusterToken),
				EtcdName:                     viper.GetString(flagEtcdNodeName),
				EtcdCipherSuites:             viper.GetStringSlice(flagEtcdCipherSuites),
				EtcdQuotaBackendBytes:        viper.GetInt64(flagEtcdQuotaBackendBytes),
				EtcdMaxRequestBytes:          viper.GetUint(flagEtcdMaxRequestBytes),
				EtcdHeartbeatInterval:        viper.GetUint(flagEtcdHeartbeatInterval),
				EtcdElectionTimeout:          viper.GetUint(flagEtcdElectionTimeout),
				NoEmbedEtcd:                  true,
			}

			// Sensu APIs TLS config
			certFile := viper.GetString(flagCertFile)
			keyFile := viper.GetString(flagKeyFile)
			insecureSkipTLSVerify := viper.GetBool(flagInsecureSkipTLSVerify)
			trustedCAFile := viper.GetString(flagTrustedCAFile)

			if certFile != "" && keyFile != "" {
				cfg.TLS = &corev2.TLSOptions{
					CertFile:           certFile,
					KeyFile:            keyFile,
					TrustedCAFile:      trustedCAFile,
					InsecureSkipVerify: insecureSkipTLSVerify,
				}
			} else if certFile != "" || keyFile != "" {
				return fmt.Errorf(
					"tls configuration error, both flags --%s & --%s are required",
					flagCertFile, flagKeyFile)
			}

			// Etcd TLS config
			cfg.EtcdClientTLSInfo = etcd.TLSInfo{
				CertFile:       viper.GetString(flagEtcdCertFile),
				KeyFile:        viper.GetString(flagEtcdKeyFile),
				TrustedCAFile:  viper.GetString(flagEtcdTrustedCAFile),
				ClientCertAuth: viper.GetBool(flagEtcdClientCertAuth),
			}

			// Convert the TLS config into etcd's transport.TLSInfo
			tlsInfo := (transport.TLSInfo)(cfg.EtcdClientTLSInfo)
			tlsConfig, err := tlsInfo.ClientConfig()
			if err != nil {
				return err
			}

			client, err := clientv3.New(clientv3.Config{
				Endpoints:   cfg.EtcdAdvertiseClientURLs,
				DialTimeout: 5 * time.Second,
				TLS:         tlsConfig,
			})

			if err != nil {
				return fmt.Errorf("error initializing cluster: %s", err)
			}

			seedConfig := seedConfig{Config: *cfg}

			return seedCluster(client, seedConfig)
		},
	}

	setupErr = handleConfig(cmd)

	return cmd
}

func seedCluster(client *clientv3.Client, config seedConfig) error {
	store := etcdstore.NewStore(client, config.EtcdName)
	if err := seeds.SeedInitialData(store); err != nil {
		return fmt.Errorf("error initializing cluster: %s", err)
	}
	return nil
}
