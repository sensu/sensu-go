package graphql

import (
	"strconv"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

func subscriptionsFn(res corev2.Resource) []string {
	return res.(Subscribable).GetSubscriptions()
}

func commandFn(res corev2.Resource) []string {
	return []string{res.(Commandable).GetCommand()}
}

func timeoutFn(res corev2.Resource) []string {
	t := res.(Timeoutable).GetTimeout()
	return []string{strconv.FormatUint(uint64(t), 10)}
}

func DefaultSuggestSchema() suggest.Register {
	return suggest.Register{
		&suggest.Resource{
			Group:      "core/v2",
			Name:       "asset",
			FilterFunc: corev2.AssetFields,
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
					FieldFunc: func(res corev2.Resource) []string {
						return res.(*corev2.Asset).Filters
					},
				},
			},
		},
		&suggest.Resource{
			Group:      "core/v2",
			Name:       "check_config",
			FilterFunc: corev2.CheckConfigFields,
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
					FieldFunc: func(res corev2.Resource) []string {
						return []string{res.(*corev2.CheckConfig).ProxyEntityName}
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
			Group:      "core/v2",
			Name:       "entity",
			FilterFunc: corev2.EntityFields,
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
							FieldFunc: func(res corev2.Resource) []string {
								return []string{res.(*corev2.Entity).System.OS}
							},
						},
						&suggest.CustomField{
							Name: "platform",
							FieldFunc: func(res corev2.Resource) []string {
								return []string{res.(*corev2.Entity).System.Platform}
							},
						},
						&suggest.CustomField{
							Name: "platform_family",
							FieldFunc: func(res corev2.Resource) []string {
								return []string{res.(*corev2.Entity).System.PlatformFamily}
							},
						},
						&suggest.CustomField{
							Name: "arch",
							FieldFunc: func(res corev2.Resource) []string {
								return []string{res.(*corev2.Entity).System.Arch}
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
					FieldFunc: func(res corev2.Resource) []string {
						return []string{res.(*corev2.Entity).User}
					},
				},
			},
		},
		&suggest.Resource{
			Group:      "core/v2",
			Name:       "filter",
			FilterFunc: corev2.EventFilterFields,
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
			Group:      "core/v2",
			Name:       "handler",
			FilterFunc: corev2.HandlerFields,
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
			Group:      "core/v2",
			Name:       "hook_config",
			FilterFunc: corev2.HookConfigFields,
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
			Group:      "core/v2",
			Name:       "mutator",
			FilterFunc: corev2.MutatorFields,
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
			Group:      "core/v2",
			Name:       "pipeline",
			FilterFunc: corev2.PipelineFields,
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
			Group:      "core/v2",
			Name:       "silenced",
			FilterFunc: corev2.SilencedFields,
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
					FieldFunc: func(res corev2.Resource) []string {
						return []string{res.(*corev2.Silenced).Check}
					},
				},
				&suggest.CustomField{
					Name: "subscription",
					FieldFunc: func(res corev2.Resource) []string {
						return []string{res.(*corev2.Silenced).Subscription}
					},
				},
				&suggest.CustomField{
					Name: "creator",
					FieldFunc: func(res corev2.Resource) []string {
						return []string{res.(*corev2.Silenced).Creator}
					},
				},
			},
		},
		&suggest.Resource{
			Group:      "core/v2",
			Name:       "user",
			FilterFunc: corev2.UserFields,
			Fields: []suggest.Field{
				&suggest.CustomField{
					Name: "username",
					FieldFunc: func(res corev2.Resource) []string {
						return []string{res.(*corev2.User).Username}
					},
				},
				&suggest.CustomField{
					Name: "groups",
					FieldFunc: func(res corev2.Resource) []string {
						return res.(*corev2.User).Groups
					},
				},
			},
		},
	}
}
