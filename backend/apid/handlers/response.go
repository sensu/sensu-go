package handlers

import (
	corev3 "github.com/sensu/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

type HandlerResponse struct {
	Resource     corev3.Resource
	ResourceList []corev3.Resource
	TxInfo       storev2.TxInfo
	GraphQL      interface{} // unfortunate
}

func (h HandlerResponse) IsEmpty() bool {
	return (h.Resource == nil &&
		h.ResourceList == nil &&
		h.GraphQL == nil)
}
