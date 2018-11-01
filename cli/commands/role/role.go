package role

import (
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/pflag"
)

type roleOpts struct {
	Name      string
	Namespace string
}

func (opts *roleOpts) withFlags(flags *pflag.FlagSet) {
	if namespace := helpers.GetChangedStringValueFlag("namespace", flags); namespace != "" {
		opts.Namespace = namespace
	}
}
