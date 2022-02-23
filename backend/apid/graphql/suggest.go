package graphql

import (
	"strconv"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/suggest"
)

var (
	SuggestSchema = DefaultSuggestSchema()
)

type Subscribable interface {
	GetSubscriptions() []string
}

type Commandable interface {
	GetCommand() string
}

type Timeoutable interface {
	GetTimeout() uint32
}

func subscriptionsFn(res v2.Resource) []string {
	return res.(Subscribable).GetSubscriptions()
}

func commandFn(res v2.Resource) []string {
	return []string{res.(Commandable).GetCommand()}
}

func timeoutFn(res v2.Resource) []string {
	t := res.(Timeoutable).GetTimeout()
	return []string{strconv.FormatUint(uint64(t), 10)}
}

func DefaultSuggestSchema() suggest.Register {
	return suggest.Register{
		&suggest.Resource{
			Group: "core/v2",
			Name:  "asset",
			Path:  "/api/core/v2/namespaces/{namespace}/assets",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
				&suggest.CustomField{
					Name: "filters",
					FieldFunc: func(res v2.Resource) []string {
						return res.(*v2.Asset).Filters
					},
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "check_config",
			Path:  "/api/core/v2/namespaces/{namespace}/checks",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
				&suggest.CustomField{
					Name: "proxy_entity_name",
					FieldFunc: func(res v2.Resource) []string {
						return []string{res.(*v2.CheckConfig).ProxyEntityName}
					},
				},
				&suggest.CustomField{
					Name:      "command",
					FieldFunc: commandFn,
				},
				&suggest.CustomField{
					Name:      "subscriptions",
					FieldFunc: subscriptionsFn,
				},
				&suggest.CustomField{
					Name:      "timeout",
					FieldFunc: timeoutFn,
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "entity",
			Path:  "/api/core/v2/namespaces/{namespace}/entities",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
				&suggest.ObjectField{
					Name: "system",
					Fields: []suggest.Field{
						&suggest.CustomField{
							Name: "os",
							FieldFunc: func(res v2.Resource) []string {
								return []string{res.(*v2.Entity).System.OS}
							},
						},
						&suggest.CustomField{
							Name: "platform",
							FieldFunc: func(res v2.Resource) []string {
								return []string{res.(*v2.Entity).System.Platform}
							},
						},
						&suggest.CustomField{
							Name: "platform_family",
							FieldFunc: func(res v2.Resource) []string {
								return []string{res.(*v2.Entity).System.PlatformFamily}
							},
						},
						&suggest.CustomField{
							Name: "arch",
							FieldFunc: func(res v2.Resource) []string {
								return []string{res.(*v2.Entity).System.Arch}
							},
						},
					},
				},
				&suggest.CustomField{
					Name:      "subscriptions",
					FieldFunc: subscriptionsFn,
				},
				&suggest.CustomField{
					Name: "user",
					FieldFunc: func(res v2.Resource) []string {
						return []string{res.(*v2.Entity).User}
					},
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "filter",
			Path:  "/api/core/v2/namespaces/{namespace}/filters",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "handler",
			Path:  "/api/core/v2/namespaces/{namespace}/handlers",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
				&suggest.CustomField{
					Name:      "command",
					FieldFunc: commandFn,
				},
				&suggest.CustomField{
					Name:      "timeout",
					FieldFunc: timeoutFn,
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "hook_config",
			Path:  "/api/core/v2/namespaces/{namespace}/hooks",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
				&suggest.CustomField{
					Name:      "command",
					FieldFunc: commandFn,
				},
				&suggest.CustomField{
					Name:      "timeout",
					FieldFunc: timeoutFn,
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "mutator",
			Path:  "/api/core/v2/namespaces/{namespace}/mutators",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
				&suggest.CustomField{
					Name:      "command",
					FieldFunc: commandFn,
				},
				&suggest.CustomField{
					Name:      "timeout",
					FieldFunc: timeoutFn,
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "silenced",
			Path:  "/api/core/v2/namespaces/{namespace}/silenced",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
				&suggest.CustomField{
					Name: "check",
					FieldFunc: func(res v2.Resource) []string {
						return []string{res.(*v2.Silenced).Check}
					},
				},
				&suggest.CustomField{
					Name: "subscription",
					FieldFunc: func(res v2.Resource) []string {
						return []string{res.(*v2.Silenced).Subscription}
					},
				},
				&suggest.CustomField{
					Name: "creator",
					FieldFunc: func(res v2.Resource) []string {
						return []string{res.(*v2.Silenced).Creator}
					},
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "user",
			Path:  "/api/core/v2/users",
			Fields: []suggest.Field{
				&suggest.CustomField{
					Name: "username",
					FieldFunc: func(res v2.Resource) []string {
						return []string{res.(*v2.User).Username}
					},
				},
				&suggest.CustomField{
					Name: "groups",
					FieldFunc: func(res v2.Resource) []string {
						return res.(*v2.User).Groups
					},
				},
			},
		},
		&suggest.Resource{
			Group: "core/v2",
			Name:  "pipeline",
			Path:  "/api/core/v2/pipeline",
			Fields: []suggest.Field{
				&suggest.ObjectField{
					Name: "metadata",
					Fields: []suggest.Field{
						suggest.NameField,
						suggest.LabelsField,
					},
				},
			},
		},
	}
}
