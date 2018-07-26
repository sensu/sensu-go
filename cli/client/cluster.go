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
		return nil, UnmarshalError(res)
	}
	var result clientv3.MemberListResponse
	return &result, json.Unmarshal(res.Body(), &result)
}

func (c *RestClient) MemberAdd(peerAddrs []string) (*clientv3.MemberAddResponse, error) {
	values := url.Values{"peer-addrs": {strings.Join(peerAddrs, ",")}}.Encode()
	endpoint := fmt.Sprintf("%s?%s", clusterMembersBasePath, values)
	res, err := c.R().Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("POST %q: %s", endpoint, err)
	}
	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}
	var result clientv3.MemberAddResponse
	return &result, json.Unmarshal(res.Body(), &result)
}

func (c *RestClient) MemberUpdate(id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error) {
	values := url.Values{"peer-addrs": {strings.Join(peerAddrs, ",")}}.Encode()
	endpoint := fmt.Sprintf("%s/%x?%s", clusterMembersBasePath, id, values)
	res, err := c.R().Put(endpoint)
	if err != nil {
		return nil, fmt.Errorf("PUT %q: %s", endpoint, err)
	}
	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}
	var result clientv3.MemberUpdateResponse
	return &result, json.Unmarshal(res.Body(), &result)
}

func (c *RestClient) MemberRemove(id uint64) (*clientv3.MemberRemoveResponse, error) {
	endpoint := fmt.Sprintf("%s/%x", clusterMembersBasePath, id)
	res, err := c.R().Delete(endpoint)
	if err != nil {
		return nil, fmt.Errorf("DELETE %q: %s", endpoint, err)
	}
	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}
	var result clientv3.MemberRemoveResponse
	return &result, json.Unmarshal(res.Body(), &result)
}
