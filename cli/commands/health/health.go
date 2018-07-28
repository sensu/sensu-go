package health

import (
	"fmt"
	"io"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

func HealthCommand(cli *cli.SensuCli) {
	cmd := &cobra.Command{
		Use:          "health",
		Short:        "get sensu health status",
		SilenceUsage: false,
		Run: func(cmd *cobra.Command, args []string) {
			result, _ := cli.Client.Health()
			return helpers.Print(cmd, cli.Config.Format(), printToTable, nil, result)
		},
	}
}

func printToTable(result interface{}, w io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "ID",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth := data.(*types.ClusterHealth)
				return fmt.Sprintf("%x", clusterHealth.MemberID)
			},
		},
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth := data.(*types.ClusterHealth)
				return clusterHealth.Name
			},
		},
		{
			Title:       "Error",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth := data.(*types.ClusterHealth)
				return clusterHealth.Err.Error()
			},
		},
		{
			Title:       "Healthy",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterHealth := data.(*types.ClusterHealth)
				return fmt.Sprintf("%t", clusterHealth.Healthy)
			},
		},
	})

	table.Render(w, result)
}
