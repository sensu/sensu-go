Sensu RPC Facilities
====================

Extensions
----------

```
+---------------+       +-------------+
|               +------->             |
| sensu-backend |       | sensu-agent |
|               <-------+             |
+-----+---^-----+  wss  +-------------+
      |   |
      |   | grpc
      |   |
  +---v---+---+
  |           |
  | extension |
  |           |
  +-----------+
```

Sensu enables user-defined filters, mutators and handlers via a gRPC interface.
Developers can implement a gRPC service, and connect it to Sensu. Sensu will
call the methods of the gRPC service during pipeline evaluation.

Because extensions use gRPC, developers can write extensions for Sensu in any
language that gRPC supports.

To write a Sensu extension, compile the `extension.proto` file in this package
with protoc and a language plugin. For instance,

```
protoc -I ../../../../ -I . -I ../types/ -I ../vendor/ --go_out=plugins=grpc:. extension.proto
```

Is the command for building the `extension.proto` file for Go. However, users
can compile the `extension.proto` file with any language that gRPC supports.

See  https://grpc.io/docs/quickstart/ for more information on which languages are
supported.

Once the proto has been compiled, use the generated server interface to write
an extension definition. This means writing the HandleEvent, MutateEvent and
FilterEvent methods, and setting up the service to listen on a port.

If you're working in Go, you can make use of Sensu's extension framework simply
by importing it into your code. See the
[example](https://github.com/sensu/sensu-go/blob/master/rpc/extension/example/main.go)
for details.

Once you have successfully authored an extension, it can be called by Sensu.
Extensions will need to be managed as their own service, external to Sensu.
While it is technically possible for Sensu to access these extensions
over remote links, it is far more reliable to operate them on the same node
the backend is running on.

See the `sensuctl extension` command for more information.
