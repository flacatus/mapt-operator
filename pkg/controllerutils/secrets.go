package controllerutils

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// updated signature to accept scheme and owner
func CreateGeneratedSecret(ctx context.Context, c client.Client, scheme *runtime.Scheme, data map[string][]byte, owner client.Object) (string, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: owner.GetName() + "-",
			Namespace:    owner.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         owner.GetObjectKind().GroupVersionKind().GroupVersion().String(),
					Kind:               owner.GetObjectKind().GroupVersionKind().Kind,
					Name:               owner.GetName(),
					UID:                owner.GetUID(),
					BlockOwnerDeletion: pointer.Bool(false),
				},
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	if err := c.Create(ctx, secret); err != nil {
		if !errors.IsAlreadyExists(err) {
			return "", fmt.Errorf("failed to create secret: %w", err)
		}
	}

	return secret.Name, nil
}
