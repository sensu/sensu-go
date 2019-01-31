package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/coreos/etcd/clientv3"
)

var clusterMembersPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "cluster", "members")

func (c *RestClient) MemberList() (*clientv3.MemberListResponse, error) {
	path := clusterMembersPath()
	res, err := c.R().Get(path)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %s", path, err)
	}
	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}
	var result clientv3.MemberListResponse
	return &result, json.Unmarshal(res.Body(), &result)
}

func (c *RestClient) MemberAdd(peerAddrs []string) (*clientv3.MemberAddResponse, error) {
	values := url.Values{"peer-addrs": {strings.Join(peerAddrs, ",")}}.Encode()
	endpoint := fmt.Sprintf("%s?%s", clusterMembersPath(), values)
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
	endpoint := fmt.Sprintf("%s/%x?%s", clusterMembersPath(), id, values)
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
	endpoint := fmt.Sprintf("%s/%x", clusterMembersPath(), id)
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
