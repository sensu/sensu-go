package mockapi

//go:generate go install github.com/golang/mock/mockgen@v1.6.0
//go:generate go install github.com/rjeczalik/interfaces/cmd/interfacer@v0.1.1
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.AssetClient -as mockapi.AssetClient -o asset_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.AuthenticationClient -as mockapi.AuthenticationClient -o authn_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.CheckClient -as mockapi.CheckClient -o check_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.EntityClient -as mockapi.EntityClient -o entity_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.EventClient -as mockapi.EventClient -o event_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.EventFilterClient -as mockapi.EventFilterClient -o filter_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.HandlerClient -as mockapi.HandlerClient -o handler_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.HookConfigClient -as mockapi.HookConfigClient -o hook_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.NamespaceClient -as mockapi.NamespaceClient -o namespace_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.RBACClient -as mockapi.RBACClient -o rbac_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.SilencedClient -as mockapi.SilencedClient -o silenced_iface.go
//go:generate interfacer -for github.com/sensu/sensu-go/backend/api.UserClient -as mockapi.UserClient -o user_iface.go
//go:generate mockgen -destination mocks.go -package mockapi github.com/sensu/sensu-go/backend/api/mockapi AssetClient,AuthenticationClient,CheckClient,EntityClient,EventClient,EventFilterClient,HandlerClient,HookConfigClient,NamespaceClient,RBACClient,SilencedClient,UserClient
