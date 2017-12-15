## Relevant Material

Example of expected output: backend/apid/graphqlschema/event.go
Example of code generator: backend/apid/graphql/generator/scalar.go
Reference for code generator library: https://github.com/dave/jennifer
Example schema: backend/apid/graphql/schema/schema-kitchen-sink.graphql

Run command to test:

```shell
go run scripts/gengraphql.go -debug backend/apid/graphql/schema
cat backend/apid/graphql/schema/schema-kitchen-sink.gql.go
```
