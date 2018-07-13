package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.StandardErrorFieldResolvers = (*stdErrImpl)(nil)

type stdErr struct {
	code    schema.ErrCode
	input   string
	message string
}

func newStdErr(input string, err error) stdErr {
	out := stdErr{code: schema.ErrCodes.ERR_INTERNAL, input: input}
	switch terr := err.(type) {
	case actions.Error:
		out.message = terr.Message
		out.code = mapServiceErrCode(terr.Code)
	case error:
		out.message = err.Error()
	}
	return out
}

func mapServiceErrCode(code actions.ErrCode) schema.ErrCode {
	switch code {
	case actions.NotFound:
		return schema.ErrCodes.ERR_NOT_FOUND
	case actions.AlreadyExistsErr:
		return schema.ErrCodes.ERR_ALREADY_EXISTS
	case actions.PermissionDenied:
		return schema.ErrCodes.ERR_PERMISSION_DENIED
	case actions.InternalErr:
		fallthrough
	default:
		return schema.ErrCodes.ERR_INTERNAL
	}
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
