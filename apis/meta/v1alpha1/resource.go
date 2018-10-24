package v1alpha1

// Automatically generated file, do not edit!

/*
This file contains methods on the types in the v1alpha1 package for
determining resource names.

Resource names are specified with the '+resource-name' special comment, on
types containing meta.TypeMeta. Resource names are specified statically,
and do not change at runtime.
*/

// ResourceName returns the resource name for a ObjectMeta.
// The resource name for ObjectMeta is "".
func (r ObjectMeta) ResourceName() string {
	return ""
}

// ResourceName returns the resource name for a TypeMeta.
// The resource name for TypeMeta is "".
func (r TypeMeta) ResourceName() string {
	return ""
}
