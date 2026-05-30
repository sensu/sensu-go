package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.StandardErrorFieldResolvers = (*stdErrImpl)(nil)
var _ graphql.InterfaceTypeResolver = (*errImpl)(nil)

type stdErr struct {
	code    schema.ErrCode
	input   string
	message string
}

func newStdErr(input string, err error) stdErr {
	out := stdErr{
		code:    schema.ErrCodes.ERR_INTERNAL,
		input:   input,
		message: err.Error(),
	}
	if err == authorization.ErrUnauthorized || err == authorization.ErrNoClaims {
		out.code = schema.ErrCodes.ERR_PERMISSION_DENIED
	}
	switch err.(type) {
	case (*store.ErrAlreadyExists):
		out.code = schema.ErrCodes.ERR_ALREADY_EXISTS
	case (*store.ErrNotFound):
		out.code = schema.ErrCodes.ERR_NOT_FOUND
	case (*store.ErrThreshold):
		out.code = schema.ErrCodes.ERR_THRESHOLD_REACHED
	}
	return out
}

func wrapInputErrors(input string, errs ...error) []stdErr {
	out := make([]stdErr, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		out = append(out, newStdErr(input, err))
	}
	return out
}

//
// Implement StandardError
//

type stdErrImpl struct{}

func (stdErrImpl) Input(p graphql.ResolveParams) (string, error) {
	record := p.Source.(stdErr)
	return record.input, nil
}

func (stdErrImpl) Code(p graphql.ResolveParams) (schema.ErrCode, error) {
	record := p.Source.(stdErr)
	return record.code, nil
}

func (stdErrImpl) Message(p graphql.ResolveParams) (string, error) {
	record := p.Source.(stdErr)
	return record.message, nil
}

func (stdErrImpl) IsTypeOf(record interface{}, _ graphql.IsTypeOfParams) bool {
	_, ok := record.(stdErr)
	return ok
}

//
// Implement Error
//

type errImpl struct{}

func (errImpl) ResolveType(obj interface{}, _ graphql.ResolveTypeParams) *graphql.Type {
	switch obj.(type) {
	case stdErr:
		return &schema.StandardErrorType
	}
	return nil
}
