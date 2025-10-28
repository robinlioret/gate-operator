package internal

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetReferencedObject(
	ctx context.Context,
	r client.Reader,
	ref *corev1.ObjectReference,
	namespace string,
) (client.Object, error) {
	if ref == nil {
		return nil, errors.NewBadRequest("ObjectReference is nil")
	}

	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return nil, err
	}

	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    ref.Kind,
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

	ns := ref.Namespace
	if ns == "" {
		ns = namespace
	}

	key := client.ObjectKey{Name: ref.Name, Namespace: ns}
	if err := r.Get(ctx, key, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
