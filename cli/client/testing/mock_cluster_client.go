package testing

import "github.com/coreos/etcd/clientv3"

func (c *MockClient) MemberList() (*clientv3.MemberListResponse, error) {
	args := c.Called()
	return args.Get(0).(*clientv3.MemberListResponse), args.Error(1)
}
