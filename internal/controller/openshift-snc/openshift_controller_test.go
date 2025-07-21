// File: openshift_controller_test.go

package openshiftsnc

import (
	"context"

	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	maptv1alpha1 "github.com/mapt-oss/mapt-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Openshift Controller", func() {
	const (
		resourceName      = "openshift-sample"
		resourceNamespace = "default"
	)

	ctx := context.Background()
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}
	openshift := &maptv1alpha1.Openshift{}

	BeforeEach(func() {
		By("Creating the Openshift custom resource if not exists")
		err := k8sClient.Get(ctx, namespacedName, openshift)
		if errors.IsNotFound(err) {
			openshift := &maptv1alpha1.Openshift{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: resourceNamespace,
				},
				Spec: maptv1alpha1.OpenshiftSpec{
					// Populate fields if needed
				},
			}
			Expect(k8sClient.Create(ctx, openshift)).To(Succeed())
		}
	})

	AfterEach(func() {
		By("Cleaning up the Openshift custom resource")
		res := &maptv1alpha1.Openshift{}
		Expect(k8sClient.Get(ctx, namespacedName, res)).To(Succeed())
		Expect(k8sClient.Delete(ctx, res)).To(Succeed())
	})

	It("should reconcile successfully", func() {
		By("Reconciling Openshift resource")
		reconciler := &OpenshiftReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := reconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: namespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Validating updated status phase")
		Eventually(func() v1alpha1.OpenshiftSncPhase {
			updated := &maptv1alpha1.Openshift{}
			Expect(k8sClient.Get(ctx, namespacedName, updated)).To(Succeed())
			return updated.Status.Phase
		}).Should(Not(BeEmpty()))
	})
})
