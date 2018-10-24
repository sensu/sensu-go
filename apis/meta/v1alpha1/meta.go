package v1alpha1

import "strings"

type GroupVersionKind interface {
	GetKind() string
	GetGroup() string
	GetVersion() string
	GetTypeMeta() TypeMeta
}

func (tm TypeMeta) GetTypeMeta() TypeMeta { return tm }
func (gvk TypeMeta) GetKind() string      { return gvk.Kind }
func (gvk TypeMeta) GetGroup() string     { return strings.Split(gvk.APIVersion, "/")[0] }
func (gvk TypeMeta) GetVersion() string   { return strings.Split(gvk.APIVersion, "/")[1] }

// Object lets you work with object metadata from any of the versioned or
// internal API objects. Attempting to set or retrieve a field on an object that does
// not support that field (Name, UID, Namespace on lists) will be a no-op and return
// a default value.
type Object interface {
	GetNamespace() string
	SetNamespace(namespace string)
	GetName() string
	SetName(name string)
	GetUUID() string
	SetUUID(uuid string)
	GetResourceVersion() string
	SetResourceVersion(version string)
	GetCreationTimestamp() *Time
	SetCreationTimestamp(timestamp *Time)
	GetDeletionTimestamp() *Time
	SetDeletionTimestamp(timestamp *Time)
	GetLabels() map[string]string
	SetLabels(labels map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	GetClusterName() string
	SetClusterName(clusterName string)
}

// Allow easy access to object metadata for API objects.
func (meta *ObjectMeta) GetNamespace() string                         { return meta.Namespace }
func (meta *ObjectMeta) SetNamespace(namespace string)                { meta.Namespace = namespace }
func (meta *ObjectMeta) GetName() string                              { return meta.Name }
func (meta *ObjectMeta) SetName(name string)                          { meta.Name = name }
func (meta *ObjectMeta) GetUUID() string                              { return meta.UUID }
func (meta *ObjectMeta) SetUUID(uuid string)                          { meta.UUID = uuid }
func (meta *ObjectMeta) GetResourceVersion() string                   { return meta.ResourceVersion }
func (meta *ObjectMeta) SetResourceVersion(version string)            { meta.ResourceVersion = version }
func (meta *ObjectMeta) GetCreationTimestamp() *Time                  { return meta.CreationTimestamp }
func (meta *ObjectMeta) SetCreationTimestamp(time *Time)              { meta.CreationTimestamp = time }
func (meta *ObjectMeta) GetDeletionTimestamp() *Time                  { return meta.DeletionTimestamp }
func (meta *ObjectMeta) SetDeletionTimestamp(time *Time)              { meta.DeletionTimestamp = time }
func (meta *ObjectMeta) GetLabels() map[string]string                 { return meta.Labels }
func (meta *ObjectMeta) SetLabels(labels map[string]string)           { meta.Labels = labels }
func (meta *ObjectMeta) GetAnnotations() map[string]string            { return meta.Annotations }
func (meta *ObjectMeta) SetAnnotations(annotations map[string]string) { meta.Annotations = annotations }
func (meta *ObjectMeta) GetClusterName() string                       { return meta.ClusterName }
func (meta *ObjectMeta) SetClusterName(name string)                   { meta.ClusterName = name }
