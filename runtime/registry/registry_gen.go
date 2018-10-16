package registry

//go:generate go install github.com/sensu/sensu-go/internal/cmd/gen-register
//go:generate gen-register -pkg github.com/sensu/sensu-go -t registry.go.tmpl -o registry.go
//go:generate goimports -w registry.go
