# Etcd Watcher usage tree

I wish there was some tool that would just make this for me.

```
etcd clientv3.Watcher.Watch(...)
|- sensu-go/backend/livensess
|   ... not included
|-sensu-go/backend/queue
|   ... not included
|-sensu-go/backend/ringv2
|   ... not included
 -sensu-go/backend/store/etcd func Watch(...) Watcher
  |- sensu-go/backend/agentd GetEntityConfigWatcher(...) <-chan store.WatchEventEntityConfig
  |  |- sensu-go/backend/backend.go for agentd daemon
  |
  |- sensu-go/backend/store/etcd store.GetCheckConfigWatcher(...) <-chan store.WatchEventCheckConfig
  |  |- sensu-go/backend/schedulerd/check_watcher.go for schedulerd
  |
  |- sensu-go/backend/store/etcd store.GetTessenConfigWatcher(...) <-chan store.WatchEventTessenConfig
  |  |- sensu-go/backend/tessend/tessend.go for tessend
  |
  |- sensu-go/backend/store/etcd GetResourceWatcher(...) <-chan store.WatchEventResource
  |  unused
  |
  |- sensu-go/backend/store/etcd GetResourceV3Watcher(...) <-chan store.WatchEventResourceV3
  |  |- sensu-enterprise-go/backend/backend.go creates serviceComponentWatcher used by bsmd
  |
  |- sensu-enterprise-go/backend/replicatord/watcher.go GetReplicatorWatcher(...) <-chan store.WatchEventResource
  |  |- sensu-enterprise-go/backend/backend.go for replicatord daemon
  |
  |- sensu-enterprise-go/backend/secrets/watcher.go Watcher.Start(...)
  |  |- sensu-enterprise-go/backend/backend.go starts secretsProviders.NewWatcher daemon
  |  
  |- sensu-enterprise-go/backend/store/provider/watcher.go Watcher.Start(...)
  |  |- sensu-enterprise-go/backend/backend.go starts storeprovider.NewWatcher daemon
  |  
  |- sensu-enterprise-go/backend/authentication/providers/watcher.go Watcher.Start(...)
  |  |- sensu-enterprise-go/backend/backend.go starts authProviders.NewWatcher daemon
  |  
  |- sensu-enterprise-go/backend/licensing/licensing.go Watcher.Start(...)
     |- sensu-enterprise-go/backend/backend.go starts licensing.New daemon
```

## Interfaces That Need Reimplemented

Usage of configuration store watchers fall into two categories in 6.X.

1. Using `sensu-go/backend/store/etcd`'s `Watch` function directly, usually from sensu-enterprise-go Daemons.
1. Using a wrapper utility around the `sensu-go/backend/store/etcd`'s `Watch` function that type casts the resource.
These sometimes include little bits of buisness logic to default or overwite resource attributes. Ex [agentd](https://github.com/sensu/sensu-go/blob/b208e5e7adad0e53ec7ef6403236e69f48d03dee/backend/agentd/watcher.go#L58-L62), [check config watcher](https://github.com/sensu/sensu-go/blob/b208e5e7adad0e53ec7ef6403236e69f48d03dee/backend/store/etcd/watchers.go#L37-L43).

### Wrappers

#### GetEntityConfigWatcher

Watcher used by agentd to remotely manage agents. Does some weird stuff with labels.

package: `github.com/sensu-go/backend/agentd`

implementation: https://github.com/sensu/sensu-go/blob/b208e5e7adad0e53ec7ef6403236e69f48d03dee/backend/agentd/watcher.go#L20

Inlined interface
```
GetEntityConfigWatcher(context.Context, *clientv3.Client) <-chan struct {
	Entity *corev3.EntityConfig
	Action WatchActionType
}
```

#### GetCheckConfigWatcher

Watcher used by schedulerd to start/stop/restart check scheduling. Does some weird stuff with the scheduler field.

package: `github.com/sensu-go/backend/store`

implementation: https://github.com/sensu/sensu-go/blob/b208e5e7adad0e53ec7ef6403236e69f48d03dee/backend/store/etcd/watchers.go#L17

Inlined interface
```
GetCheckConfigWatcher(context.Context) <-chan struct {
    CheckConfig *types.CheckConfig
    Action      WatchActionType
}
```

#### GetTessenConfigWatcher

Watcher used by tessend. Currently defiend in in the store interface.

package: `github.com/sensu-go/backend/store`

implementation: https://github.com/sensu/sensu-go/blob/b208e5e7adad0e53ec7ef6403236e69f48d03dee/backend/store/etcd/watchers.go#L59

Inlined interface
```
GetTessenConfigWatcher(context.Context) <-chan struct {
    TessenConfig *corev2.TessenConfig
    Action      WatchActionType
}
```

#### GetResourceV3Watcher

Watcher used by bsmd. Current defined in store interface.

package: `github.com/sensu-go/backend/store`

implementation: https://github.com/sensu/sensu-go/blob/4e17edcbe360e48ee88e9af028ee00a88484c197/backend/store/etcd/watchers.go#L131 

Inlined interface
```
GetResourceV3Watcher(ctx context.Context, client *clientv3.Client, key string) <-chan struct {
    Resource corev3.Resource
    Action   WatchActionType
}
```


## Proposal

I think we should standardize on watcher replacement for store/v2 resources that:
1. Uses the storev2 `ResourceRequest` or similar construct for identifying the resources to watch. This can be converted to either an etcd key prefix or postgres query.
2. Returns a structure similar to the `GetResourceV3Watcher` and other wrappers do today (tuple of Action and Resource.)
There was no situations where the sensu-go etcd Watch function was used that didn't involve some sort of condition around the action type and unmarshalling the payload to an assumed type.
Any situations where buisness logic is included in the watcher wrapper today can be lifted into wrappers around this new generic watcher interface.

```
// store/v2 interface similar to GetResourceV3Watcher
type Watcher interface {
    Watch(context.Context, ResourceRequest) <-chan WatchEvent // ResourceRequest already frequently resolved to etcd object/prefix key, and postgres record(s)
}

type WatchEvent struct {
    Action store.WatchActionType // or equiv storev2 replacement CREATE|UPDATE|DELETE|Error?
    Resource corev3.Resource
}
```
