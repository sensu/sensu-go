package schedulerd

import (
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

var _ = testSubscriber{}

type testSubscriber struct {
	ch chan interface{}
}

func (ts testSubscriber) Receiver() chan<- interface{} {
	return ts.ch
}

func isAssetResourceRequest(req storev2.ResourceRequest) bool {
	return req.APIVersion == "core/v2" && req.Type == "Asset"
}

func isCheckResourceRequest(req storev2.ResourceRequest) bool {
	return req.APIVersion == "core/v2" && req.Type == "CheckConfig"
}

func isHookResourceRequest(req storev2.ResourceRequest) bool {
	return req.APIVersion == "core/v2" && req.Type == "HookConfig"
}
