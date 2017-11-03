package useractions

import (
	"reflect"

	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

func addOrgEnvToContext(ctx context.Context, record types.MultitenantResource) context.Context {
	ctx = context.WithValue(ctx, types.OrganizationKey, record.GetOrg())
	ctx = context.WithValue(ctx, types.EnvironmentKey, record.GetEnv())
	return ctx
}

func copyFields(target interface{}, source interface{}, fields ...string) {
	t := reflect.Indirect(reflect.ValueOf(target))
	s := reflect.Indirect(reflect.ValueOf(source))

	for _, f := range fields {
		reflect.Copy(t.FieldByName(f), s.FieldByName(f))
	}
}
