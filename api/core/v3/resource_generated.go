package v3

// automatically generated file, do not edit!

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"reflect"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// implement validator to add validation to Validate methods
type validator interface {
	validate() error
}

// implement storeNamer to override StoreName methods
type storeNamer interface {
	storeName() string
}

// implement rbacNamer to override RBACName methods
type rbacNamer interface {
	rbacName() string
}

// implement uriPather to override URIPath methods
type uriPather interface {
	uriPath() string
}

type getMetadataer interface {
	GetMetadata() *corev2.ObjectMeta
}

func uriPath(typename, namespace, name string) string {
	if namespace == "" {
		return path.Join("api", "core", "v3", typename, url.PathEscape(name))
	}
	return path.Join("api", "core", "v3", "namespaces", url.PathEscape(namespace), typename, url.PathEscape(name))
}

// SetMetadata sets the provided metadata on the type. If the type does not
// have any metadata, nothing will happen.
func (e *EntityConfig) SetMetadata(meta *corev2.ObjectMeta) {
	// The function has to use reflection, since not all of the generated types
	// will have metadata.
	value := reflect.Indirect(reflect.ValueOf(e))
	field := value.FieldByName("Metadata")
	if !field.CanSet() {
		return
	}
	field.Set(reflect.ValueOf(meta))
}

// StoreName returns the store name for EntityConfig. It will be
// overridden if there is a method for EntityConfig called "storeName".
func (e *EntityConfig) StoreName() string {
	var iface interface{} = e
	if prefixer, ok := iface.(storeNamer); ok {
		return prefixer.storeName()
	}
	return "entity_configs"
}

// RBACName returns the RBAC name for EntityConfig. It will be overridden if
// there is a method for EntityConfig called "rbacName".
func (e *EntityConfig) RBACName() string {
	var iface interface{} = e
	if namer, ok := iface.(rbacNamer); ok {
		return namer.rbacName()
	}
	return "entity_configs"
}

// URIPath returns the URI path for EntityConfig. It will be overridden if
// there is a method for EntityConfig called uriPath.
func (e *EntityConfig) URIPath() string {
	var iface interface{} = e
	if pather, ok := iface.(uriPather); ok {
		return pather.uriPath()
	}
	metaer, ok := iface.(getMetadataer)
	if !ok {
		return ""
	}
	meta := metaer.GetMetadata()
	if meta == nil {
		return uriPath("entity-configs", "", "")
	}
	return uriPath("entity-configs", meta.Namespace, meta.Name)
}

// Validate validates the EntityConfig. If the EntityConfig has metadata,
// it will be validated via ValidateMetadata. If there is a method for
// EntityConfig called validate, then it will be used to cooperatively
// validate the EntityConfig.
func (e *EntityConfig) Validate() error {
	if e == nil {
		return errors.New("nil EntityConfig")
	}
	var iface interface{} = e
	if resource, ok := iface.(Resource); ok {
		if err := ValidateMetadata(resource.GetMetadata()); err != nil {
			return fmt.Errorf("invalid EntityConfig: %s", err)
		}
	}
	if validator, ok := iface.(validator); ok {
		if err := validator.validate(); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalJSON is provided in order to ensure that metadata labels and
// annotations are never nil.
func (e *EntityConfig) UnmarshalJSON(msg []byte) error {
	type Clone EntityConfig
	var clone Clone
	if err := json.Unmarshal(msg, &clone); err != nil {
		return err
	}
	*e = *(*EntityConfig)(&clone)
	var iface interface{} = e
	var meta *corev2.ObjectMeta
	if metaer, ok := iface.(getMetadataer); ok {
		meta = metaer.GetMetadata()
	}
	if meta != nil {
		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}
		if meta.Annotations == nil {
			meta.Annotations = make(map[string]string)
		}
	}
	return nil
}

// GetTypeMeta gets the type metadata for a EntityConfig.
func (e *EntityConfig) GetTypeMeta() corev2.TypeMeta {
	return corev2.TypeMeta{
		APIVersion: "core/v3",
		Type:       "EntityConfig",
	}
}

// SetMetadata sets the provided metadata on the type. If the type does not
// have any metadata, nothing will happen.
func (e *EntityState) SetMetadata(meta *corev2.ObjectMeta) {
	// The function has to use reflection, since not all of the generated types
	// will have metadata.
	value := reflect.Indirect(reflect.ValueOf(e))
	field := value.FieldByName("Metadata")
	if !field.CanSet() {
		return
	}
	field.Set(reflect.ValueOf(meta))
}

// StoreName returns the store name for EntityState. It will be
// overridden if there is a method for EntityState called "storeName".
func (e *EntityState) StoreName() string {
	var iface interface{} = e
	if prefixer, ok := iface.(storeNamer); ok {
		return prefixer.storeName()
	}
	return "entity_states"
}

// RBACName returns the RBAC name for EntityState. It will be overridden if
// there is a method for EntityState called "rbacName".
func (e *EntityState) RBACName() string {
	var iface interface{} = e
	if namer, ok := iface.(rbacNamer); ok {
		return namer.rbacName()
	}
	return "entity_states"
}

// URIPath returns the URI path for EntityState. It will be overridden if
// there is a method for EntityState called uriPath.
func (e *EntityState) URIPath() string {
	var iface interface{} = e
	if pather, ok := iface.(uriPather); ok {
		return pather.uriPath()
	}
	metaer, ok := iface.(getMetadataer)
	if !ok {
		return ""
	}
	meta := metaer.GetMetadata()
	if meta == nil {
		return uriPath("entity-states", "", "")
	}
	return uriPath("entity-states", meta.Namespace, meta.Name)
}

// Validate validates the EntityState. If the EntityState has metadata,
// it will be validated via ValidateMetadata. If there is a method for
// EntityState called validate, then it will be used to cooperatively
// validate the EntityState.
func (e *EntityState) Validate() error {
	if e == nil {
		return errors.New("nil EntityState")
	}
	var iface interface{} = e
	if resource, ok := iface.(Resource); ok {
		if err := ValidateMetadata(resource.GetMetadata()); err != nil {
			return fmt.Errorf("invalid EntityState: %s", err)
		}
	}
	if validator, ok := iface.(validator); ok {
		if err := validator.validate(); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalJSON is provided in order to ensure that metadata labels and
// annotations are never nil.
func (e *EntityState) UnmarshalJSON(msg []byte) error {
	type Clone EntityState
	var clone Clone
	if err := json.Unmarshal(msg, &clone); err != nil {
		return err
	}
	*e = *(*EntityState)(&clone)
	var iface interface{} = e
	var meta *corev2.ObjectMeta
	if metaer, ok := iface.(getMetadataer); ok {
		meta = metaer.GetMetadata()
	}
	if meta != nil {
		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}
		if meta.Annotations == nil {
			meta.Annotations = make(map[string]string)
		}
	}
	return nil
}

// GetTypeMeta gets the type metadata for a EntityState.
func (e *EntityState) GetTypeMeta() corev2.TypeMeta {
	return corev2.TypeMeta{
		APIVersion: "core/v3",
		Type:       "EntityState",
	}
}
