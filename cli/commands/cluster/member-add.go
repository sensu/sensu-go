package cluster

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	"go.etcd.io/etcd/clientv3"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// MemberAddCommand adds a member to a cluster
func MemberAddCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "member-add [NAME] [PEER-ADDRS]",
		Short:        "add cluster member to an existing cluster, with comma-separated peer addresses",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			memberName := args[0]
			peerAddrs := splitAndTrim(args[1])

			resp, err := cli.Client.MemberAdd(peerAddrs)
			if err != nil {
				return fmt.Errorf("couldn't add cluster member: %s", err)
			}

			tData := templateData{
				Name:              memberName,
				MemberAddResponse: resp,
			}

			return memberAddTmpl.Execute(cmd.OutOrStdout(), tData)
		},
	}

	return cmd
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

type templateData struct {
	Name string
	*clientv3.MemberAddResponse
}

func joinMembers(data templateData) string {
	result := make([]string, 0, len(data.Members))
	for _, m := range data.Members {
		if m.ID == data.Member.ID {
			m.Name = data.Name
		}
		for _, url := range m.PeerURLs {
			result = append(result, fmt.Sprintf("%s=%s", m.Name, url))
		}
	}
	return strings.Join(result, ",")
}

const memberAddTmplStr = `added member {{ hex .Member.ID }} to cluster

ETCD_NAME="{{ .Name }}"
ETCD_INITIAL_CLUSTER="{{ joinMembers . }}"
ETCD_INITIAL_CLUSTER_STATE="existing"
`

func hex(id uint64) string {
	return fmt.Sprintf("%x", id)
}

var memberAddTmpl = template.Must(
	template.New("memberadd").Funcs(
		template.FuncMap{"joinMembers": joinMembers, "hex": hex}).Parse(
		memberAddTmplStr))
