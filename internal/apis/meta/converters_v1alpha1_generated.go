package meta

import (
	"unsafe"

	"github.com/sensu/sensu-go/apis/meta/v1alpha1"
)

func Convert_meta_ObjectMeta_To_v1alpha1_ObjectMeta(dst, src interface{}) error {
	dstp := dst.(*v1alpha1.ObjectMeta)
	srcp := src.(*ObjectMeta)

	*dstp = *(*v1alpha1.ObjectMeta)(unsafe.Pointer(srcp))

	return nil
}

func Convert_v1alpha1_ObjectMeta_To_meta_ObjectMeta(dst, src interface{}) error {
	dstp := dst.(*ObjectMeta)
	srcp := src.(*v1alpha1.ObjectMeta)

	*dstp = *(*ObjectMeta)(unsafe.Pointer(srcp))

	return nil
}

func Convert_meta_TypeMeta_To_v1alpha1_TypeMeta(dst, src interface{}) error {
	dstp := dst.(*v1alpha1.TypeMeta)
	srcp := src.(*TypeMeta)

	*dstp = *(*v1alpha1.TypeMeta)(unsafe.Pointer(srcp))

	return nil
}

func Convert_v1alpha1_TypeMeta_To_meta_TypeMeta(dst, src interface{}) error {
	dstp := dst.(*TypeMeta)
	srcp := src.(*v1alpha1.TypeMeta)

	*dstp = *(*TypeMeta)(unsafe.Pointer(srcp))

	return nil
}
