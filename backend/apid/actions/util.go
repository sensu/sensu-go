package actions

import (
	"reflect"

	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

func addOrgEnvToContext(
	ctx context.Context,
	record types.MultitenantResource,
) context.Context {
	return types.SetContextFromResource(ctx, record)
}

func copyFields(target interface{}, source interface{}, fields ...string) {
	t := reflect.Indirect(reflect.ValueOf(target))
	s := reflect.Indirect(reflect.ValueOf(source))

	for _, f := range fields {
		t.FieldByName(f).Set(s.FieldByName(f))
	}
}
