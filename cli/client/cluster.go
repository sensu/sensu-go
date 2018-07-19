package client

import (
	"encoding/json"
	"fmt"

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
