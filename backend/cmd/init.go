package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/seeds"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

const (
	defaultTimeout = "5s"

	flagIgnoreAlreadyInitialized = "ignore-already-initialized"
	flagInitAdminUsername        = "cluster-admin-username"
	flagInitAdminPassword        = "cluster-admin-password"
	flagInteractive              = "interactive"
	flagTimeout                  = "timeout"
	flagWait                     = "wait"
	flagInitAdminAPIKey          = "cluster-admin-api-key"
)

var errEtcdEndpointUnreachable = errors.New("etcd endpoint could not be reached")

type initConfig struct {
	backend.Config
	SeedConfig seeds.Config
	Timeout    time.Duration
}

func (c *initConfig) Validate() error {
	if c.SeedConfig.AdminUsername == "" || c.SeedConfig.AdminPassword == "" {
		return fmt.Errorf("both %s and %s are required to be set (or an API key)", flagInitAdminUsername, flagInitAdminPassword)
	}
	return nil
}

type initOpts struct {
	AdminUsername             string `survey:"cluster-admin-username"`
	AdminPassword             string `survey:"cluster-admin-password"`
	AdminPasswordConfirmation string `survey:"cluster-admin-password-confirmation"`
	AdminAPIKey               string `survey:"cluster-admin-api-key"`
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
		{
			Name: "cluster-admin-password-confirmation",
			Prompt: &survey.Password{
				Message: "Retype Cluster Admin Password:",
			},
			Validate: survey.Required,
		},
		{
			Name: "cluster-admin-api-key",
			Prompt: &survey.Input{
				Message: "Cluster Admin API Key:",
			},
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
				EtcdClientURLs: viper.GetStringSlice(flagEtcdClientURLs),
			}

			// Sensu APIs TLS config
			certFile := viper.GetString(flagCertFile)
			keyFile := viper.GetString(flagKeyFile)
			insecureSkipTLSVerify := viper.GetBool(flagInsecureSkipTLSVerify)
			trustedCAFile := viper.GetString(flagTrustedCAFile)

			// Optional username/password auth
			etcdClientUsername := viper.GetString(envEtcdClientUsername)
			etcdClientPassword := viper.GetString(envEtcdClientPassword)

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

			timeout := viper.GetDuration(flagTimeout)
			if timeout < 1*time.Second {
				timeout = timeout * time.Second
			}

			initConfig := initConfig{
				Config: *cfg,
				SeedConfig: seeds.Config{
					AdminUsername: viper.GetString(flagInitAdminUsername),
					AdminPassword: viper.GetString(flagInitAdminPassword),
					AdminAPIKey:   viper.GetString(flagInitAdminAPIKey),
				},
				Timeout: timeout,
			}

			wait := viper.GetBool(flagWait)

			if viper.GetBool(flagInteractive) {
				var opts initOpts
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
				if opts.AdminPassword != opts.AdminPasswordConfirmation {
					//lint:ignore ST1005 this error is written to stdout/stderr
					return errors.New("Password confirmation doesn't match the password")
				}
				initConfig.SeedConfig.AdminUsername = opts.AdminUsername
				initConfig.SeedConfig.AdminPassword = opts.AdminPassword
				initConfig.SeedConfig.AdminAPIKey = opts.AdminAPIKey
			}

			if err := initConfig.Validate(); err != nil {
				return err
			}

			// Make sure at least one of the provided endpoints is reachable. This is
			// required to debug TLS errors because the seeding below will not print
			// the latest connection error (see
			// https://github.com/sensu/sensu-go/issues/3663)
			var clientConfig clientv3.Config
			for {
				for _, url := range clientURLs {
					logger.Infof("attempting to connect to etcd server: %s", url)

					if etcdClientUsername != "" && etcdClientPassword != "" {
						clientConfig = clientv3.Config{
							Endpoints: []string{url},
							Username:  etcdClientUsername,
							Password:  etcdClientPassword,
							TLS:       tlsConfig,
							DialOptions: []grpc.DialOption{
								grpc.WithReturnConnectionError(),
								grpc.WithBlock(),
							},
						}
					} else {
						clientConfig = clientv3.Config{
							Endpoints: []string{url},
							TLS:       tlsConfig,
							DialOptions: []grpc.DialOption{
								grpc.WithReturnConnectionError(),
								grpc.WithBlock(),
							},
						}
					}
					err := initializeStore(clientConfig, initConfig, url)
					if err != nil {
						if errors.Is(err, seeds.ErrAlreadyInitialized) {
							if viper.GetBool(flagIgnoreAlreadyInitialized) {
								return nil
							}
							return err
						}
						logger.Error(err.Error())
						continue
					}
					return nil
				}
				if !wait {
					return errors.New("no etcd endpoints are available or cluster is unhealthy")
				}
			}
		},
	}

	cmd.Flags().Bool(flagIgnoreAlreadyInitialized, false, "exit 0 if the cluster has already been initialized")
	cmd.Flags().String(flagInitAdminUsername, "", "cluster admin username")
	cmd.Flags().String(flagInitAdminPassword, "", "cluster admin password")
	cmd.Flags().Bool(flagInteractive, false, "interactive mode")
	cmd.Flags().String(flagTimeout, defaultTimeout, "duration to wait before a connection attempt to etcd is considered failed (must be >= 1s)")
	cmd.Flags().Bool(flagWait, false, "continuously retry to establish a connection to etcd until it is successful")
	cmd.Flags().String(flagInitAdminAPIKey, "", "cluster admin API key")

	setupErr = handleConfig(cmd, os.Args[1:], false)

	return cmd
}

func initializeStore(clientConfig clientv3.Config, initConfig initConfig, endpoint string) error {
	ctx, cancel := context.WithTimeout(
		clientv3.WithRequireLeader(context.Background()), initConfig.Timeout)
	defer cancel()

	clientConfig.Context = ctx

	client, err := clientv3.New(clientConfig)
	if err != nil {
		return fmt.Errorf("error connecting to etcd endpoint: %w", err)
	}
	defer func() { _ = client.Close() }()

	// Check if etcd endpoint is reachable
	if _, err := client.Status(ctx, endpoint); err != nil {
		// Etcd's client interceptor will log the actual underlying error.
		return errEtcdEndpointUnreachable
	}

	// The endpoint did not return any error, therefore we can proceed
	store := etcdstore.NewStore(client)
	if err := seeds.SeedCluster(ctx, store, client, initConfig.SeedConfig); err != nil {
		if errors.Is(err, seeds.ErrAlreadyInitialized) {
			return err
		}
		return fmt.Errorf("error seeding cluster, is cluster healthy? %w", err)
	}

	return nil
}
