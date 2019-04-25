package cluster

import (
	"fmt"
	"io"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/spf13/cobra"
)

// MemberListCommand lists the cluster members
func MemberListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "member-list",
		Short:        "list cluster members",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := cli.Client.MemberList()
			if err != nil {
				return fmt.Errorf("error listing cluster members: %s", err)
			}
			err = helpers.PrintFormatted(helpers.GetChangedStringValueFlag("format", cmd.Flags()), cli.Config.Format(), result.Header, cmd.OutOrStdout(), printTitleClusterID)
			if err != nil {
				return err
			}
			return helpers.Print(cmd, cli.Config.Format(), printMemberListToTable, nil, result)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printMemberListToTable(result interface{}, w io.Writer) {
	memberList, ok := result.(*clientv3.MemberListResponse)
	if !ok {
		return
	}
	table := table.New([]*table.Column{
		{
			Title:       "ID",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				member, ok := data.(*etcdserverpb.Member)
				if !ok {
					return cli.TypeError
				}
				return fmt.Sprintf("%x", member.ID)
			},
		},
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				member, ok := data.(*etcdserverpb.Member)
				if !ok {
					return cli.TypeError
				}
				return member.Name
			},
		},
		{
			Title: "Peer URLs",
			CellTransformer: func(data interface{}) string {
				member, ok := data.(*etcdserverpb.Member)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(member.PeerURLs, ",")
			},
		},
		{
			Title: "Client URLs",
			CellTransformer: func(data interface{}) string {
				member, ok := data.(*etcdserverpb.Member)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(member.ClientURLs, ",")
			},
		},
	})

	table.Render(w, memberList.Members)
}
