package provider

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Info represents the config of a store provider.
type Info struct {
	corev2.TypeMeta
	corev2.ObjectMeta
}

// InfoGetter gets info about a store provider.
type InfoGetter interface {
	// GetProviderInfo gets info about a store provider.
	GetProviderInfo() *Info
}
