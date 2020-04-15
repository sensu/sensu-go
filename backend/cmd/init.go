package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey"
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

const (
	defaultTimeout = "5"

	flagInitAdminUsername = "cluster-admin-username"
	flagInitAdminPassword = "cluster-admin-password"
	flagInteractive       = "interactive"
	flagTimeout           = "timeout"
)

type seedConfig struct {
	backend.Config
	SeedConfig seeds.Config
	Timeout    time.Duration
}

type initOpts struct {
	AdminUsername string `survey:"cluster-admin-username"`
	AdminPassword string `survey:"cluster-admin-password"`
}

func (i *initOpts) administerQuestionnaire() error {
	qs := []*survey.Question{
		{
			Name: "cluster-admin-username",
			Prompt: &survey.Input{
				Message: "Cluster Admin Username:",
			},
			Validate: survey.Required,
		},
		{
			Name: "cluster-admin-password",
			Prompt: &survey.Password{
				Message: "Cluster Admin Password:",
			},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, i)
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
				EtcdClientURLs:      fallbackStringSlice(flagEtcdClientURLs, flagEtcdAdvertiseClientURLs),
				EtcdCipherSuites:    viper.GetStringSlice(flagEtcdCipherSuites),
				EtcdMaxRequestBytes: viper.GetUint(flagEtcdMaxRequestBytes),
				NoEmbedEtcd:         true,
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

			clientURLs := viper.GetStringSlice(flagEtcdClientURLs)
			if len(clientURLs) == 0 {
				clientURLs = viper.GetStringSlice(flagEtcdAdvertiseClientURLs)
			}

			timeout := viper.GetDuration(flagTimeout)

			client, err := clientv3.New(clientv3.Config{
				Endpoints:   clientURLs,
				DialTimeout: timeout * time.Second,
				TLS:         tlsConfig,
			})

			if err != nil {
				return fmt.Errorf("error connecting to cluster: %s", err)
			}

			uname := viper.GetString(flagInitAdminUsername)
			pword := viper.GetString(flagInitAdminPassword)

			if viper.GetBool(flagInteractive) {
				var opts initOpts
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
				uname = opts.AdminUsername
				pword = opts.AdminPassword
			}

			if uname == "" || pword == "" {
				return fmt.Errorf("both %s and %s are required to be set", flagInitAdminUsername, flagInitAdminPassword)
			}

			seedConfig := seedConfig{
				Config: *cfg,
				SeedConfig: seeds.Config{
					AdminUsername: uname,
					AdminPassword: pword,
				},
				Timeout: timeout,
			}

			// Make sure at least one of the provided endpoints is reachable. This is
			// required to debug TLS errors because the seeding below will not print
			// the latest connection error (see
			// https://github.com/sensu/sensu-go/issues/3663)
			for _, url := range clientURLs {
				tctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
				defer cancel()
				_, err = client.Status(tctx, url)
				if err != nil {
					// We do not need to log the error, etcd's client interceptor will log
					// the actual underlying error
					continue
				}
				// The endpoint did not return any error, therefore we can proceed
				goto seed
			}
			// All endpoints returned an error, return the latest one
			return err

		seed:
			return seedCluster(client, seedConfig)
		},
	}

	cmd.Flags().String(flagInitAdminUsername, "", "cluster admin username")
	cmd.Flags().String(flagInitAdminPassword, "", "cluster admin password")
	cmd.Flags().Bool(flagInteractive, false, "interactive mode")
	cmd.Flags().String(flagTimeout, defaultTimeout, "timeout, in seconds, for failing to establish a connection to etcd")

	setupErr = handleConfig(cmd, false)

	return cmd
}

func seedCluster(client *clientv3.Client, config seedConfig) error {
	store := etcdstore.NewStore(client, "")
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout*time.Second)
	defer cancel()
	if err := seeds.SeedCluster(ctx, store, config.SeedConfig); err != nil {
		return err
	}
	return nil
}
