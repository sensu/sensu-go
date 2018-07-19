package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/coreos/etcd/clientv3"
)

const clusterMembersBasePath = "/cluster/members"

func (c *RestClient) MemberList() (*clientv3.MemberListResponse, error) {
	res, err := c.R().Get(clusterMembersBasePath)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %s", clusterMembersBasePath, err)
	}
	if res.StatusCode() >= 400 {
		return nil, unmarshalError(res)
	}
	var result clientv3.MemberListResponse
	return &result, json.Unmarshal(res.Body(), &result)
}

func (c *RestClient) MemberAdd(peerAddrs []string) (*clientv3.MemberAddResponse, error) {
	values := url.Values{"peer-addrs": {strings.Join(peerAddrs, ",")}}
	endpoint := fmt.Sprintf("%s?%s", clusterMembersBasePath, values.Encode())
	res, err := c.R().Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("POST %q: %s", endpoint, err)
	}
	fmt.Println(endpoint)
	if res.StatusCode() >= 400 {
		return nil, unmarshalError(res)
	}
	var result clientv3.MemberAddResponse
	return &result, json.Unmarshal(res.Body(), &result)
}
