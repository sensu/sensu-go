package middlewares

import (
	"encoding/json"
	"net/http"

	"github.com/sensu/sensu-go/backend/apid/actions"
)

func writeErr(w http.ResponseWriter, err error) {
	var errRes actions.Error
	if erro, ok := err.(actions.Error); ok {
		errRes = erro
	} else {
		errRes = actions.Error{
			Code:    actions.InternalErr,
			Message: err.Error(),
		}
	}

	var st int
	switch errRes.Code {
	case actions.InternalErr:
		st = http.StatusInternalServerError
	case actions.InvalidArgument:
		st = http.StatusBadRequest
	case actions.NotFound:
		st = http.StatusNotFound
	case actions.AlreadyExistsErr:
		st = http.StatusConflict
	case actions.PermissionDenied:
		st = http.StatusForbidden
	case actions.Unauthenticated:
		st = http.StatusUnauthorized
	}

	errJSON, err := json.Marshal(errRes)
	if err != nil {
		Logger.WithError(err).Error("unable to marshal error")
	}

	w.WriteHeader(st)
	_, _ = w.Write(errJSON)
}
