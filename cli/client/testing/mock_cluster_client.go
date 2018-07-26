package testing

import "github.com/coreos/etcd/clientv3"

func (c *MockClient) MemberList() (*clientv3.MemberListResponse, error) {
	args := c.Called()
	return args.Get(0).(*clientv3.MemberListResponse), args.Error(1)
}

func (c *MockClient) MemberAdd(peerAddrs []string) (*clientv3.MemberAddResponse, error) {
	args := c.Called(peerAddrs)
	return args.Get(0).(*clientv3.MemberAddResponse), args.Error(1)
}

func (c *MockClient) MemberUpdate(id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error) {
	args := c.Called(id, peerAddrs)
	return args.Get(0).(*clientv3.MemberUpdateResponse), args.Error(1)
}

func (c *MockClient) MemberRemove(id uint64) (*clientv3.MemberRemoveResponse, error) {
	args := c.Called(id)
	return args.Get(0).(*clientv3.MemberRemoveResponse), args.Error(1)
}
