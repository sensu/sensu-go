package v1alpha1

import (
	"unsafe"

	"github.com/sensu/sensu-go/internal/apis/meta"
)

// ConvertTo converts a *ObjectMeta to a *meta.ObjectMeta.
// It panics if the to parameter is not a *meta.ObjectMeta.
func (r *ObjectMeta) ConvertTo(to interface{}) {
	ptr := to.(*meta.ObjectMeta)
	convert_ObjectMeta_To_meta_ObjectMeta(r, ptr)
}

var convert_ObjectMeta_To_meta_ObjectMeta = func(from *ObjectMeta, to *meta.ObjectMeta) {
	 *to = *(*meta.ObjectMeta)(unsafe.Pointer(from))
}

// ConvertFrom converts the receiver to a *meta.ObjectMeta.
// It panics if the from parameter is not a *meta.ObjectMeta.
func (r *ObjectMeta) ConvertFrom(from interface{}) {
	ptr := from.(*meta.ObjectMeta)
	convert_meta_ObjectMeta_To_ObjectMeta(ptr, r)
}

var convert_meta_ObjectMeta_To_ObjectMeta = func(from *meta.ObjectMeta, to *ObjectMeta) {
	 *to = *(*ObjectMeta)(unsafe.Pointer(from))
}

// ConvertTo converts a *TypeMeta to a *meta.TypeMeta.
// It panics if the to parameter is not a *meta.TypeMeta.
func (r *TypeMeta) ConvertTo(to interface{}) {
	ptr := to.(*meta.TypeMeta)
	convert_TypeMeta_To_meta_TypeMeta(r, ptr)
}

var convert_TypeMeta_To_meta_TypeMeta = func(from *TypeMeta, to *meta.TypeMeta) {
	 *to = *(*meta.TypeMeta)(unsafe.Pointer(from))
}

// ConvertFrom converts the receiver to a *meta.TypeMeta.
// It panics if the from parameter is not a *meta.TypeMeta.
func (r *TypeMeta) ConvertFrom(from interface{}) {
	ptr := from.(*meta.TypeMeta)
	convert_meta_TypeMeta_To_TypeMeta(ptr, r)
}

var convert_meta_TypeMeta_To_TypeMeta = func(from *meta.TypeMeta, to *TypeMeta) {
	 *to = *(*TypeMeta)(unsafe.Pointer(from))
}

