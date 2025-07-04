package controller

import (
	"context"
	"os"

	v1alpha1 "github.com/flacatus/mapt-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func createKubeconfigSecret(path, secretName string, owner *v1alpha1.Kind, c client.Client) error {
	kubeconfigBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: secretName,
			Namespace:    owner.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         owner.APIVersion,
					Kind:               owner.Kind,
					Name:               owner.Name,
					UID:                owner.UID,
					BlockOwnerDeletion: pointer.Bool(false),
				},
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{"kubeconfig": kubeconfigBytes},
	}

	if err := controllerutil.SetControllerReference(owner, secret, c.Scheme()); err != nil {
		return err
	}

	err = c.Create(context.TODO(), secret)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
