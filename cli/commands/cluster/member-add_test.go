package cluster

import (
	"bytes"
	"testing"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/etcdserverpb"
)

func TestRenderMemberAddTemplate(t *testing.T) {
	data := templateData{
		Name: "foo",
		MemberAddResponse: &clientv3.MemberAddResponse{
			Member: &etcdserverpb.Member{
				ID: 12345678,
			},
			Members: []*etcdserverpb.Member{
				{
					ID:       12345678,
					PeerURLs: []string{"http://127.0.0.1", "http://192.168.0.1"},
				},
				{
					ID:       42424242,
					Name:     "bar",
					PeerURLs: []string{"http://example.com"},
				},
			},
		},
	}

	want := `added member bc614e to cluster

ETCD_NAME="foo"
ETCD_INITIAL_CLUSTER="foo=http://127.0.0.1,foo=http://192.168.0.1,bar=http://example.com"
ETCD_INITIAL_CLUSTER_STATE="existing"
`

	buf := new(bytes.Buffer)

	if err := memberAddTmpl.Execute(buf, data); err != nil {
		t.Fatal(err)
	}

	if got := buf.String(); got != want {
		t.Errorf("bad template result: got %q, want %q", got, want)
	}
}
