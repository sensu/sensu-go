package cluster

import (
	"fmt"
	"io"

	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// HealthCommand gets the Sensu health status of a cluster
func HealthCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "health",
		Short:        "get sensu health status",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := cli.Client.Health()
			if err != nil {
				return err
			}
			err = helpers.PrintTitle(helpers.GetChangedStringValueFlag("format", cmd.Flags()), cli.Config.Format(), fmt.Sprintf("Cluster ID: %x", result.Header.ClusterId), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			clusterHealth := result.ClusterHealth
			alarms := result.Alarms
			err = helpers.Print(cmd, cli.Config.Format(), printHealthToTable, nil, clusterHealth)
			if err != nil {
				return err
			}

			if alarms != nil {
				err = helpers.Print(cmd, cli.Config.Format(), printAlarmsToTable, nil, alarms)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	return cmd
}

func printHealthToTable(result interface{}, w io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "ID",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth, ok := data.(*types.ClusterHealth)
				if !ok {
					return cli.TypeError
				}
				return fmt.Sprintf("%x", clusterHealth.MemberID)
			},
		},
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth, ok := data.(*types.ClusterHealth)
				if !ok {
					return cli.TypeError
				}
				return clusterHealth.Name
			},
		},
		{
			Title:       "Error",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth, ok := data.(*types.ClusterHealth)
				if !ok {
					return cli.TypeError
				}
				return fmt.Sprintf("%v", clusterHealth.Err)
			},
		},
		{
			Title:       "Healthy",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth, ok := data.(*types.ClusterHealth)
				if !ok {
					return cli.TypeError
				}
				return fmt.Sprintf("%t", clusterHealth.Healthy)
			},
		},
	})

	table.Render(w, result)
}

func printAlarmsToTable(result interface{}, w io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "ID",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				alarm, ok := data.(*etcdserverpb.AlarmMember)
				if !ok {
					return cli.TypeError
				}
				return fmt.Sprintf("%x", alarm.GetMemberID())
			},
		},
		{
			Title:       "Alarm Type",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				alarm, ok := data.(*etcdserverpb.AlarmMember)
				if !ok {
					return cli.TypeError
				}
				return alarm.Alarm.String()
			},
		},
	})
	table.Render(w, result)
}
