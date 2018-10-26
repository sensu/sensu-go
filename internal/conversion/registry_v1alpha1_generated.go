package conversion

import (
	"github.com/sensu/sensu-go/internal/apis/meta"
)

func init() {
	registry[key{
		SourceAPIVersion: "meta",
		DestAPIVersion:   "meta/v1alpha1",
		Kind:             "ObjectMeta",
	}] = meta.Convert_meta_ObjectMeta_To_v1alpha1_ObjectMeta

	registry[key{
		SourceAPIVersion: "meta/v1alpha1",
		DestAPIVersion:   "meta",
		Kind:             "ObjectMeta",
	}] = meta.Convert_v1alpha1_ObjectMeta_To_meta_ObjectMeta
}

func init() {
	registry[key{
		SourceAPIVersion: "meta",
		DestAPIVersion:   "meta/v1alpha1",
		Kind:             "TypeMeta",
	}] = meta.Convert_meta_TypeMeta_To_v1alpha1_TypeMeta

	registry[key{
		SourceAPIVersion: "meta/v1alpha1",
		DestAPIVersion:   "meta",
		Kind:             "TypeMeta",
	}] = meta.Convert_v1alpha1_TypeMeta_To_meta_TypeMeta
}
