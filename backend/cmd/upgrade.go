package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/etcd"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagSkipConfirm = "skip-confirm"
)

func UpgradeCommand() *cobra.Command {
	var setupErr error
	cmd := &cobra.Command{
		Use:           "upgrade",
		Short:         "upgrade a sensu installation from 5.x to 6.x",
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

			skipConfirm := viper.GetBool(flagSkipConfirm)
			if !skipConfirm {
				var confirm bool
				prompt := &survey.Confirm{
					Message: "Do you really want to upgrade your Sensu 5.x database to 6.x? This operation cannot be undone; make sure you back up your database!",
				}
				if err := survey.AskOne(prompt, &confirm, nil); err != nil {
					return err
				}
				if !confirm {
					return errors.New("upgrade aborted by operator")
				}
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
				goto upgrade
			}
			// All endpoints returned an error, return the latest one
			return err

		upgrade:
			if err := etcdstore.MigrateDB(context.Background(), client, etcdstore.Migrations); err != nil {
				return err
			}
			if len(etcdstore.EnterpriseMigrations) > 0 {
				if err := etcdstore.MigrateEnterpriseDB(context.Background(), client, etcdstore.EnterpriseMigrations); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().String(flagTimeout, defaultTimeout, "timeout, in seconds, for failing to establish a connection to etcd")
	cmd.Flags().Bool(flagSkipConfirm, false, "skip interactive confirmation")

	setupErr = handleConfig(cmd, os.Args[1:], false)

	return cmd
}
